// Package moqt implements the MOQ Transfork protocol, providing a multiplexer for track routing.
// It allows handling of track subscriptions, announcements, and track serving.
package moqt

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// DefaultMux is the default trackMux used by the top-level functions.
// It can be used directly instead of creating a new trackMux.
var DefaultMux *TrackMux = defaultMux

var defaultMux = NewTrackMux()

// NewTrackMux creates a new trackMux for handling track and announcement routing.
// It initializes the routing and announcement trees with empty root nodes.
func NewTrackMux() *TrackMux {
	return &TrackMux{
		announcementTree:  *newAnnouncingNode(),
		handlerIndex:      make(map[BroadcastPath]TrackHandler),
		announcementIndex: make(map[BroadcastPath]*Announcement),
	}
}

// Publish registers the handler for the given track path in the DefaultMux.
// The handler will remain active until the context is canceled.
func Publish(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	DefaultMux.Publish(ctx, path, handler)
}

func PublishFunc(ctx context.Context, path BroadcastPath, f func(tw *TrackWriter)) {
	DefaultMux.PublishFunc(ctx, path, f)
}

func Announce(announcement *Announcement, handler TrackHandler) {
	DefaultMux.Announce(announcement, handler)
}

// TrackMux is a multiplexer for routing track requests and announcements.
// It maintains separate trees for track routing and announcements, allowing efficient
// lookup of handlers and distribution of announcements to interested subscribers.
type TrackMux struct {
	handlerMu    sync.RWMutex
	handlerIndex map[BroadcastPath]TrackHandler

	announcementTree  announcingNode
	treeMu            sync.RWMutex
	announcementIndex map[BroadcastPath]*Announcement
}

func (mux *TrackMux) PublishFunc(ctx context.Context, path BroadcastPath, f func(tw *TrackWriter)) {
	mux.Publish(ctx, path, TrackHandlerFunc(f))
}

// Handle registers the handler for the given track path.
// The handler will remain active until the context is canceled.
func (mux *TrackMux) Publish(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	if ctx == nil {
		panic("[TrackMux] nil context")
	}

	if !isValidPath(path) {
		panic("[TrackMux] invalid track path: " + path)
	}
	ann, _ := NewAnnouncement(ctx, path)
	mux.Announce(ann, handler)
}

func (mux *TrackMux) addHandler(path BroadcastPath, handler TrackHandler) {
	mux.handlerMu.Lock()
	defer mux.handlerMu.Unlock()

	mux.handlerIndex[path] = handler
}

func (mux *TrackMux) removeHandler(path BroadcastPath) {
	mux.handlerMu.Lock()
	defer mux.handlerMu.Unlock()
	delete(mux.handlerIndex, path)
}

func (mux *TrackMux) addAnnouncement(announcement *Announcement) (lastNode *announcingNode) {
	path := announcement.BroadcastPath()

	var old *Announcement
	mux.treeMu.Lock()
	old, ok := mux.announcementIndex[path]
	if ok {
		if old == announcement {
			mux.treeMu.Unlock()
			return // Already registered
		}
	}
	mux.announcementIndex[path] = announcement
	mux.treeMu.Unlock()

	if old != nil {
		old.end()
	}

	segments := strings.Split(string(path), "/")
	return mux.announcementTree.addAnnouncement(segments, announcement)
}

func (mux *TrackMux) removeAnnouncement(path BroadcastPath) {
	mux.treeMu.Lock()
	old, ok := mux.announcementIndex[path]
	if ok {
		delete(mux.announcementIndex, path)
	}
	mux.treeMu.Unlock()

	if ok {
		// Call end outside the lock to avoid deadlocks
		old.end()
		return
	}

}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	path := announcement.BroadcastPath()

	if !announcement.IsActive() {
		slog.Warn("[TrackMux] announcement is not active",
			"path", path,
		)
		return
	}

	// Add the handler to the mux if it is not already registered
	mux.addHandler(path, handler)

	// add Announcement to the mux
	lastNode := mux.addAnnouncement(announcement)

	// Send announcement to all registered channels with retry mechanism
	lastNode.mu.RLock()
	for ch := range lastNode.channels {
		select {
		case ch <- announcement:
			// Successfully sent to channel
		default:
			// Channel is busy, start retry goroutine
			go func(channel chan *Announcement) {
				for {
					select {
					case channel <- announcement:
						// Successfully sent to channel
						return
					case <-time.After(100 * time.Millisecond):
						// Timeout, retry
						continue
					case <-announcement.AwaitEnd():
						// Announcement ended, no need to send
						return
					}
				}
			}(ch)
		}
	}
	lastNode.mu.RUnlock()

	announcement.OnEnd(func() {
		// Remove the announcement from the tree unconditionally
		lastNode.removeAnnouncement(announcement)

		// Remove from the mux (index and handler) only if this announcement is still
		// the current one registered for the path. This prevents a replaced
		// announcement's OnEnd from deleting the newer handler/announcement.
		mux.treeMu.Lock()
		cur, ok := mux.announcementIndex[path]
		if ok && cur == announcement {
			delete(mux.announcementIndex, path)
			mux.treeMu.Unlock()
			mux.removeHandler(path)
			return
		}
		mux.treeMu.Unlock()
	})
}

// Handler returns the handler for the specified track path.
// If no handler is found, NotFoundTrackHandler is returned.
func (mux *TrackMux) Publishr(path BroadcastPath) TrackHandler {
	if !isValidPath(path) {
		panic("mux: invalid track path: " + path)
	}

	mux.handlerMu.RLock()
	handler, ok := mux.handlerIndex[path]
	mux.handlerMu.RUnlock()
	// Treat typed-nil handler functions as absent too
	if hf, okHF := handler.(TrackHandlerFunc); okHF && hf == nil {
		handler = nil
		ok = false
	}
	if handler == nil || !ok {
		slog.Warn("mux: no handler found for path", "path", path)
		return NotFoundTrackHandler
	}
	return handler
}

// ServeTrack serves the track at the specified path using the appropriate handler.
// It finds the handler for the path and delegates the serving to it.
func (mux *TrackMux) ServeTrack(tw *TrackWriter) {
	if tw == nil {
		slog.Error("mux: nil track writer")
		return
	}
	handler := mux.Publishr(tw.BroadcastPath)

	handler.ServeTrack(tw)
}

// ServeAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) ServeAnnouncements(aw *AnnouncementWriter, prefix string) {
	if aw == nil {
		slog.Error("mux: nil announcement writer")
		return
	}

	if !isValidPrefix(prefix) {
		aw.CloseWithError(InvalidPrefixErrorCode)
		return
	}

	segments := strings.Split(prefix, "/")

	// Register the handler on the routing tree (protect children with per-node lock)
	current := &mux.announcementTree
	for _, seg := range segments[1 : len(segments)-1] {
		current.mu.Lock()
		if current.children == nil {
			current.children = make(map[string]*announcingNode)
		}
		child, ok := current.children[seg]
		if !ok {
			child = newAnnouncingNode()
			current.children[seg] = child
		}
		next := child
		current.mu.Unlock()
		current = next
	}

	ch := make(chan *Announcement, 1)
	current.mu.Lock()
	current.channels[ch] = struct{}{}
	current.mu.Unlock()

	// Cleanup channel when done
	defer func() {
		current.mu.Lock()
		delete(current.channels, ch)
		current.mu.Unlock()
		close(ch)
	}()

	// Get existing announcements in a separate goroutine to avoid blocking new announcements
	current.mu.RLock()
	err := aw.init(current.announcements)
	current.mu.RUnlock()
	if err != nil {
		slog.Error("[TrackMux] failed to initialize announcement writer", "error", err)
		aw.CloseWithError(InternalAnnounceErrorCode)
		return
	}

	// Process announcements from channel
	for {
		select {
		case ann, ok := <-ch:
			if !ok {
				return // Channel closed
			}
			err := aw.SendAnnouncement(ann)
			if err != nil {
				slog.Error("[TrackMux] failed to send announcement", "error", err)
				aw.CloseWithError(InternalAnnounceErrorCode)
				return
			}
		case <-aw.Context().Done():
			// Writer context cancelled
			return
		}
	}
}

func (mux *TrackMux) Clear() {
	mux.treeMu.Lock()
	mux.announcementTree = *newAnnouncingNode()
	mux.announcementIndex = make(map[BroadcastPath]*Announcement)
	mux.treeMu.Unlock()

	mux.handlerMu.Lock()
	mux.handlerIndex = make(map[BroadcastPath]TrackHandler)
	mux.handlerMu.Unlock()
}

// newAnnouncingNode creates and initializes a new routing tree node.
func newAnnouncingNode() *announcingNode {
	return &announcingNode{
		announcements: make(map[*Announcement]struct{}),
		children:      make(map[string]*announcingNode),
		channels:      make(map[chan *Announcement]struct{}),
	}
}

type prefixSegment = string

type announcingNode struct {
	mu sync.RWMutex

	channels map[chan *Announcement]struct{}

	parent *announcingNode

	children map[prefixSegment]*announcingNode

	announcements map[*Announcement]struct{}
}

func (node *announcingNode) createChild(seg prefixSegment) *announcingNode {
	node.mu.Lock()
	defer node.mu.Unlock()
	child, ok := node.children[seg]
	if !ok {
		child = newAnnouncingNode()
		child.parent = node
		node.children[seg] = child
	}
	return child
}

func (node *announcingNode) addAnnouncement(segments []string, announcement *Announcement) (lastNode *announcingNode) {
	node.mu.Lock()
	node.announcements[announcement] = struct{}{}
	node.mu.Unlock()
	if len(segments) == 0 {
		return node
	}

	child := node.createChild(segments[0])

	return child.addAnnouncement(segments[1:], announcement)
}

func (node *announcingNode) removeAnnouncement(announcement *Announcement) {
	node.mu.Lock()
	delete(node.announcements, announcement)
	node.mu.Unlock()

	if node.parent != nil {
		node.parent.removeAnnouncement(announcement)
	}
}

func isValidPath(path BroadcastPath) bool {
	if path == "" {
		return false
	}

	p := string(path)
	if !strings.HasPrefix(p, "/") {
		return false
	}

	// Disallow parent directory segments to avoid path traversal like "/../x"
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return false
		}
	}

	return true
}

func isValidPrefix(prefix string) bool {
	if prefix == "" {
		return false
	}

	if !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") {
		return false
	}

	return true
}

type TrackHandler interface {
	ServeTrack(*TrackWriter)
}

var NotFound = func(tw *TrackWriter) {
	if tw == nil {
		return
	}

	tw.CloseWithError(TrackNotFoundErrorCode)
}

var NotFoundTrackHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(*TrackWriter)

func (f TrackHandlerFunc) ServeTrack(tw *TrackWriter) {
	f(tw)
}

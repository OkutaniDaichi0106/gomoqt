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
		trackHandlerIndex: make(map[BroadcastPath]*announcedTrackHandler),
	}
}

// Publish registers the handler for the given track path in the DefaultMux.
// The handler will remain active until the context is canceled.
func Publish(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	DefaultMux.Publish(ctx, path, handler)
}

func PublishFunc(ctx context.Context, path BroadcastPath, f func(ctx context.Context, tw *TrackWriter)) {
	DefaultMux.PublishFunc(ctx, path, f)
}

func Announce(announcement *Announcement, handler TrackHandler) {
	DefaultMux.Announce(announcement, handler)
}

// TrackMux is a multiplexer for routing track requests and announcements.
// It maintains separate trees for track routing and announcements, allowing efficient
// lookup of handlers and distribution of announcements to interested subscribers.
type TrackMux struct {
	handlerMu         sync.RWMutex
	trackHandlerIndex map[BroadcastPath]*announcedTrackHandler

	announcementTree announcingNode
	treeMu           sync.RWMutex
}

func (mux *TrackMux) PublishFunc(ctx context.Context, path BroadcastPath, f func(ctx context.Context, tw *TrackWriter)) {
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

func (mux *TrackMux) registerHandler(path BroadcastPath, handler *announcedTrackHandler) {
	mux.handlerMu.Lock()
	old, ok := mux.trackHandlerIndex[path]
	if !ok {
		mux.trackHandlerIndex[path] = handler
		mux.handlerMu.Unlock()
		return
	}
	mux.trackHandlerIndex[path] = handler
	mux.handlerMu.Unlock()

	old.end()
}

func (mux *TrackMux) removeHandler(handler *announcedTrackHandler) {
	path := handler.BroadcastPath()
	mux.handlerMu.Lock()
	defer mux.handlerMu.Unlock()
	if ath, ok := mux.trackHandlerIndex[path]; ok && ath == handler {
		delete(mux.trackHandlerIndex, path)
	}
}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	if announcement == nil {
		slog.Warn("[TrackMux] Announce called with nil Announcement")
		return
	}

	path := announcement.BroadcastPath()

	if !announcement.IsActive() {
		slog.Warn("[TrackMux] announcement is not active",
			"path", path,
		)
		return
	}

	announcedHandler := &announcedTrackHandler{
		Announcement: announcement,
		TrackHandler: handler,
	}
	// Add the handler to the mux if it is not already registered
	mux.registerHandler(path, announcedHandler)

	// add Announcement to the mux
	segments := strings.Split(string(path), "/")
	lastNode := mux.announcementTree.addAnnouncement(segments, announcement)

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
					case <-announcement.Context().Done():
						// Announcement ended, no need to send
						return
					}
				}
			}(ch)
		}
	}
	lastNode.mu.RUnlock()

	// Ensure the announcement is removed when it ends
	announcement.OnEnd(func() {
		// Remove the announcement from the tree unconditionally
		lastNode.removeAnnouncement(announcement)

		mux.removeHandler(announcedHandler)
	})
}

// Handler returns the handler for the specified track path.
// If no handler is found, NotFoundTrackHandler is returned.
func (mux *TrackMux) TrackHandler(path BroadcastPath) (*Announcement, TrackHandler) {
	ath := mux.findTrackHandler(path)
	if ath == nil {
		return nil, NotFoundTrackHandler
	}
	return ath.Announcement, ath.TrackHandler
}

func (mux *TrackMux) findTrackHandler(path BroadcastPath) *announcedTrackHandler {
	if !isValidPath(path) {
		return nil
	}

	mux.handlerMu.RLock()
	ath, ok := mux.trackHandlerIndex[path]
	mux.handlerMu.RUnlock()
	if !ok {
		return nil
	}

	if ath == nil || ath.Announcement == nil || ath.TrackHandler == nil {
		return nil
	}

	// Treat typed-nil handler functions as absent too
	if hf, ok := ath.TrackHandler.(TrackHandlerFunc); ok && hf == nil {
		slog.Warn("mux: handler function is nil for path", "path", path)
		return nil
	}

	return ath
}

// serveTrack serves the track at the specified path using the appropriate handler.
// It finds the handler for the path and delegates the serving to it.
func (mux *TrackMux) serveTrack(tw *TrackWriter) {
	if tw == nil {
		slog.Error("mux: nil track writer")
		return
	}

	path := tw.BroadcastPath

	mux.handlerMu.RLock()
	ath, ok := mux.trackHandlerIndex[path]
	mux.handlerMu.RUnlock()
	if !ok || ath == nil || ath.Announcement == nil {
		slog.Warn("[TrackMux] no announcement found for path", "path", path)
		tw.CloseWithError(TrackNotFoundErrorCode)
		return
	}

	if ath.TrackHandler == nil {
		slog.Warn("mux: no handler found for path", "path", path)
		tw.CloseWithError(TrackNotFoundErrorCode)
		return
	}

	ath.TrackHandler.ServeTrack(ath.Announcement.ctx, tw)
}

// serveAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) serveAnnouncements(aw *AnnouncementWriter, prefix string) {
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
	mux.treeMu.Unlock()

	mux.handlerMu.Lock()
	mux.trackHandlerIndex = make(map[BroadcastPath]*announcedTrackHandler)
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
	ServeTrack(context.Context, *TrackWriter)
}

var NotFound = func(ctx context.Context, tw *TrackWriter) {
	if tw == nil {
		return
	}

	tw.CloseWithError(TrackNotFoundErrorCode)
}

var NotFoundTrackHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(context.Context, *TrackWriter)

func (f TrackHandlerFunc) ServeTrack(ctx context.Context, tw *TrackWriter) {
	f(ctx, tw)
}

var _ TrackHandler = (*announcedTrackHandler)(nil)

type announcedTrackHandler struct {
	TrackHandler
	*Announcement
}

func (h *announcedTrackHandler) Context() context.Context {
	return h.ctx
}

// Package moqt implements the MOQ Transfork protocol, providing a multiplexer for track routing.
// It allows handling of track subscriptions, announcements, and track serving.
package moqt

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
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
		announcementTree: *newAnnouncingNode(),
		handlerIndex:     make(map[BroadcastPath]TrackHandler),
	}
}

// Handle registers the handler for the given track path in the DefaultMux.
// The handler will remain active until the context is canceled.
func Handle(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	DefaultMux.Handle(ctx, path, handler)
}

func HandleFunc(ctx context.Context, path BroadcastPath, f func(tw *TrackWriter)) {
	DefaultMux.HandleFunc(ctx, path, f)
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

	treeMu           sync.RWMutex
	announcementTree announcingNode
}

func (mux *TrackMux) HandleFunc(ctx context.Context, path BroadcastPath, f func(tw *TrackWriter)) {
	mux.Handle(ctx, path, TrackHandlerFunc(f))
}

// Handle registers the handler for the given track path.
// The handler will remain active until the context is canceled.
func (mux *TrackMux) Handle(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	if ctx == nil {
		panic("[TrackMux] nil context")
	}

	if !isValidPath(path) {
		panic("[TrackMux] invalid track path: " + path)
	}

	mux.Announce(NewAnnouncement(ctx, path), handler)
}

func (mux *TrackMux) addHandler(path BroadcastPath, handler TrackHandler) bool {
	mux.handlerMu.Lock()
	defer mux.handlerMu.Unlock()
	if _, ok := mux.handlerIndex[path]; ok {
		return false
	}
	mux.handlerIndex[path] = handler
	return true
}

func (mux *TrackMux) removeHandler(path BroadcastPath) {
	mux.handlerMu.Lock()
	defer mux.handlerMu.Unlock()
	delete(mux.handlerIndex, path)
}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	path := announcement.BroadcastPath()

	if !isValidPath(path) {
		panic("[TrackMux] invalid track path: " + path)
	}

	if !announcement.IsActive() {
		slog.Warn("[TrackMux] announcement is not active",
			"path", path,
		)
		return
	}

	// Add the handler to the mux if it is not already registered
	if !mux.addHandler(path, handler) {
		slog.Warn("[TrackMux] handler already registered for path", "path", path)
		return
	}

	segments := strings.Split(string(path), "/")
	node := mux.announcementTree.findNode(segments[1:len(segments)-1], countAnnouncement)

	lastPart := segments[len(segments)-1]
	node.addAnnouncement(lastPart, announcement)

	// Send announcement to all registered channels with retry mechanism
	node.mu.RLock()
	for ch := range node.channels {
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
	node.mu.RUnlock()

	// // Delete failed writers from the current node
	// if len(failedWriters) > 0 {
	// 	node.mu.Lock()
	// 	for _, writer := range failedWriters {
	// 		delete(node.writers, writer)
	// 	}
	// 	node.mu.Unlock()
	// }

	announcement.OnEnd(func() {
		// Remove the handler
		mux.removeHandler(path)

		// Remove the announcement
		node.removeAnnouncement(lastPart, announcement)
	})
}

// Handler returns the handler for the specified track path.
// If no handler is found, NotFoundTrackHandler is returned.
func (mux *TrackMux) Handler(path BroadcastPath) TrackHandler {
	if !isValidPath(path) {
		panic("mux: invalid track path: " + path)
	}

	mux.handlerMu.RLock()
	handler, ok := mux.handlerIndex[path]
	mux.handlerMu.RUnlock()
	if handler == nil || !ok {
		slog.Warn("mux: no handler found for path", "path", path)
		return NotFoundHandler
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

	handler := mux.Handler(tw.BroadcastPath)

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

	// Use unbuffered channel with retry mechanism for memory efficiency
	ch := make(chan *Announcement)
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
	var initErr error
	initDone := make(chan struct{})

	go func() {
		defer close(initDone)
		announcements := current.appendAnnouncements(nil)
		initErr = aw.init(announcements)
	}()

	// Wait for initialization to complete or context cancellation
	select {
	case <-initDone:
		if initErr != nil {
			slog.Error("[TrackMux] failed to initialize announcement writer", "error", initErr)
			aw.CloseWithError(InternalAnnounceErrorCode)
			return
		}
	case <-aw.Context().Done():
		// Writer context cancelled during initialization
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
	mux.handlerIndex = make(map[BroadcastPath]TrackHandler)
	mux.handlerMu.Unlock()
}

// newAnnouncingNode creates and initializes a new routing tree node.
func newAnnouncingNode() *announcingNode {
	return &announcingNode{
		announcements: make(map[segment]*Announcement),
		children:      make(map[string]*announcingNode),
		channels:      make(map[chan *Announcement]struct{}),
	}
}

type segment = string

type announcingNode struct {
	mu sync.RWMutex

	announcements map[segment]*Announcement

	channels map[chan *Announcement]struct{}

	announcementsCount atomic.Uint64
	children           map[string]*announcingNode
}

func (node *announcingNode) findNode(segments []string, op func(*announcingNode)) *announcingNode {
	current := node

	for _, seg := range segments {
		if op != nil {
			op(current)
		}

		current.mu.Lock()
		if current.children == nil {
			current.children = make(map[string]*announcingNode)
		}
		child, ok := current.children[seg]
		if !ok {
			child = newAnnouncingNode()
			current.children[seg] = child
		}
		current.mu.Unlock()

		current = child
	}

	return current
}

func (node *announcingNode) addAnnouncement(segment segment, announcement *Announcement) {
	node.mu.Lock()
	defer node.mu.Unlock()

	node.announcements[segment] = announcement
}

func (node *announcingNode) removeAnnouncement(segment segment, announcement *Announcement) {
	node.mu.Lock()
	defer node.mu.Unlock()

	if node.announcements[segment] == announcement {
		delete(node.announcements, segment)
	}
}

func (node *announcingNode) appendAnnouncements(anns []*Announcement) []*Announcement {
	if anns == nil {
		anns = make([]*Announcement, 0, node.announcementsCount.Load())
	}

	node.mu.RLock()
	// Take a snapshot of announcements and children
	for _, a := range node.announcements {
		anns = append(anns, a)
	}
	children := make([]*announcingNode, 0, len(node.children))
	for _, c := range node.children {
		children = append(children, c)
	}
	node.mu.RUnlock()

	// Recursively announce to child nodes
	for _, child := range children {
		anns = child.appendAnnouncements(anns)
	}

	return anns
}

func countAnnouncement(node *announcingNode) {
	if node == nil {
		return
	}

	node.announcementsCount.Add(1)
}

func isValidPath(path BroadcastPath) bool {
	if path == "" {
		return false
	}

	return strings.HasPrefix(string(path), "/")
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

// Package moqt implements the MOQ Transfork protocol, providing a multiplexer for track routing.
// It allows handling of track subscriptions, announcements, and track serving.
package moqt

import (
	"context"
	"log/slog"
	"strings"
	"sync"
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

	mux.Announce(NewAnnouncement(ctx, path), handler)
}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	path := announcement.BroadcastPath()

	if !isValidPath(path) {
		panic("[TrackMux] invalid track path: " + path)
	}

	if !announcement.IsActive() {
		slog.Warn("[TrackMux] announcement is not active")
		return
	}

	// Protect handlerIndex registration and duplicate check with handlerMu
	mux.handlerMu.Lock()
	if _, ok := mux.handlerIndex[path]; ok {
		mux.handlerMu.Unlock()
		slog.Warn("[TrackMux] handler already registered for path", "path", path)
		return
	}
	mux.handlerIndex[path] = handler
	mux.handlerMu.Unlock()

	// Protect announcementTree structure modification with per-node lock
	segments := strings.Split(string(path), "/")
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
		current = child
		current.mu.Unlock()
	}

	current.mu.Lock()
	current.announcements[path] = announcement
	current.mu.Unlock()

	// Collect failed writers and clean them up with proper locking
	var failedWriters []*AnnouncementWriter

	current.mu.RLock()
	for writer := range current.writers {
		err := writer.SendAnnouncement(announcement)
		if err != nil {
			writer.CloseWithError(InternalAnnounceErrorCode)
			failedWriters = append(failedWriters, writer)
			slog.Error("mux: failed to send announcement", "error", err)
			continue
		}
	}
	current.mu.RUnlock()

	// Delete failed writers from the current node
	if len(failedWriters) > 0 {
		current.mu.Lock()
		for _, writer := range failedWriters {
			delete(current.writers, writer)
		}
		current.mu.Unlock()
	}

	announcement.OnEnd(func() {
		// Remove the handler
		mux.handlerMu.Lock()
		delete(mux.handlerIndex, path)
		mux.handlerMu.Unlock()

		// Remove the announcement
		current.mu.Lock()
		delete(current.announcements, path)
		current.mu.Unlock()
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

func (mux *TrackMux) ServeAnnouncements(w *AnnouncementWriter, prefix string) {
	if w == nil {
		slog.Error("mux: nil announcement writer")
		return
	}

	if !isValidPrefix(prefix) {
		w.CloseWithError(InvalidPrefixErrorCode)
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

	// Find existing announcements and initialize the writer
	announcements := make([]*Announcement, 0, len(current.announcements))
	announcements = current.appendAnnouncements(announcements)
	err := w.init(announcements)
	if err != nil {
		slog.Error("mux: failed to initialize announcement writer", "error", err)
		w.CloseWithError(InternalAnnounceErrorCode)
		return
	}

	// Register the writer in the current node
	current.writers[w] = struct{}{}
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
		announcements: make(map[BroadcastPath]*Announcement),
		writers:       make(map[*AnnouncementWriter]struct{}),
		children:      make(map[string]*announcingNode),
	}
}

type announcingNode struct {
	mu sync.RWMutex

	announcements map[BroadcastPath]*Announcement
	writers       map[*AnnouncementWriter]struct{}

	children map[string]*announcingNode
}

func (node *announcingNode) appendAnnouncements(anns []*Announcement) []*Announcement {
	node.mu.RLock()
	// Take a snapshot of announcements and children
	for _, a := range node.announcements {
		slog.Info("found announcement", "suffix", a.BroadcastPath())
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

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

func HandleFunc(ctx context.Context, path BroadcastPath, f func(pub *Publisher)) {
	DefaultMux.HandleFunc(ctx, path, f)
}

func Announce(announcement *Announcement, handler TrackHandler) {
	DefaultMux.Announce(announcement, handler)
}

// TrackMux is a multiplexer for routing track requests and announcements.
// It maintains separate trees for track routing and announcements, allowing efficient
// lookup of handlers and distribution of announcements to interested subscribers.
type TrackMux struct {
	mu sync.RWMutex

	announcementTree announcingNode

	handlerIndex map[BroadcastPath]TrackHandler
}

func (mux *TrackMux) HandleFunc(ctx context.Context, path BroadcastPath, f func(pub *Publisher)) {
	mux.Handle(ctx, path, TrackHandlerFunc(f))
}

// Handle registers the handler for the given track path.
// The handler will remain active until the context is canceled.
func (mux *TrackMux) Handle(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	if ctx == nil {
		slog.Error("mux: nil context")
		return
	}

	mux.Announce(NewAnnouncement(ctx, path), handler)
}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	path := announcement.BroadcastPath()

	if !isValidPath(path) {
		panic("mux: invalid track path: " + path)
	}

	if !announcement.IsActive() {
		slog.Warn("mux: announcement is not active")
		return
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	_, ok := mux.handlerIndex[path]
	if ok {
		slog.Warn("mux: handler already registered for path", "path", path)
		return
	}

	mux.handlerIndex[path] = handler

	segments := strings.Split(string(path), "/")

	current := &mux.announcementTree
	for _, seg := range segments[1 : len(segments)-1] {
		if current.children == nil {
			current.children = make(map[string]*announcingNode)
		}
		if child, ok := current.children[seg]; ok {
			current = child
		} else {
			child := newAnnouncingNode()
			current.children[seg] = child
			current = child
		}
	}
	current.mu.Lock()
	current.announcements[path] = announcement
	current.mu.Unlock()

	// Collect failed writers and clean them up with proper locking
	var failedWriters []AnnouncementWriter

	current.mu.RLock()
	for writer := range current.writers {
		err := writer.SendAnnouncement(announcement)
		if err != nil {
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

	go func() {
		<-announcement.AwaitEnd()

		// Remove the handler
		mux.mu.Lock()
		delete(mux.handlerIndex, path)
		mux.mu.Unlock()

		// Remove the announcement
		current.mu.Lock()
		delete(current.announcements, path)
		current.mu.Unlock()

		slog.Debug("removed track handler",
			"track_path", path,
		)
	}()

	slog.Debug("registered track handler",
		"track_path", path,
	)
}

// Handler returns the handler for the specified track path.
// If no handler is found, NotFoundTrackHandler is returned.
func (mux *TrackMux) Handler(path BroadcastPath) TrackHandler {
	if !isValidPath(path) {
		panic("mux: invalid track path: " + path)
	}

	mux.mu.RLock()
	defer mux.mu.RUnlock()

	handler, ok := mux.handlerIndex[path]
	if handler == nil || !ok {
		slog.Warn("mux: no handler found for path", "path", path)
		return NotFoundHandler
	}

	return handler
}

// ServeTrack serves the track at the specified path using the appropriate handler.
// It finds the handler for the path and delegates the serving to it.
func (mux *TrackMux) ServeTrack(pub *Publisher) {
	if pub == nil {
		slog.Error("mux: nil publisher")
		return
	}
	if pub.TrackWriter == nil {
		slog.Error("mux: nil track writer")
		return
	}
	if pub.SubscribeStream == nil {
		slog.Error("mux: nil subscribe stream")
		return
	}

	handler := mux.Handler(pub.BroadcastPath)

	handler.ServeTrack(pub)
}

// ServeAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) ServeAnnouncements(w AnnouncementWriter, prefix string) {
	if w == nil {
		slog.Error("mux: nil announcement writer")
		return
	}

	if !isValidPrefix(prefix) {
		w.CloseWithError(InvalidPrefixErrorCode)
		return
	}

	segments := strings.Split(prefix, "/")

	slog.Debug("mux: serving announcements for prefix", "prefix", prefix)

	// Register the handler on the routing tree
	mux.mu.Lock()
	current := &mux.announcementTree
	for _, seg := range segments[1 : len(segments)-1] {
		if current.children == nil {
			current.children = make(map[string]*announcingNode)
		}

		if child, ok := current.children[seg]; ok {
			current = child
		} else {
			child := newAnnouncingNode()
			current.children[seg] = child
			current = child
		}
	}
	mux.mu.Unlock()

	current.mu.Lock()
	slog.Debug("mux: registering announcement writer", "writer", w)
	current.writers[w] = struct{}{}
	current.mu.Unlock()

	var announce func(node *announcingNode)
	announce = func(node *announcingNode) {
		slog.Debug("mux: announcing to node", "node", node)
		node.mu.RLock()
		defer node.mu.RUnlock()

		// Send announcements for this node
		for _, announcement := range node.announcements {
			err := w.SendAnnouncement(announcement)
			if err != nil {
				slog.Error("mux: failed to send announcement", "error", err)
				return
			}
		}

		// Recursively announce to child nodes
		for _, child := range node.children {
			announce(child)
		}
	}

	announce(current)
}

func (mux *TrackMux) Clear() {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.announcementTree = *newAnnouncingNode()

	mux.handlerIndex = make(map[BroadcastPath]TrackHandler)
}

// newAnnouncingNode creates and initializes a new routing tree node.
func newAnnouncingNode() *announcingNode {
	return &announcingNode{
		announcements: make(map[BroadcastPath]*Announcement),
		writers:       make(map[AnnouncementWriter]struct{}),
		children:      make(map[string]*announcingNode),
	}
}

type announcingNode struct {
	mu sync.RWMutex

	announcements map[BroadcastPath]*Announcement
	writers       map[AnnouncementWriter]struct{}

	children map[string]*announcingNode
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

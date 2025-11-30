package moqt

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

// DefaultMux is the package-level TrackMux used by convenience top-level functions such as
// Publish and Announce. It provides a global multiplexer suitable for simple server
// implementations.
var DefaultMux *TrackMux = defaultMux

var defaultMux = NewTrackMux()

// NewTrackMux creates a new trackMux for handling track and announcement routing.
// It initializes the routing and announcement trees with empty root nodes.
func NewTrackMux() *TrackMux {
	return &TrackMux{
		announcementTree:  *newAnnouncingNode(""),
		trackHandlerIndex: make(map[BroadcastPath]*announcedTrackHandler),
	}
}

// Publish registers the handler for the given track path in the DefaultMux.
// The handler remains active until the provided context is canceled.
// This is a convenience wrapper around DefaultMux.Publish.
func Publish(ctx context.Context, path BroadcastPath, handler TrackHandler) {
	DefaultMux.Publish(ctx, path, handler)
}

// PublishFunc is a convenience wrapper that registers a simple handler
// function for the given track path on the DefaultMux.
func PublishFunc(ctx context.Context, path BroadcastPath, f func(tw *TrackWriter)) {
	DefaultMux.PublishFunc(ctx, path, f)
}

// Announce registers an Announcement and associated handler in the
// DefaultMux. It is used to publish an Announcement object alongside
// a TrackHandler that will serve any subscribers of the announced path.
func Announce(announcement *Announcement, handler TrackHandler) {
	DefaultMux.Announce(announcement, handler)
}

// TrackMux is a multiplexer for routing track requests and announcements.
// It maintains separate trees for track routing and announcements.
// TrackMux routes announcements and subscribe requests to the correct TrackHandler.
// It keeps an index of broadcast paths to handlers and an announcement routing tree
// that efficiently notifies listeners of announcements matching a prefix.
type TrackMux struct {
	mu                sync.RWMutex
	trackHandlerIndex map[BroadcastPath]*announcedTrackHandler

	announcementTree announcingNode
	// treeMu           sync.RWMutex
}

// PublishFunc registers a simple function handler for the provided path on
// the TrackMux. It wraps the function into a TrackHandlerFunc.
func (mux *TrackMux) PublishFunc(ctx context.Context, path BroadcastPath, f func(tw *TrackWriter)) {
	mux.Publish(ctx, path, TrackHandlerFunc(f))
}

// Handle registers the handler for the given track path on the TrackMux.
// The handler remains active until the provided context is canceled.
// Publish registers the handler for a specific track path on the TrackMux.
// The handler remains active until the provided context is canceled.
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

func (mux *TrackMux) registerHandler(ann *Announcement, handler TrackHandler) *announcedTrackHandler {
	path := ann.BroadcastPath()
	mux.mu.Lock()
	announced, ok := mux.trackHandlerIndex[path]

	newHandler := &announcedTrackHandler{
		Announcement: ann,
		TrackHandler: handler,
	}
	mux.trackHandlerIndex[path] = newHandler
	mux.mu.Unlock()

	if ok {
		announced.end()
	}

	return newHandler
}

func (mux *TrackMux) removeHandler(handler *announcedTrackHandler) {
	path := handler.BroadcastPath()
	mux.mu.Lock()
	defer mux.mu.Unlock()
	if ath, ok := mux.trackHandlerIndex[path]; ok && ath == handler {
		delete(mux.trackHandlerIndex, path)
	}
}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	if announcement == nil {
		slog.Debug("[TrackMux] Announce called with nil Announcement")
		return
	}

	path := announcement.path

	if !announcement.IsActive() {
		slog.Debug("[TrackMux] announcement is not active",
			"path", path,
		)
		return
	}

	// Add the handler to the mux if it is not already registered
	announced := mux.registerHandler(announcement, handler)

	prefixSegments, _ := pathSegments(announcement.BroadcastPath())

	// Add announcement to the announcement tree so that init captures it
	current := &mux.announcementTree

	// Build a list of nodes from root to leaf and add the announcement to each node
	nodes := []*announcingNode{current}
	for _, seg := range prefixSegments {
		current = current.getChild(seg)
		nodes = append(nodes, current)
	}

	// Reserve a subscription slice to reuse across nodes and avoid allocations
	type awChan struct {
		aw *AnnouncementWriter
		ch chan *Announcement
	}
	subs := make([]awChan, 0, 8)

	for _, node := range nodes {
		node.addAnnouncement(announcement)

		// Snapshot subscriptions under RLock and send without holding the lock
		node.mu.RLock()
		subs = subs[:0]
		for aw, ch := range node.subscriptions {
			subs = append(subs, awChan{aw: aw, ch: ch})
		}
		node.mu.RUnlock()

		for _, ac := range subs {
			ch := ac.ch
			// Non-blocking send to avoid deadlocks; drop if buffer is full
			select {
			case ch <- announcement:
			case <-announcement.Done():
			default:
				// Drop the message if the subscriber is not keeping up. Remove subscription
				// from the node and close the channel safely to signal the writer side.
				// delete by AnnouncementWriter pointer without scanning
				node.mu.Lock()
				delete(node.subscriptions, ac.aw)
				node.mu.Unlock()
				// Close the AW to signal the writer to cleanup and close its channel.
				go func(a *AnnouncementWriter) {
					// Use InternalAnnounceErrorCode to indicate an internal error condition
					a.CloseWithError(InternalAnnounceErrorCode)
				}(ac.aw)
			}
		}
	}

	lastNode := current

	// Send announcement to all registered channels with retry mechanism
	// lastNode.mu.RLock()
	// for ch := range lastNode.channels {
	// 	go func(channel chan *Announcement) {
	// 		defer func() { recover() }() // ignore panic on closed channel
	// 		select {
	// 		case channel <- announcement:
	// 			// Successfully sent to channel
	// 		default:
	// 			// Channel is busy, start retry goroutine
	// 			ticker := time.NewTicker(10 * time.Millisecond)
	// 			defer ticker.Stop()

	// 			for {
	// 				select {
	// 				case <-announcement.Done():
	// 					// Announcement ended, no need to send
	// 					return
	// 				default:
	// 					// Check if context is done before sending
	// 					if !announced.IsActive() {
	// 						return
	// 					}
	// 					select {
	// 					case channel <- announcement:
	// 						// Successfully sent to channel
	// 						return
	// 					case <-ticker.C:
	// 						// Timeout, retry
	// 						continue
	// 					case <-announcement.Done():
	// 						// Announcement ended during send attempt
	// 						return
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}(ch)
	// }
	// lastNode.mu.RUnlock()

	// Ensure the announcement is removed when it ends
	announcement.AfterFunc(func() {
		// Remove the announcement from the tree unconditionally
		lastNode.removeAnnouncement(announcement)

		mux.removeHandler(announced)
	})
}

// TrackHandler returns the Announcement and associated TrackHandler for the specified broadcast path.
// If no handler is found, TrackHandler returns nil and NotFoundTrackHandler.
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

	mux.mu.RLock()
	ath, ok := mux.trackHandlerIndex[path]
	mux.mu.RUnlock()
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

	mux.mu.RLock()
	announced, ok := mux.trackHandlerIndex[path]
	mux.mu.RUnlock()
	if !ok || announced == nil || announced.Announcement == nil {
		slog.Warn("[TrackMux] no announcement found for path", "path", path)
		tw.CloseWithError(TrackNotFoundErrorCode)
		return
	}

	if announced.TrackHandler == nil {
		slog.Warn("mux: no handler found for path", "path", path)
		tw.CloseWithError(TrackNotFoundErrorCode)
		return
	}

	// Ensure track is closed when announcement ends
	stop := announced.AfterFunc(func() {
		tw.Close()
	})
	defer stop()

	announced.TrackHandler.ServeTrack(tw)
}

// serveAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) serveAnnouncements(aw *AnnouncementWriter) {
	if aw == nil {
		slog.Error("mux: nil announcement writer")
		return
	}

	slog.Debug("serveAnnouncements start", "prefix", aw.prefix)

	if !isValidPrefix(aw.prefix) {
		aw.CloseWithError(InvalidPrefixErrorCode)
		return
	}

	// The AnnouncementWriter lifecycle is managed by the session/creator;
	// Do not close the writer here to avoid double-close semantics and to let
	// the caller (e.g. session) decide when to close the stream.

	// Register the handler on the routing tree (protect children with per-leafNode lock)
	leafNode := mux.announcementTree.createNode(prefixSegments(aw.prefix))
	// for _, seg := range prefixSegments(aw.prefix) {
	// 	current.mu.Lock()
	// 	if current.children == nil {
	// 		current.children = make(map[string]*announcingNode)
	// 	}
	// 	child, ok := current.children[seg]
	// 	if !ok {
	// 		child = newAnnouncingNode()
	// 		current.children[seg] = child
	// 	}
	// 	next := child
	// 	current.mu.Unlock()
	// 	current = next
	// }

	// ch := make(chan *Announcement, 8) // TODO: configurable buffer size
	// current.mu.Lock()
	// current.subscriptions[aw] = ch
	// current.mu.Unlock()

	// // Cleanup channel when done
	// defer func() {
	// 	current.mu.Lock()
	// 	delete(current.subscriptions, aw)
	// 	current.mu.Unlock()
	// 	close(ch)
	// }()

	// current.serve(aw)

	// // Get existing announcements in a separate goroutine to avoid blocking new announcements
	// current.mu.RLock()
	// err := aw.init(current.announcements)
	// current.mu.RUnlock()
	// if err != nil {
	// 	slog.Error("[TrackMux] failed to initialize announcement writer", "error", err)
	// 	aw.CloseWithError(InternalAnnounceErrorCode)
	// 	return
	// }

	// Process announcements from channel
	// for {
	// 	select {
	// 	case ann, ok := <-ch:
	// 		if !ok {
	// 			return // Channel closed
	// 		}
	// 		err := aw.SendAnnouncement(ann)
	// 		if err != nil {
	// 			slog.Error("[TrackMux] failed to send announcement", "error", err)
	// 			aw.CloseWithError(InternalAnnounceErrorCode)
	// 			return
	// 		}
	// 	case <-aw.Context().Done():
	// 		// Writer context cancelled
	// 		return
	// 	}
	// }

	leafNode.mu.Lock()

	// Snapshot current active announcements
	actives := make(map[*Announcement]struct{})
	for ann := range leafNode.announcements {
		actives[ann] = struct{}{}
	}

	// Channel to receive announcements
	ch := make(chan *Announcement, 8) // TODO: configurable buffer size
	if leafNode.subscriptions == nil {
		leafNode.subscriptions = make(map[*AnnouncementWriter](chan *Announcement))
	}
	leafNode.subscriptions[aw] = ch
	leafNode.mu.Unlock()

	defer func() {
		leafNode.mu.Lock()
		delete(leafNode.subscriptions, aw)
		leafNode.mu.Unlock()
	}()

	err := aw.init(actives)
	if err != nil {
		slog.Error("[TrackMux] failed to initialize announcement writer", "error", err)
		aw.CloseWithError(InternalAnnounceErrorCode)
		return
	}

	// Process announcements and exit when writer context is cancelled
	for {
		select {
		case ann, ok := <-ch:
			if !ok {
				return
			}
			if err := aw.SendAnnouncement(ann); err != nil {
				aw.CloseWithError(InternalAnnounceErrorCode)
				return
			}
		case <-aw.Context().Done():
			return
		}
	}
}

// Clear removed: previously used for resetting state in tests. Use NewTrackMux() for test isolation or implement a shutdown API for production.

// newAnnouncingNode creates and initializes a new routing tree node.
func newAnnouncingNode(segment prefixSegment) *announcingNode {
	return &announcingNode{
		prefixSegment: segment,
		announcements: make(map[*Announcement]struct{}),
		children:      make(map[string]*announcingNode),
		subscriptions: make(map[*AnnouncementWriter](chan *Announcement)),
	}
}

type prefixSegment = string

type announcingNode struct {
	mu sync.RWMutex

	parent *announcingNode

	prefixSegment prefixSegment

	children map[prefixSegment]*announcingNode

	// channels map[chan *Announcement]struct{}
	subscriptions map[*AnnouncementWriter](chan *Announcement)

	announcements map[*Announcement]struct{}
}

func (node *announcingNode) getChild(seg prefixSegment) *announcingNode {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.children == nil {
		node.children = make(map[string]*announcingNode)
	}
	child, ok := node.children[seg]
	if !ok {
		child = newAnnouncingNode(seg)
		child.parent = node
		node.children[seg] = child
	}
	return child
}

// addAnnouncement adds an announcement to the tree by traversing from parent to child nodes.
func (node *announcingNode) addAnnouncement(announcement *Announcement) {
	node.mu.Lock()
	defer node.mu.Unlock()

	if node.announcements == nil {
		node.announcements = make(map[*Announcement]struct{})
	}

	node.announcements[announcement] = struct{}{}
}

// removeAnnouncement removes an announcement from the tree by traversing from child to parent nodes.
func (node *announcingNode) removeAnnouncement(announcement *Announcement) {
	node.mu.Lock()
	delete(node.announcements, announcement)
	shouldRemove := len(node.announcements) == 0 && len(node.children) == 0
	node.mu.Unlock()

	if shouldRemove && node.parent != nil {
		node.parent.mu.Lock()
		delete(node.parent.children, node.prefixSegment)
		node.parent.mu.Unlock()
		node.parent.removeAnnouncement(announcement)
	}
}

func (node *announcingNode) createNode(segments []prefixSegment) *announcingNode {
	if len(segments) == 0 {
		return node
	}

	child := node.getChild(segments[0])

	return child.createNode(segments[1:])
}

func isValidPath(path BroadcastPath) bool {
	if path == "" {
		return false
	}

	p := string(path)

	return strings.HasPrefix(p, "/")
}

func isValidPrefix(prefix string) bool {
	if prefix == "" {
		return false
	}

	return strings.HasPrefix(prefix, "/") && strings.HasSuffix(prefix, "/")
}

// TrackHandler handles a published track.
// Implementations will be invoked when a subscriber requests a track and are provided with a
// TrackWriter to send group frames for that track.
type TrackHandler interface {
	ServeTrack(*TrackWriter)
}

// NotFound is a default convenience handler function which responds to
// subscribers by closing the track writer with a TrackNotFound error.
var NotFound = func(tw *TrackWriter) {
	if tw == nil {
		return
	}

	tw.CloseWithError(TrackNotFoundErrorCode)
}

// NotFoundTrackHandler is a TrackHandler that implements a not-found
// behavior by calling NotFound handler.
var NotFoundTrackHandler TrackHandler = TrackHandlerFunc(NotFound)

// TrackHandlerFunc is an adapter to allow ordinary functions to act as a
// TrackHandler. It implements the TrackHandler interface.
type TrackHandlerFunc func(*TrackWriter)

func (f TrackHandlerFunc) ServeTrack(tw *TrackWriter) {
	f(tw)
}

var _ TrackHandler = (*announcedTrackHandler)(nil)

type announcedTrackHandler struct {
	TrackHandler
	*Announcement
}

func prefixSegments(prefix string) []prefixSegment {
	segments := strings.Split(prefix, "/")
	return segments[1 : len(segments)-1]
}

func pathSegments(path BroadcastPath) (prefixSegments []prefixSegment, last string) {
	segments := strings.Split(string(path), "/")
	return segments[1 : len(segments)-1], segments[len(segments)-1]
}

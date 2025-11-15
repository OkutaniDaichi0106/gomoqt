package moqt

import (
	"context"
	"fmt"
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
		announcementTree:  *newAnnouncingNode(""),
		trackHandlerIndex: make(map[BroadcastPath]*announcedTrackHandler),
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
	handlerMu         sync.RWMutex
	trackHandlerIndex map[BroadcastPath]*announcedTrackHandler

	announcementTree announcingNode
	// treeMu           sync.RWMutex
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

func (mux *TrackMux) registerHandler(ann *Announcement, handler TrackHandler) *announcedTrackHandler {
	path := ann.BroadcastPath()
	mux.handlerMu.Lock()
	announced, ok := mux.trackHandlerIndex[path]

	newHandler := &announcedTrackHandler{
		Announcement: ann,
		TrackHandler: handler,
	}
	mux.trackHandlerIndex[path] = newHandler
	mux.handlerMu.Unlock()

	if ok {
		announced.end()
	}

	return newHandler
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

	current := &mux.announcementTree
	for _, seg := range prefixSegments {
		current = current.getChild(seg)
		// Snapshot subscriptions under RLock and send without holding the lock
		current.mu.RLock()
		subs := make([]chan *Announcement, 0, len(current.subscriptions))
		for _, ch := range current.subscriptions {
			subs = append(subs, ch)
		}
		current.mu.RUnlock()
		for _, ch := range subs {
			// Non-blocking send to avoid deadlocks; drop if buffer is full
			select {
			case ch <- announcement:
			case <-announcement.Done():
			default:
				// Retry sending announcement asynchronously if buffer is full
				go func() {
					select {
					case ch <- announcement:
					case <-announcement.Done():
					}
				}()
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
	announced, ok := mux.trackHandlerIndex[path]
	mux.handlerMu.RUnlock()
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

	announced.TrackHandler.ServeTrack(tw)

	// Stop the announcement watcher when done
	stop()
}

// serveAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) serveAnnouncements(aw *AnnouncementWriter) {
	if aw == nil {
		slog.Error("mux: nil announcement writer")
		return
	}
	slog.Debug("serveAnnouncements start", "prefix", aw.prefix)
	// debug
	fmt.Printf("serveAnnouncements: prefix=%q valid=%v\n", aw.prefix, isValidPrefix(aw.prefix))
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

// // addAnnouncement adds an announcement to the tree by traversing from parent to child nodes.
// func (node *announcingNode) addAnnouncement(prefixSegments []string, announcement *Announcement) {
// 	// store announcement on this node
// 	node.mu.Lock()
// 	node.announcements[announcement] = struct{}{}
// 	node.mu.Unlock()

// 	// if no further segments, stop here
// 	if len(prefixSegments) == 0 {
// 		announcement.AfterFunc(func() {
// 			node.removeAnnouncement(announcement)
// 		})
// 		return
// 	}

// 	// create or find the child for the next segment and recurse
// 	child := node.getChild(prefixSegments[0])

// 	child.addAnnouncement(prefixSegments[1:], announcement)
// }

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

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

// DefaultMux is the default TrackMux used by the top-level functions.
// It can be used directly instead of creating a new TrackMux.
var DefaultMux *TrackMux = defaultMux

var defaultMux = NewTrackMux()

// NewTrackMux creates a new TrackMux for handling track and announcement routing.
// It initializes the routing and announcement trees with empty root nodes.
func NewTrackMux() *TrackMux {
	return &TrackMux{
		trackTree:    *newRoutingNode(),
		announceTree: *newAnnouncingNode(),
	}
}

// Handle registers the handler for the given track path in the DefaultMux.
// The handler will remain active until the context is canceled.
func Handle(ctx context.Context, path TrackPath, handler Handler) {
	DefaultMux.Handle(ctx, path, handler)
}

// ServeTrack serves the track at the specified path to the given TrackWriter using DefaultMux.
// It finds the appropriate handler for the path and delegates the serving to it.
func ServeTrack(w TrackWriter, config *SubscribeConfig) {
	DefaultMux.ServeTrack(w, config)
}

// ServeAnnouncements serves announcements for tracks matching the given pattern using DefaultMux.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	DefaultMux.ServeAnnouncements(w, config)
}

// GetInfo retrieves information about the track at the specified path using DefaultMux.
// Returns track information and any error encountered during the lookup.
func GetInfo(path TrackPath) (Info, error) {
	return DefaultMux.GetInfo(path)
}

// BuildTrack creates a new TrackBuffer for the specified path with the given info and expiration time.
// It registers the TrackBuffer as a handler in DefaultMux and returns it.
func BuildTrack(ctx context.Context, path TrackPath, info Info, expires time.Duration) *TrackBuffer {
	return DefaultMux.BuildTrack(ctx, path, info, expires)
}

// func Unhandle(path TrackPath) bool {
// 	return DefaultMux.Unhandle(path)
// }

// TrackMux implements the Handler interface.
var _ Handler = (*TrackMux)(nil)

// TrackMux is a multiplexer for routing track requests and announcements.
// It maintains separate trees for track routing and announcements, allowing efficient
// lookup of handlers and distribution of announcements to interested subscribers.
type TrackMux struct {
	mu sync.RWMutex

	// trackTree is the root node of the routing tree for tracks.
	// It is used to find the handler for a given track path.
	trackTree routingNode

	// announceTree is the root node of the announcement tree.
	// It is used to announce new tracks to existing announcement writers.
	announceTree announcingNode
}

// Handle registers the handler for the given track path.
// The handler will remain active until the context is canceled.
func (mux *TrackMux) Handle(ctx context.Context, path TrackPath, handler Handler) {
	if path == "" {
		slog.Warn("mux: empty track path")
		return
	}

	p := newPath(path)

	mux.mu.Lock()
	defer mux.mu.Unlock()

	// Register the handler on the routing tree
	mux.registerHandler(ctx, p, handler)

	// Announce the track to all announcement writers
	mux.announce(p, handler)

	slog.Debug("registered track handler",
		"track_path", path,
	)
}

// Handler returns the handler for the specified track path.
// If no handler is found, NotFoundHandler is returned.
func (mux *TrackMux) Handler(path TrackPath) Handler {
	if path == "" {
		slog.Warn("mux: empty track path for handler lookup")
		return NotFoundHandler
	}

	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPath(path)

	// Find the handler for the given path
	current := &mux.trackTree
	for _, seg := range p.segments {
		if current.children == nil {
			return NotFoundHandler
		}

		child, ok := current.children[seg]
		if !ok {
			return NotFoundHandler
		}

		current = child
	}

	if current.handler == nil {
		return NotFoundHandler
	}

	return current.handler
}

// BuildTrack creates a new TrackBuffer for the specified path with the given info and expiration time.
// It registers the TrackBuffer as a handler and returns it.
func (mux *TrackMux) BuildTrack(ctx context.Context, path TrackPath, info Info, expires time.Duration) *TrackBuffer {
	if path == "" {
		slog.Warn("mux: empty track path for track building")
		return nil
	}

	buf := NewTrackBuffer(path, info, expires)

	mux.Handle(ctx, path, buf)

	return buf
}

// ServeTrack serves the track at the specified path to the given TrackWriter.
// It finds the appropriate handler for the path and delegates the serving to it.
func (mux *TrackMux) ServeTrack(w TrackWriter, config *SubscribeConfig) {
	if w == nil {
		slog.Error("mux: nil track writer")
		return
	}

	path := w.TrackPath()
	if path == "" {
		slog.Warn("mux: empty track path for serving")
		return
	}

	mux.mu.RLock()

	h := mux.findTrackHandler(newPath(path))

	mux.mu.RUnlock()

	if h == NotFoundHandler {
		slog.Debug("track not found for serving", "track_path", path)
	} else if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("serving track", "track_path", path)
	}

	h.ServeTrack(w, config)
}

// ServeAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	if w == nil {
		slog.Error("mux: nil announcement writer")
		return
	}

	if config == nil || config.TrackPattern == "" {
		slog.Warn("mux: empty or nil track pattern for announcements")
		return
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	pattern := newPattern(config.TrackPattern)

	// Register the handler on the routing tree
	current := &mux.announceTree
	for _, seg := range pattern.segments {
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

	// Set the handler on the leaf node
	if current.pattern == nil && current.buffer == nil {
		current.setPattern(pattern, config)
	}

	// Start serving announcements
	current.buffer.deliverAnnouncements(w)
}

// GetInfo retrieves information about the track at the specified path.
// Returns track information and any error encountered during the lookup.
func (mux *TrackMux) GetInfo(path TrackPath) (Info, error) {
	if path == "" {
		slog.Warn("mux: empty track path for info lookup")
		return NotFoundInfo, ErrTrackDoesNotExist
	}

	p := newPath(path)

	info, err := mux.trackTree.getInfo(0, p)
	if err != nil {
		slog.Debug("track info not found", "track_path", path, "error", err)
		return NotFoundInfo, err
	}

	return info, nil
}

// registerHandler registers the handler for the given track path in the routing tree.
// It traverses the tree to find the appropriate node and sets the handler at the leaf node.
func (mux *TrackMux) registerHandler(ctx context.Context, path *path, handler Handler) {
	// Register the handler on the routing tree
	current := &mux.trackTree
	var parents []*routingNode
	var segments []string

	for _, seg := range path.segments {
		parents = append(parents, current)
		segments = append(segments, seg)

		if current.children == nil {
			current.children = make(map[string]*routingNode)
		}

		if child, ok := current.children[seg]; ok {
			current = child
		} else {
			child := newRoutingNode()
			current.children[seg] = child
			current = child
		}
	}

	// Check if this node has a handler
	if current.handler != nil {
		slog.Warn("mux: overwriting existing handler", "path", path.str)

		// Cancel the previous handler
		if current.cancelFunc != nil {
			slog.Debug("cancelling previous handler", "track_path", path.str)

			// Unlock the mutex to avoid deadlock
			mux.mu.Unlock()

			// Signal the previous handler to stop handling the track and clean up resources.
			current.cancelFunc()

			// Lock the mutex again
			mux.mu.Lock()
		}

		// Reset the fields of the node
		current.handler = nil
		current.pattern = nil
		current.cancelFunc = nil
	}

	// Wrap the context with a cancel function
	ctx, cancelFunc := context.WithCancel(ctx)

	// Set the handler on the leaf node
	current.setHandler(path, handler, cancelFunc)

	slog.Debug("registered a handler", "track_path", path.str)

	go func() {
		<-ctx.Done()

		mux.mu.Lock()
		defer mux.mu.Unlock()

		if current.handler == nil {
			return
		}

		// Remove the handler
		current.handler = nil
		current.pattern = nil

		for i := len(parents) - 1; i > 0; i-- {
			parent := parents[i]
			seg := segments[i]

			node := parent.children[seg]
			if node.handler == nil && len(node.children) == 0 {
				delete(parent.children, seg)
			}
		}

		// TODO: Unannounce the track

		slog.Debug("unregistered a handler", "track_path", path.str)
	}()
}

func (mux *TrackMux) findTrackHandler(path *path) TrackHandler {
	return mux.trackTree.findTrackHandler(0, path)
}

// func (mux *TrackMux) removeHandler(path *path) bool {
// 	if len(path.segments) == 0 {
// 		return true
// 	}

// 	// Navigate to the correct node
// 	current := &mux.trackTree
// 	var parents []*routingNode
// 	var segments []string

// 	for _, seg := range path.segments {
// 		parents = append(parents, current)
// 		segments = append(segments, seg)

// 		if current.children == nil {
// 			return false
// 		}

// 		child, ok := current.children[seg]
// 		if !ok {
// 			return false
// 		}

// 		current = child
// 	}

// 	// Check if this node has a handler
// 	if current.handler == nil {
// 		return false
// 	}

// 	// Remove the handler
// 	current.handler = nil
// 	current.pattern = nil

// 	// Clean up empty nodes (nodes without children and handlers)
// 	// Start from the leaf node and move up the tree
// 	for i := len(parents) - 1; i >= 0; i-- {
// 		parent := parents[i]
// 		seg := segments[i]

// 		node := parent.children[seg]
// 		if node.handler == nil && len(node.children) == 0 {
// 			delete(parent.children, seg)
// 		}
// 	}

// 	slog.Debug("unregistered a handler", "track_path", path.str)
// 	return true
// }

// announce dispatches track announcements to all registered announcement writers
// that match the given path.
func (mux *TrackMux) announce(path *path, handler AnnouncementHandler) {
	// Announce the track to all announcement writers
	mux.announceTree.announce(path.segments[1:], handler)

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("announced new track",
			"track_path", path.str,
		)
	}
}

// newRoutingNode creates and initializes a new routing tree node.
func newRoutingNode() *routingNode {
	return &routingNode{
		children: make(map[string]*routingNode),
	}
}

// routingNode represents a node in the track routing tree.
// It contains references to child nodes and may contain a handler if it's a leaf node.
type routingNode struct {
	// If this node is a leaf node, pattern and handler are set.
	// If this node is not a leaf node, pattern and handler are nil.
	pattern    *path
	handler    Handler
	cancelFunc context.CancelFunc

	children map[string]*routingNode
}

// setHandler sets the handler, pattern and cancel function for a routing node.
// This method is called when registering a new handler at a leaf node in the routing tree.
func (node *routingNode) setHandler(ptn *path, handler Handler, cancelFunc context.CancelFunc) {
	if ptn == nil {
		panic("mux: nil pattern")
	} else if handler == nil {
		panic("mux: nil handler")
	}

	node.pattern = ptn
	node.handler = handler
	node.cancelFunc = cancelFunc
}

// findTrackHandler searches for a handler matching the given path starting at this node.
// It recursively traverses the routing tree to find the appropriate handler for the path.
// Returns NotFoundHandler if no matching handler is found.
func (node *routingNode) findTrackHandler(depth int, path *path) TrackHandler {
	// If we've gone past the path depth or this is a leaf node, check if it has a handler
	if depth > path.depth() {
		slog.Debug("mux: path depth exceeded", "path", path.str, "depth", depth)
		return NotFoundHandler
	} else if depth == path.depth() {
		if node.handler == nil {
			slog.Debug("mux: no handler at leaf node", "path", path.str)
			return NotFoundHandler
		}

		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			slog.Debug("mux: handler found",
				"path", path.str,
			)
		}
		return node.handler
	}

	// If this node is not a leaf node, get the next segment and look for a matching child
	segment := path.segments[depth]
	child, ok := node.children[segment]
	if !ok {
		slog.Debug("mux: no child node for segment",
			"path", path.str,
			"segment", segment,
			"depth", depth)
		return NotFoundHandler
	}

	// Continue traversal with the child node at the next depth
	return child.findTrackHandler(depth+1, path)
}

// // serveAnnouncements sends announcements to the given writer for tracks that match the pattern.
// // It traverses the routing tree to find handlers that match the pattern and calls their
// // ServeAnnouncements method.
// //
// // Parameters:
// //   - depth: Current depth in the pattern tree
// //   - pattern: The pattern to match against tracks
// //   - w: The announcement writer to send announcements to
// //   - config: Configuration for the announcement service
// func (node *routingNode) serveAnnouncements(depth int, pattern *pattern, w AnnouncementWriter, config *AnnounceConfig) {
// 	// If we've reached a matching leaf node with a handler, serve announcements
// 	if depth == pattern.depth() && node.handler != nil {
// 		if h, ok := node.handler.(AnnouncementHandler); ok {
// 			slog.Debug("serving announcement",
// 				"path", node.pattern.str,
// 				"pattern", pattern.str,
// 			)
// 			h.ServeAnnouncements(w, config)
// 		}
// 		return
// 	}

// 	// Stop recursion if we've gone past the pattern depth or this node has no children
// 	if depth >= pattern.depth() || node.children == nil || len(node.children) == 0 {
// 		return
// 	}

// 	// Get the current segment to match in the pattern
// 	segment := pattern.segments[depth+1]

// 	// Case 1: Exact segment match
// 	if child, ok := node.children[segment]; ok {
// 		child.serveAnnouncements(depth+1, pattern, w, config)
// 	}

// 	// Case 2: Single wildcard match (*) - matches any single segment
// 	if segment == "*" {
// 		for childSegment, child := range node.children {
// 			slog.Debug("* wildcard matching child", "segment", childSegment)
// 			child.serveAnnouncements(depth+1, pattern, w, config)
// 		}
// 	}

// 	// Case 3: Double wildcard match (**) - matches any number of segments
// 	if segment == "**" {
// 		// Match zero segments by continuing at the next pattern depth
// 		node.serveAnnouncements(depth+1, pattern, w, config)

// 		// Match one or more segments by recursively checking all children
// 		for childSegment, child := range node.children {
// 			slog.Debug("** wildcard matching child", "segment", childSegment)
// 			// Keep the same pattern depth to continue matching with ** wildcard
// 			child.serveAnnouncements(depth, pattern, w, config)
// 		}
// 	}
// }

// getInfo retrieves track information for the given path starting at this node.
// It recursively traverses the routing tree to find the handler for the path,
// and then calls GetInfo on that handler.
// Returns NotFoundInfo and an error if no matching handler is found.
func (node *routingNode) getInfo(depth int, path *path) (Info, error) {
	if depth > path.depth() {
		if node.handler == nil {
			slog.Debug("mux: no handler at node for info", "path", path.str)
			return NotFoundInfo, ErrTrackDoesNotExist
		}

		info, err := node.handler.GetInfo(TrackPath(path.str))
		if err != nil {
			slog.Debug("track info retrieval failed",
				"path", path.str,
				"error", err,
			)
		}
		return info, err
	}

	// If we haven't reached the end of the path, continue traversal
	segment := path.segments[depth]
	child, ok := node.children[segment]
	if !ok {
		slog.Debug("mux: no child node for info segment",
			"path", path.str,
			"segment", segment,
			"depth", depth)
		return NotFoundInfo, ErrTrackDoesNotExist
	}

	return child.getInfo(depth+1, path)
}

// newAnnouncingNode creates and initializes a new announcement tree node.
func newAnnouncingNode() *announcingNode {
	return &announcingNode{
		children: make(map[string]*announcingNode),
	}
}

// announcingNode is a node in the announcement tree.
// It maintains a list of announcement writers that are interested in tracks matching its pattern.
type announcingNode struct {
	// If this node is a leaf node, pattern and config are set.
	// If this node is not a leaf node, pattern and config are nil.
	pattern *pattern

	// config is the announcement configuration for this node
	config *AnnounceConfig

	//
	buffer *announcementsBuffer

	// children maps segment names to child nodes
	children map[string]*announcingNode
}

// setPattern sets the pattern and announcement configuration for this node.
// This is called when registering a new announcement writer at a leaf node.
func (node *announcingNode) setPattern(pattern *pattern, config *AnnounceConfig) {
	if config == nil {
		panic("mux: nil announce config")
	}

	node.pattern = pattern
	node.config = config
	node.buffer = newAnnouncementsBuffer()
}

// announce dispatches announcements to all registered announcement writers that match the given path.
// It traverses the announcement tree and delivers announcements to all matching nodes.
//
// Parameters:
//   - segments: The path segments to match against patterns in the tree
//   - handler: The handler that will serve announcements to matching writers
func (node *announcingNode) announce(segments []string, handler AnnouncementHandler) {
	if len(segments) == 0 {
		// Serve announcements to the buffer
		slog.Debug("mux: serving announcements to buffer",
			"pattern", node.pattern.str,
		)
		go handler.ServeAnnouncements(node.buffer, node.config)
		return
	}

	if node.children == nil {
		return
	}

	segment := segments[0]
	child, ok := node.children[segment]
	if !ok {
		return
	}

	child.announce(segments[1:], handler)
}

// newPath creates a path from a TrackPath.
// It converts the string-based path to a structured path for efficient routing.
func newPath(p TrackPath) *path {
	if p == "" {
		slog.Warn("mux: creating pattern from empty string")
		p = "/"
	}

	str := string(p)

	if !strings.HasPrefix(str, "/") {
		slog.Error("mux: pattern must start with '/'", "pattern", str)
		panic("mux: pattern must start with '/'")
	}
	return &path{
		str:      p,
		segments: strings.Split(str, "/"),
	}
}

// path represents a structured track path for routing.
// It is derived from a pattern but used specifically for exact matching.
type path struct {
	// str is the original string representation of the path
	str TrackPath

	// segments is the path split into segments
	segments []string
}

// depth returns the effective depth of the path (number of segments minus 1).
// The depth is used to determine when we've reached the end of a path in the routing tree.
func (p *path) depth() int {
	return len(p.segments) - 1
}

// newPattern creates a pattern from a string.
// It validates and splits the string into segments for pattern matching.
func newPattern(str string) *pattern {
	if str == "" {
		slog.Warn("mux: creating pattern from empty string")
		str = "/"
	}

	if !strings.HasPrefix(str, "/") {
		slog.Error("mux: pattern must start with '/'", "pattern", str)
		panic("mux: pattern must start with '/'")
	}

	return &pattern{
		str:      str,
		segments: strings.Split(str, "/"),
	}
}

// pattern represents a track matching pattern that may include wildcards.
// It is used for both exact matching and wildcard matching in the router.
type pattern struct {
	// str is the original string representation of the pattern
	str string

	// segments is the path split into segments
	segments []string
}

// depth returns the effective depth of the pattern (number of segments minus 1).
// The depth is used to determine when we've reached the end of a pattern in the matching algorithm.
func (p *pattern) depth() int {
	return len(p.segments) - 1
}

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
		trackTree:    *newRoutingNode(),
		announceTree: *newAnnouncingNode(),
	}
}

// Handle registers the handler for the given track path in the DefaultMux.
// The handler will remain active until the context is canceled.
func Handle(ctx context.Context, path TrackPath, handler TrackHandler) {
	DefaultMux.Handle(ctx, path, handler)
}

func Announce(announcement *Announcement, handler TrackHandler) {
	DefaultMux.Announce(announcement, handler)
}

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
func (mux *TrackMux) Handle(ctx context.Context, path TrackPath, handler TrackHandler) {
	mux.Announce(NewAnnouncement(ctx, path), handler)
}

func (mux *TrackMux) Announce(announcement *Announcement, handler TrackHandler) {
	path := announcement.TrackPath()
	if path == "" {
		slog.Warn("mux: empty track path for announcement")
		return
	}

	if !announcement.IsActive() {
		slog.Warn("mux: announcement is not active")
		return
	}

	p := newPath(path)

	mux.mu.Lock()
	defer mux.mu.Unlock()

	// Register the handler on the routing tree
	mux.registerHandler(p, announcement, handler)

	// Announce the track to all announcement writers
	mux.announce(p, announcement)

	slog.Debug("registered track handler",
		"track_path", path,
	)
}

// Handler returns the handler for the specified track path.
// If no handler is found, NotFoundTrackHandler is returned.
func (mux *TrackMux) Handler(path TrackPath) TrackHandler {
	if path == "" {
		slog.Warn("mux: empty track path for handler lookup")
		return NotFoundTrackHandler
	}

	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPath(path)

	// Find the handler for the given path
	current := &mux.trackTree
	for _, seg := range p.segments {
		if current.children == nil {
			return NotFoundTrackHandler
		}

		child, ok := current.children[seg]
		if !ok {
			return NotFoundTrackHandler
		}

		current = child
	}

	if current.handler == nil {
		return NotFoundTrackHandler
	}

	return current.handler
}

// ServeTrack serves the track at the specified path to the given TrackWriter.
// It finds the appropriate handler for the path and delegates the serving to it.
func (mux *TrackMux) ServeTrack(path TrackPath, w TrackWriter, sub ReceivedSubscription) {
	if w == nil {
		slog.Error("mux: nil track writer")
		return
	}

	// path := sub.TrackPath()

	if path == "" {
		slog.Warn("mux: empty track path for serving")
		return
	}

	mux.mu.RLock()

	h := mux.findTrackHandler(newPath(path))

	mux.mu.RUnlock()

	h.HandleTrack(w, sub)
}

// ServeAnnouncements serves announcements for tracks matching the given pattern.
// It registers the AnnouncementWriter and sends announcements for matching tracks.
func (mux *TrackMux) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	if w == nil {
		slog.Error("mux: nil announcement writer")
		return
	}

	if config == nil || config.TrackPattern == "" {
		config = anyAnnounceConfig
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	pattern := newPattern(config.TrackPattern)

	// Register the handler on the routing tree
	current := &mux.announceTree
	exists := false
	for _, seg := range pattern.segments {
		if current.children == nil {
			current.children = make(map[string]*announcingNode)
		}

		if child, ok := current.children[seg]; ok {
			exists = true
			current = child
		} else {
			child := newAnnouncingNode()
			current.children[seg] = child
			current = child
		}
	}

	if !exists {
		// find all announcements matchs to the pattern
		current.announcements = mux.findAnnouncements(pattern)
	}

	// Start serving announcements
	w.SendAnnouncements(current.announcements)
	pos := len(current.announcements)
	var err error

	for {
		for len(current.announcements) > pos {
			current.cond.Wait()
		}

		next := current.announcements[pos:]

		err = w.SendAnnouncements(next)
		if err != nil {
			return
		}
		pos = len(current.announcements)
	}
}

// GetInfo retrieves information about the track at the specified path.
// Returns track information and any error encountered during the lookup.
func (mux *TrackMux) GetInfo(path TrackPath) (Info, bool) {
	if path == "" {
		slog.Warn("mux: empty track path for handler lookup")
		return Info{}, false
	}

	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPath(path)

	// Find the handler for the given path
	current := &mux.trackTree
	for _, seg := range p.segments {
		if current.children == nil {
			return Info{}, false
		}

		child, ok := current.children[seg]
		if !ok {
			return Info{}, false
		}

		current = child
	}

	if current.handler == nil {
		return Info{}, false
	}

	return current.handler.Info()
}

func (mux *TrackMux) Clear() {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.trackTree = *newRoutingNode()
	mux.announceTree = *newAnnouncingNode()

	// TODO: Test
}

// registerHandler registers the handler for the given track path in the routing tree.
// It traverses the tree to find the appropriate node and sets the handler at the leaf node.
func (mux *TrackMux) registerHandler(path *path, ann *Announcement, handler TrackHandler) {
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
	}

	// Set the handler on the leaf node
	current.path = path
	current.announcement = ann
	current.handler = handler
	slog.Debug("registered a handler", "track_path", path.str)

	go func() {
		<-ann.AwaitEnd()

		mux.mu.Lock()
		defer mux.mu.Unlock()

		if current.handler == nil {
			return
		}

		// Remove the handler
		current.handler = nil
		current.path = nil

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

// announce dispatches track announcements to all registered announcement writers
// that match the given path.
func (mux *TrackMux) announce(path *path, announcement *Announcement) {
	// Announce the track to all announcement writers
	mux.announceTree.announce(path.segments[1:], announcement)

	slog.Debug("announced new track",
		"track_path", path.str,
	)

}

func (mux *TrackMux) findAnnouncements(pattern *pattern) []*Announcement {
	var announcements []*Announcement
	// TODO:
	var search func(node *routingNode, index int, path string)
	search = func(node *routingNode, index int, path string) {
		if index >= pattern.depth() {
			if node.handler != nil && node.announcement.IsActive() {
				announcements = append(announcements, node.announcement)
			}
			return
		}

		segment := pattern.segments[index]

		switch segment {
		case "*":
			for childSeg, childNode := range node.children {
				nextPath := path + "/" + childSeg
				search(childNode, index+1, nextPath)
			}
		case "**":
			search(node, index+1, path)
			for childSeg, childNode := range node.children {
				nextPath := path + "/" + childSeg
				search(childNode, index, nextPath)
			}
		default:
			if childNode, ok := node.children[segment]; ok {
				nextPath := path + "/" + segment
				search(childNode, index, nextPath)
			}
		}
	}
	return announcements
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
	// If this node is a leaf node, path and handler are set.
	// If this node is not a leaf node, path and handler are nil.
	path         *path
	handler      TrackHandler
	announcement *Announcement
	// info         *Info

	children map[string]*routingNode
}

// findTrackHandler searches for a handler matching the given path starting at this node.
// It recursively traverses the routing tree to find the appropriate handler for the path.
// Returns NotFoundTrackHandler if no matching handler is found.
func (node *routingNode) findTrackHandler(depth int, path *path) TrackHandler {
	// If we've gone past the path depth or this is a leaf node, check if it has a handler
	if depth > path.depth() {
		slog.Debug("mux: path depth exceeded", "path", path.str, "depth", depth)
		return NotFoundTrackHandler
	} else if depth == path.depth() {
		if node.handler == nil {
			slog.Debug("mux: no handler at leaf node", "path", path.str)
			return NotFoundTrackHandler
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
		return NotFoundTrackHandler
	}

	// Continue traversal with the child node at the next depth
	return child.findTrackHandler(depth+1, path)
}

// newAnnouncingNode creates and initializes a new announcement tree node.
func newAnnouncingNode() *announcingNode {
	node := &announcingNode{
		children: make(map[string]*announcingNode),
	}
	node.cond = sync.NewCond(&node.mu)

	return node
}

// announcingNode is a node in the announcement tree.
// It maintains a list of announcement writers that are interested in tracks matching its pattern.
type announcingNode struct {
	// If this node is a leaf node, pattern and config are set.
	// If this node is not a leaf node, pattern and config are nil.
	pattern *pattern

	// config is the announcement configuration for this node
	// config *AnnounceConfig

	//
	// buffer *announcementsBuffer
	announcements []*Announcement
	mu            sync.Mutex
	cond          *sync.Cond

	// announcers []AnnouncementWriter

	// children maps segment names to child nodes
	children map[string]*announcingNode
}

// setPattern sets the pattern and announcement configuration for this node.
// This is called when registering a new announcement writer at a leaf node.
// func (node *announcingNode) setPattern(pattern *pattern, config *AnnounceConfig) {
// 	if config == nil {
// 		panic("mux: nil announce config")
// 	}

// 	node.pattern = pattern
// 	// node.config = config
// 	// node.buffer = newAnnouncementsBuffer()
// }

// announce dispatches announcements to all registered announcement writers that match the given path.
// It traverses the announcement tree and delivers announcements to all matching nodes.
//
// Parameters:
//   - segments: The path segments to match against patterns in the tree
//   - handler: The handler that will serve announcements to matching writers
func (node *announcingNode) announce(segments []string, announcement *Announcement) {
	if len(segments) == 0 {
		// Serve announcements to the buffer
		slog.Debug("mux: serving announcements",
			"pattern", node.pattern.str,
		)
		// for _, w := range node.
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

	child.announce(segments[1:], announcement)
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

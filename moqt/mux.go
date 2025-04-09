package moqt

import (
	"log/slog"
	"strings"
	"sync"
	"time"
)

var DefaultMux *TrackMux = defaultMux

var defaultMux = NewTrackMux()

func NewTrackMux() *TrackMux {
	return &TrackMux{
		trackTree:    *newRoutingNode(),
		announceTree: *newAnnouncingNode(),
	}
}

func Handle(path TrackPath, handler TrackHandler) {
	DefaultMux.Handle(path, handler)
}

func ServeTrack(w TrackWriter, config *SubscribeConfig) {
	DefaultMux.ServeTrack(w, config)
}

func ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	DefaultMux.ServeAnnouncements(w, config)
}

func GetInfo(path TrackPath) (Info, error) {
	return DefaultMux.GetInfo(path)
}

func BuildTrack(path TrackPath, info Info, expires time.Duration) *TrackBuffer {
	return DefaultMux.BuildTrack(path, info, expires)
}

var _ TrackHandler = (*TrackMux)(nil)

type TrackMux struct {
	// http.ServeMux
	mu sync.RWMutex

	// trackTree is the root node of the routing tree for tracks.
	// It is used to find the handler for a given track path.
	trackTree routingNode

	// announceTree is the root node of the announcement tree.
	// It is used to announce new tracks to existing announcement writers.
	announceTree announcingNode
}

func (mux *TrackMux) Handle(path TrackPath, handler TrackHandler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	p := newPath(path)

	mux.registerHandler(p, handler)

	mux.announce(p, handler)

	slog.Debug("handling a track", "track_path", path)
}

func (mux *TrackMux) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
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
	if current.announcementsBuffer == nil {
		current.announcementsBuffer = newAnnouncementBuffer(config)
	}

	// Serve announcements
	current.announcementsBuffer.ServeAnnouncements(w, config)
}

func (mux *TrackMux) Handler(path TrackPath) TrackHandler {
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

func (mux *TrackMux) BuildTrack(path TrackPath, info Info, expires time.Duration) *TrackBuffer {
	buf := NewTrack(path, info, expires)

	mux.Handle(path, buf)

	return buf
}

func (mux *TrackMux) ServeTrack(w TrackWriter, config *SubscribeConfig) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPath(w.TrackPath())

	mux.trackTree.serveTrack(0, p, w, config)
}

func (mux *TrackMux) GetInfo(path TrackPath) (Info, error) {
	p := newPath(path)

	return mux.trackTree.getInfo(0, p)
}

func (mux *TrackMux) registerHandler(path *path, handler TrackHandler) {
	// Register the handler on the routing tree
	current := &mux.trackTree
	for _, seg := range path.segments {
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

	if current.handler != nil && current.pattern != nil {
		slog.Warn("trackmux: overwriting existing handler", "path", path.str)
	}

	// Set the handler on the leaf node
	current.setHandler(path, handler)

	slog.Debug("registered a handler", "track_path", path.str)
}

func (mux *TrackMux) announce(path *path, handler TrackHandler) {
	mux.announceTree.announce(0, path, handler)

	slog.Debug("announced a new track", "track_path", path.str)
}

func newRoutingNode() *routingNode {
	return &routingNode{
		children: make(map[string]*routingNode),
	}
}

type routingNode struct {
	// If this node is a leaf node, pattern and handler are set.
	// If this node is not a leaf node, pattern and handler are nil.
	pattern *path
	handler TrackHandler

	children map[string]*routingNode
}

func (node *routingNode) setHandler(ptn *path, handler TrackHandler) {
	if ptn == nil {
		panic("mux: nil pattern")
	} else if handler == nil {
		panic("mux: nil handler")
	}

	node.pattern = ptn
	node.handler = handler
}

func (node *routingNode) serveTrack(depth int, path *path, w TrackWriter, config *SubscribeConfig) {
	// If this node is a leaf node, call the handler.
	if depth >= path.depth() {
		if node.handler == nil {
			NotFoundHandler.ServeTrack(w, config)
			return
		}

		node.handler.ServeTrack(w, config)
		return
	}

	// If this node is not a leaf node, call the child node.
	segment := path.segments[depth]
	child, ok := node.children[segment]
	if !ok {
		NotFoundHandler.ServeTrack(w, config)
		return
	}

	child.serveTrack(depth+1, path, w, config)
}

func (node *routingNode) getInfo(depth int, path *path) (Info, error) {
	if depth >= path.depth() {
		if node.handler == nil {
			return NotFoundInfo, ErrTrackDoesNotExist
		}
		return node.handler.GetInfo(TrackPath(path.str))
	}

	// If this node is not a leaf node, call the child node.
	segment := path.segments[depth]
	child, ok := node.children[segment]
	if !ok {
		return NotFoundInfo, ErrTrackDoesNotExist
	}

	return child.getInfo(depth+1, path)
}

func newAnnouncingNode() *announcingNode {
	return &announcingNode{
		children: make(map[string]*announcingNode),
	}
}

// announcingNode is a node in the announcement tree.
type announcingNode struct {
	// If this node is a leaf node, pattern and handler are set.
	// If this node is not a leaf node, pattern and handler are nil.
	pattern             *pattern
	announcementsBuffer *announcementsBuffer

	children map[string]*announcingNode
}

func (node *announcingNode) announce(depth int, path *path, handler TrackHandler) {
	if node.pattern == nil {
		// Check for the next segment
		seg := path.segments[depth]
		child, ok := node.children[seg]
		if ok {
			child.announce(depth+1, path, handler)
		}

		// Check for single wildcard
		child, ok = node.children["*"]
		if ok {
			child.announce(depth+1, path, handler)
		}

		// Check for double wildcard
		child, ok = node.children["**"]
		if ok {
			for depth < path.depth() {
				child.announce(depth+1, path, handler)
				depth++
			}

			return
		}

		// If no matching child node is found, do nothing
		return
	}

	if node.announcementsBuffer != nil {
		handler.ServeAnnouncements(node.announcementsBuffer, node.announcementsBuffer.config)
	}
}

func (node *announcingNode) setPatternAndBuffer(p *pattern) {
	node.pattern = p
	config := &AnnounceConfig{
		TrackPattern: p.str,
	}
	node.announcementsBuffer = newAnnouncementBuffer(config)
}

func newPath(p TrackPath) *path {
	return (*path)(newPattern(string(p)))
}

func newPattern(str string) *pattern {
	segments := strings.Split(strings.Trim(str, "/"), "/")
	if len(segments) == 1 && segments[0] == "" {
		segments = []string{}
	}
	return &pattern{
		str:      str,
		segments: segments,
	}
}

type path pattern

func (p *path) depth() int {
	return len(p.segments)
}

type pattern struct {
	str      string
	segments []string
}

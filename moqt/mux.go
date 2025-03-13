package moqt

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

var DefaultMux *TrackMux = defaultMux

var defaultMux = NewTrackMux()

var _ TrackResolver = (*TrackMux)(nil)

func NewTrackMux() *TrackMux {
	return &TrackMux{
		tree: *newRoutingNode(nil, nil),
	}
}

type TrackMux struct {
	http.ServeMux
	mu   sync.RWMutex
	tree routingNode
}

func (mux *TrackMux) Handle(path TrackPath, handler TrackResolver) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	p := newPath(path.String())

	current := &mux.tree
	for _, seg := range p.segments {
		if current.children == nil {
			current.children = make(map[string]*routingNode)
		}

		if child, ok := current.children[seg]; ok {
			current = child
		} else {
			child := newRoutingNode(nil, nil)
			current.children[seg] = child
			current = child
		}
	}

	// Set the handler on the leaf node
	current.handler = handler
}

func (mux *TrackMux) Resolver(path TrackPath) TrackResolver {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPath(path.String())

	node := mux.findRoutingNode(p)
	if node == nil {
		return nil
	}

	return node.handler
}

func (mux *TrackMux) ServeTrack(w TrackWriter, config SubscribeConfig) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPath(w.TrackPath().String())
	node := mux.findRoutingNode(p)
	if node == nil || node.handler == nil {
		w.CloseWithError(ErrTrackDoesNotExist)
		return
	}

	node.handler.ServeTrack(w, config)
}

func (mux *TrackMux) ServeAnnouncements(w AnnouncementWriter) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPattern(w.AnnounceConfig().TrackPattern)

	//
	buf := newAnnouncementBuffer(w.AnnounceConfig())
	defer buf.Close()

	current := &mux.tree
	for _, seg := range p.segments {
		switch seg {
		case "*", "**":
			// Create wildcard node if it doesn't exist
			if current.wildcard == nil {
				child := newRoutingNode(p, nil)
				current.wildcard = child
			}
			current = current.wildcard
		default:
			// Create child node if it doesn't exist
			if _, ok := current.children[seg]; !ok {
				child := newRoutingNode(p, nil)
				current.children[seg] = child
			}

			current = current.children[seg]
		}

		if current.handler != nil {
			current.handler.ServeAnnouncements(buf)
		}
	}

	if current.announcers == nil {
		current.announcers = make([]AnnouncementWriter, 0)
	}

	current.announcers = append(current.announcers, w)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		anns, err := buf.ReceiveAnnouncements(ctx)
		if err != nil {
			w.CloseWithError(err)
			break
		}

		err = w.SendAnnouncements(anns)
		if err != nil {
			w.CloseWithError(err)
			break
		}
	}
}

func (mux *TrackMux) GetInfo(path TrackPath) (Info, error) {
	p := newPath(path.String())

	node := mux.findRoutingNode(p)

	if node == nil {
		return Info{}, ErrTrackDoesNotExist
	}

	return node.handler.GetInfo(path)
}

func (mux *TrackMux) findRoutingNode(ptn *path) *routingNode { // TODO
	var found *routingNode

	var search func(node *routingNode)
	search = func(node *routingNode) {
		if node == nil || found != nil {
			return
		}

		if node.pattern != nil && node.handler != nil {
			if matchGlob(node.pattern.str, ptn.str) {
				found = node
				return
			}
		}
		for _, child := range node.children {
			search(child)
			if found != nil {
				return
			}
		}
	}

	search(&mux.tree)

	return found
}

func newRoutingNode(ptn *pattern, handler TrackResolver) *routingNode {
	if handler == nil {
		handler = NotFoundHandler
	}

	return &routingNode{
		pattern: ptn,
		handler: handler,
		// announcer: make([]AnnouncementWriter, 0),
		children: make(map[string]*routingNode),
	}
}

type routingNode struct {
	pattern *pattern
	handler TrackResolver

	//
	announcers []AnnouncementWriter

	children map[string]*routingNode

	wildcard *routingNode
}

func newPath(str string) *path {
	return (*path)(newPattern(str))
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

type pattern struct {
	str      string
	segments []string
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func matchGlob(pattern, path string) bool {
	// If pattern or path is empty, only match if both are empty
	if pattern == "" {
		return path == ""
	}
	if path == "" {
		return pattern == ""
	}

	patternSegments := splitPath(pattern)
	pathSegments := splitPath(path)

	return matchSegments(patternSegments, pathSegments)
}

func matchSegments(patternSegments, pathSegments []string) bool {
	// Base cases for recursion
	if len(patternSegments) == 0 {
		return len(pathSegments) == 0
	}

	// Check for wildcard in the pattern
	if patternSegments[0] == "*" {
		// '*' matches exactly one segment
		if len(pathSegments) == 0 {
			return false
		}
		return matchSegments(patternSegments[1:], pathSegments[1:])
	} else if patternSegments[0] == "**" {
		// '**' can match zero or more segments
		if len(patternSegments) == 1 {
			return true // "**" at the end matches everything
		}

		// Try matching '**' with 0, 1, 2, ... segments
		for i := 0; i <= len(pathSegments); i++ {
			if matchSegments(patternSegments[1:], pathSegments[i:]) {
				return true
			}
		}
		return false
	}

	// Check if we have path segments left to match
	if len(pathSegments) == 0 {
		return false
	}

	// Check if segments match exactly (no wildcards in segment)
	if !strings.Contains(patternSegments[0], "*") && !strings.Contains(patternSegments[0], "?") {
		if patternSegments[0] == pathSegments[0] {
			return matchSegments(patternSegments[1:], pathSegments[1:])
		}
		return false
	}

	// Handle wildcard characters within segments (* and ?)
	if matchSegment(patternSegments[0], pathSegments[0]) {
		return matchSegments(patternSegments[1:], pathSegments[1:])
	}

	return false
}

func matchSegment(patternSegment, pathSegment string) bool {
	patternPos, pathPos := 0, 0
	starIdx, matchIdx := -1, 0

	for pathPos < len(pathSegment) {
		// Direct match or '?' wildcard.
		if patternPos < len(patternSegment) &&
			(patternSegment[patternPos] == '?' || patternSegment[patternPos] == pathSegment[pathPos]) {
			patternPos++
			pathPos++
		} else if patternPos < len(patternSegment) && patternSegment[patternPos] == '*' {
			// Record the position of '*' and the current string position.
			starIdx = patternPos
			matchIdx = pathPos
			patternPos++
		} else if starIdx != -1 {
			// Backtrack: try to match one more character with the '*' and adjust position.
			patternPos = starIdx + 1
			matchIdx++
			pathPos = matchIdx
		} else {
			return false
		}
	}

	// Consume any trailing '*' in the pattern.
	for patternPos < len(patternSegment) && patternSegment[patternPos] == '*' {
		patternPos++
	}

	return patternPos == len(patternSegment)
}

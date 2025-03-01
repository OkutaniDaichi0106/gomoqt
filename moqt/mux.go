package moqt

import (
	"strings"
	"sync"
)

var DefaultMux *ServeMux = defaultMux

var defaultMux = NewServeMux()

var _ Handler = (*ServeMux)(nil)

func NewServeMux() *ServeMux {
	return &ServeMux{}
}

type ServeMux struct {
	mu    sync.RWMutex
	tree  routingNode
	index routingIndex
}

func (mux *ServeMux) Handle(path string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	// Initialize children map on first registration
	if mux.tree.children == nil {
		mux.tree.children = make(map[string]*routingNode)
	}

	p := newPattern(path)
	mux.index.add(p)

	node := &mux.tree
	for _, seg := range p.segments {
		if child, ok := node.children[seg]; ok {
			node = child
		} else {
			child := newRoutingNode(nil, nil)
			node.children[seg] = child
			node = child
		}

		// Serve announcement on the node
		for _, announcer := range node.announcer {
			handler.ServeAnnouncement(announcer.AnnouncementWriter, announcer.AnnounceConfig)
		}
	}

	// Set the handler on the leaf node
	node.handler = handler
}

func (mux *ServeMux) ServeTrack(w TrackWriter, r SubscribeConfig) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPattern(string(r.TrackPath))
	handler := mux.findRoutingNode(p).handler
	if handler == nil {
		w.CloseWithError(ErrTrackDoesNotExist)
		return
	}

	handler.ServeTrack(w, r)
}

func (mux *ServeMux) ServeAnnouncement(w AnnouncementWriter, r AnnounceConfig) {
	// Example implementation: fetch path using r.GetPath() and call the handler from routingNode
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPattern(string(r.TrackPrefix))

	// node := mux.findRoutingNode(p)
	// if node == nil {
	// 	w.CloseWithError(ErrTrackDoesNotExist)
	// 	return
	// }

	node := &mux.tree
	for _, seg := range p.segments {
		if child, ok := node.children[seg]; ok {
			node = child
		} else {
			child := newRoutingNode(nil, nil)
			node.children[seg] = child
			node = child
		}
	}

	node.announcer = append(node.announcer, struct {
		AnnouncementWriter
		AnnounceConfig
	}{w, r})
}

func (mux *ServeMux) ServeInfo(ch chan<- Info, r InfoRequest) {
	// Example implementation: fetch path using GetPath() from InfoRequest and query handler from routingNode
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	p := newPattern(string(r.TrackPath))
	handler := mux.findRoutingNode(p).handler
	if handler == nil {
		close(ch)
		return
	}

	handler.ServeInfo(ch, r)
}

func (mux *ServeMux) findRoutingNode(ptn *pattern) *routingNode {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

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

func newRoutingNode(ptn *pattern, handler Handler) *routingNode {
	return &routingNode{
		pattern:  ptn,
		handler:  handler,
		children: make(map[string]*routingNode),
	}
}

type routingNode struct {
	pattern *pattern
	handler Handler

	announcer []struct {
		AnnouncementWriter
		AnnounceConfig
	}

	children map[string]*routingNode
}

func newRoutingIndex() routingIndex {
	return routingIndex{
		segments: make(map[routingIndexKey][]*pattern),
	}
}

type routingIndex struct {
	segments map[routingIndexKey][]*pattern
}

type routingIndexKey struct {
	pos     int
	segment string
}

func (idx *routingIndex) add(p *pattern) {
	for i, seg := range p.segments {
		key := routingIndexKey{pos: i, segment: seg}
		idx.segments[key] = append(idx.segments[key], p)
	}
}

func (idx *routingIndex) find(pos int, seg string) []*pattern {
	key := routingIndexKey{pos: pos, segment: seg}
	if pats, ok := idx.segments[key]; ok {
		return pats
	}
	return nil
}

func (idx *routingIndex) remove(p *pattern) { // TODO: review
	for i, seg := range p.segments {
		key := routingIndexKey{pos: i, segment: seg}
		for j, pat := range idx.segments[key] {
			if pat == p {
				idx.segments[key] = append(idx.segments[key][:j], idx.segments[key][j+1:]...)
				break
			}
		}
	}
}

// func (idx *routingIndex) update(old, new *pattern) {
// 	idx.remove(old)
// 	idx.add(new)
// } // TODO: Add if needed

func (idx *routingIndex) clear() {
	idx.segments = make(map[routingIndexKey][]*pattern)
}

func newPattern(str string) *pattern {
	p := &pattern{str: str}
	p.segments = strings.Split(str, "/")
	return p
}

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
	patternSegments := splitPath(pattern)
	pathSegments := splitPath(path)

	return matchSegments(patternSegments, pathSegments)
}

func matchSegments(patternSegments, pathSegments []string) bool {
	if len(patternSegments) == 0 {
		return len(pathSegments) == 0
	}

	// Check if the first segment is a '**' wildcard
	if pathSegments[0] == "**" {
		if len(patternSegments) == 1 {
			return true
		}

		for pos := 0; pos < len(patternSegments); pos++ {
			if matchSegments(patternSegments[pos:], pathSegments[1:]) {
				return true
			}
		}

		return false
	}

	// Check if the path segment is left
	if len(pathSegments) == 0 {
		return false
	}

	// Check if the first segment matches
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
		if patternPos < len(patternSegment) && (patternSegment[pathPos] == '?' || patternSegment[patternPos] == pathSegment[pathPos]) {
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

package moqt

import (
	"sync"
)

// TODO: Implement
type TrackManager struct {
	trackPathTree *trackNode
	mu            sync.Mutex
}

func NewTrackManager() *TrackManager {
	return &TrackManager{
		trackPathTree: &trackNode{children: make(map[string]*trackNode)},
	}
}

type trackNode struct {
	part string

	children map[string]*trackNode

	buffer TrackBuffer
}

func (tr *TrackManager) Publish(src TrackReader) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	node := tr.trackPathTree

	for _, part := range src.TrackPath().Parts() {
		if _, ok := node.children[part]; !ok {
			node.children[part] = &trackNode{part: part, children: make(map[string]*trackNode)}
		} else {
			return ErrDuplicatedTrack
		}
		node = node.children[part]
	}

	return nil
}

func (tr *TrackManager) Subscribe(dst TrackWriter) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	node := tr.trackPathTree

	for _, part := range dst.TrackPath().Parts() {
		if _, ok := node.children[part]; !ok {
			node.children[part] = &trackNode{part: part, children: make(map[string]*trackNode)}
		}
		node = node.children[part]
	}

	return nil
}

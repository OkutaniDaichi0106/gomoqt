package moqt

import (
	"context"
	"testing"
	"time"
)

// Test internal behavior of TrackMux - node management and cleanup
func TestTrackMux_InternalNodeManagement(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	// Test that nodes are created properly
	handler := &MockTrackHandler{}
	path := BroadcastPath("/internal/test")

	mux.Handle(ctx, path, handler)

	// Access internal node structure
	mux.mu.RLock()
	root := mux.root
	mux.mu.RUnlock()

	if root == nil {
		t.Error("Expected root node to be created")
	}

	// Verify node hierarchy
	if root.children == nil {
		t.Error("Expected root node to have children map")
	}

	if _, exists := root.children["internal"]; !exists {
		t.Error("Expected 'internal' node to exist")
	}
}

func TestTrackMux_InternalContextCancellation(t *testing.T) {
	mux := NewTrackMux()

	// Test context cancellation cleanup
	ctx, cancel := context.WithCancel(context.Background())

	handler := &MockTrackHandler{}
	path := BroadcastPath("/cancellable/resource")

	mux.Handle(ctx, path, handler)

	// Verify handler is registered
	mux.mu.RLock()
	nodeCount := countNodes(mux.root)
	mux.mu.RUnlock()

	if nodeCount < 2 { // root + at least one child
		t.Errorf("Expected at least 2 nodes, got %d", nodeCount)
	}

	// Cancel context
	cancel()

	// Give time for cleanup
	time.Sleep(50 * time.Millisecond)

	// Verify cleanup occurred (nodes should be removed)
	mux.mu.RLock()
	newNodeCount := countNodes(mux.root)
	mux.mu.RUnlock()

	if newNodeCount >= nodeCount {
		t.Error("Expected node cleanup after context cancellation")
	}
}

func TestTrackMux_InternalAnnouncementPropagation(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	// Test that announcements are properly propagated to subscribers
	announcementReceived := false
	announcer := &MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*Announcement) error {
			announcementReceived = true
			return nil
		},
	}

	config := &AnnounceConfig{TrackPrefix: "/**"}
	go mux.ServeAnnouncements(announcer, config)

	// Give announcer time to register
	time.Sleep(10 * time.Millisecond)

	// Register a handler (should trigger announcement)
	handler := &MockTrackHandler{}
	mux.Handle(ctx, "/test/track", handler)

	// Give time for announcement processing
	time.Sleep(50 * time.Millisecond)

	if !announcementReceived {
		t.Error("Expected announcement to be sent to announcer")
	}
}

func TestTrackMux_InternalWildcardMatching(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	// Test wildcard pattern matching in announcements
	audioAnnouncements := 0
	allAnnouncements := 0

	audioAnnouncer := &MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*Announcement) error {
			audioAnnouncements += len(announcements)
			return nil
		},
	}

	allAnnouncer := &MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*Announcement) error {
			allAnnouncements += len(announcements)
			return nil
		},
	}

	// Register announcers with different patterns
	go mux.ServeAnnouncements(audioAnnouncer, &AnnounceConfig{TrackPrefix: "/audio/*"})
	go mux.ServeAnnouncements(allAnnouncer, &AnnounceConfig{TrackPrefix: "/**"})

	// Give announcers time to register
	time.Sleep(10 * time.Millisecond)

	// Register handlers for different paths
	handler1 := &MockTrackHandler{}
	handler2 := &MockTrackHandler{}
	handler3 := &MockTrackHandler{}

	mux.Handle(ctx, "/audio/stream1", handler1)
	mux.Handle(ctx, "/video/stream1", handler2)
	mux.Handle(ctx, "/audio/stream2", handler3)

	// Give time for announcement processing
	time.Sleep(50 * time.Millisecond)

	if audioAnnouncements != 2 {
		t.Errorf("Expected 2 audio announcements, got %d", audioAnnouncements)
	}

	if allAnnouncements != 3 {
		t.Errorf("Expected 3 total announcements, got %d", allAnnouncements)
	}
}

// Helper function to count nodes in the tree
func countNodes(node *muxNode) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.children {
		count += countNodes(child)
	}

	return count
}

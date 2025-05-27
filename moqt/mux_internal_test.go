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

	mux.Handle(ctx, "/internal/test", TrackHandlerFunc(func(p *Publisher) {}))
	// Access internal node structure
	mux.mu.RLock()
	root := &mux.trackTree
	mux.mu.RUnlock()

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

	mux.Handle(ctx, "/cancellable/resource", TrackHandlerFunc(func(p *Publisher) {}))
	// Verify handler is registered
	mux.mu.RLock()
	nodeCount := countNodes(&mux.trackTree)
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
	newNodeCount := countNodes(&mux.trackTree)
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
	config := "/**"
	go mux.ServeAnnouncements(announcer, config)

	// Give announcer time to register
	time.Sleep(10 * time.Millisecond)

	// Register a handler (should trigger announcement)
	mux.Handle(ctx, "/test/track", TrackHandlerFunc(func(p *Publisher) {}))

	// Give time for announcement processing
	time.Sleep(50 * time.Millisecond)

	if !announcementReceived {
		t.Error("Expected announcement to be sent to announcer")
	}
}

// Helper function to count nodes in the tree
func countNodes(node *routingNode) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.children {
		count += countNodes(child)
	}

	return count
}

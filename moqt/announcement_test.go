package moqt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAnnouncement(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	if announcement == nil {
		t.Fatal("Expected non-nil announcement")
	}

	if announcement.path != path {
		t.Errorf("Expected path %s, got %s", path, announcement.path)
	}

	if announcement.ctx == nil {
		t.Error("Expected non-nil context")
	}

	if announcement.cancel == nil {
		t.Error("Expected non-nil cancel function")
	}
}

func TestAnnouncement_TrackPath(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	if got := announcement.BroadcastPath(); got != path {
		t.Errorf("TrackPath() = %v, want %v", got, path)
	}
}

func TestAnnouncement_String(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	expected := "{ AnnounceStatus: ACTIVE, BroadcastPath: test/path }"
	if got := announcement.String(); got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}

	// Test with ended announcement
	announcement.End()
	expected = "{ AnnounceStatus: ENDED, BroadcastPath: test/path }"
	if got := announcement.String(); got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}
}

func TestAnnouncement_IsActive(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	if !announcement.IsActive() {
		t.Error("Expected announcement to be active")
	}

	announcement.End()

	if announcement.IsActive() {
		t.Error("Expected announcement to be inactive after End()")
	}
}

func TestAnnouncement_End(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	if !announcement.IsActive() {
		t.Error("Expected announcement to be active initially")
	}

	announcement.End()

	if announcement.IsActive() {
		t.Error("Expected announcement to be inactive after End()")
	}
}

func TestAnnouncement_AwaitEnd(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	// Test that AwaitEnd returns a channel that is not closed initially
	select {
	case <-announcement.AwaitEnd():
		t.Error("Expected AwaitEnd() channel to not be closed initially")
	default:
		// This is the expected behavior
	}

	// Test that AwaitEnd returns a channel that is closed after End() is called
	go func() {
		time.Sleep(100 * time.Millisecond)
		announcement.End()
	}()

	select {
	case <-announcement.AwaitEnd():
		// This is the expected behavior
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected AwaitEnd() channel to be closed after End()")
	}
}

func TestAnnouncement_WithCancelledContext(t *testing.T) {
	// Test with a context that gets cancelled externally
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	announcement := NewAnnouncement(ctx, "test/path")

	if !announcement.IsActive() {
		t.Error("Expected announcement to be active initially")
	}

	// Cancel the parent context
	cancel()

	// Give a little time for propagation
	time.Sleep(10 * time.Millisecond)

	// The announcement should become inactive since the parent context was cancelled
	// and context.WithCancel creates a child context that inherits cancellation
	if announcement.IsActive() {
		t.Error("Expected announcement to become inactive after parent context cancellation")
	}
}

func TestAnnouncement_Fork(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	original := NewAnnouncement(ctx, path)

	// Fork the announcement
	forked := original.Fork()

	assert.Equal(t, original.path, forked.path)

	// Verify both are initially active
	assert.True(t, original.IsActive(), "Expected original announcement to be active")
	assert.True(t, forked.IsActive(), "Expected forked announcement to be active")

	// End the original announcement
	original.End()

	// Both original and forked should be inactive, but forked should still be active
	assert.False(t, original.IsActive(), "Expected original announcement to be inactive after End()")
	assert.False(t, forked.IsActive(), "Expected forked announcement to be inactive after End()")
}

func TestAnnouncement_ForkWithCancelledParent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	path := BroadcastPath("test/path")

	original := NewAnnouncement(ctx, path)

	// Cancel the original context
	cancel()

	// Give time for cancellation to propagate
	time.Sleep(10 * time.Millisecond)

	// Original should become inactive
	if original.IsActive() {
		t.Error("Expected original announcement to be inactive after context cancellation")
	}

	// Fork from the original (which has a cancelled context)
	forked := original.Fork()

	// Forked announcement should also be inactive since it inherits the cancelled context
	if forked.IsActive() {
		t.Error("Expected forked announcement to be inactive when forked from cancelled context")
	}
}

func TestAnnouncement_MultipleEnd(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	if !announcement.IsActive() {
		t.Error("Expected announcement to be active initially")
	}

	// Call End() multiple times
	announcement.End()
	announcement.End()
	announcement.End()

	if announcement.IsActive() {
		t.Error("Expected announcement to be inactive after multiple End() calls")
	}
}

func TestAnnouncement_ConcurrentAwaitEnd(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	// Start multiple goroutines waiting for end
	const numGoroutines = 10
	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			select {
			case <-announcement.AwaitEnd():
				results <- true
			case <-time.After(500 * time.Millisecond):
				results <- false
			}
		}()
	}

	// Give goroutines time to start waiting
	time.Sleep(50 * time.Millisecond)

	// End the announcement
	announcement.End()

	// All goroutines should receive the signal
	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			if !result {
				t.Error("Expected all goroutines to receive end signal")
			}
		case <-time.After(200 * time.Millisecond):
			t.Error("Timeout waiting for goroutine to complete")
		}
	}
}

func TestAnnouncement_WithAlreadyCancelledContext(t *testing.T) {
	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path := BroadcastPath("test/path")
	announcement := NewAnnouncement(ctx, path)

	// Announcement should be inactive since context is already cancelled
	if announcement.IsActive() {
		t.Error("Expected announcement to be inactive when created with already-cancelled context")
	}

	// AwaitEnd should be immediately available
	select {
	case <-announcement.AwaitEnd():
		// This is expected
	case <-time.After(10 * time.Millisecond):
		t.Error("Expected AwaitEnd() to be immediately available for cancelled context")
	}
}

func TestAnnouncement_EmptyPath(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("")

	announcement := NewAnnouncement(ctx, path)

	if announcement == nil {
		t.Fatal("Expected non-nil announcement even with empty path")
	}

	if announcement.BroadcastPath() != path {
		t.Errorf("Expected path %s, got %s", path, announcement.BroadcastPath())
	}

	expected := "Announcement: { AnnounceStatus: ACTIVE, BroadcastPath:  }"
	if got := announcement.String(); got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}
}

func TestAnnouncement_StringWithSpecialCharacters(t *testing.T) {
	ctx := context.Background()
	path := BroadcastPath("test/path with spaces/and-dashes_and.dots")

	announcement := NewAnnouncement(ctx, path)

	expected := "Announcement: { AnnounceStatus: ACTIVE, BroadcastPath: test/path with spaces/and-dashes_and.dots }"
	if got := announcement.String(); got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}
}

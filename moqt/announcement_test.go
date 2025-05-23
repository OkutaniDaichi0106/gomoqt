package moqt

import (
	"context"
	"testing"
	"time"
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

	expected := "Announcement: { AnnounceStatus: ACTIVE, TrackPath: test/path }"
	if got := announcement.String(); got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}

	// Test with ended announcement
	announcement.End()
	expected = "Announcement: { AnnounceStatus: ENDED, TrackPath: test/path }"
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
	path := BroadcastPath("test/path")

	announcement := NewAnnouncement(ctx, path)

	if !announcement.IsActive() {
		t.Error("Expected announcement to be active initially")
	}

	// Cancel the parent context
	cancel()

	// Give a little time for propagation
	time.Sleep(10 * time.Millisecond)

	// The announcement should still be active since we used WithCancel
	// which creates a new derived context
	if !announcement.IsActive() {
		t.Error("Expected announcement to remain active after parent context cancellation")
	}

	// Now end the announcement explicitly
	announcement.End()
	if announcement.IsActive() {
		t.Error("Expected announcement to be inactive after End()")
	}
}

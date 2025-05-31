package moqt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
)

func TestNewSendAnnounceStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test/prefix"

	sas := newSendAnnounceStream(mockStream, prefix, moqtrace.DefaultQUICStreamAccepted(0))

	if sas == nil {
		t.Fatal("newSendAnnounceStream returned nil")
	}

	if sas.prefix != prefix {
		t.Errorf("prefix = %v, want %v", sas.prefix, prefix)
	}

	if sas.stream != mockStream {
		t.Error("stream not set correctly")
	}

	if sas.actives == nil {
		t.Error("actives map should not be nil")
	}

	if sas.pendings == nil {
		t.Error("pendings map should not be nil")
	}

	if sas.sendCh == nil {
		t.Error("sendCh should not be nil")
	}

	// Give time for goroutine to start
	time.Sleep(10 * time.Millisecond)
}

func TestSendAnnounceStreamSendAnnouncements(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test"

	sas := newSendAnnounceStream(mockStream, prefix, moqtrace.DefaultQUICStreamAccepted(0))

	// Create test announcements
	ctx := context.Background()
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

	announcements := []*Announcement{ann1, ann2}

	err := sas.SendAnnouncements(announcements)
	if err != nil {
		t.Errorf("SendAnnouncements() error = %v", err)
	}

	// Verify announcements are stored
	if len(sas.actives) != 2 {
		t.Errorf("actives length = %v, want 2", len(sas.actives))
	}

	// Check that streams with wrong prefix are ignored
	ann3 := NewAnnouncement(ctx, BroadcastPath("/other/stream"))
	err = sas.SendAnnouncements([]*Announcement{ann3})
	if err != nil {
		t.Errorf("SendAnnouncements() with wrong prefix error = %v", err)
	}

	// Should still have only 2 active announcements
	if len(sas.actives) != 2 {
		t.Errorf("actives length after wrong prefix = %v, want 2", len(sas.actives))
	}
}

func TestSendAnnounceStreamSendAnnouncementsClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test"

	sas := newSendAnnounceStream(mockStream, prefix, moqtrace.DefaultQUICStreamAccepted(0))
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream"))

	err := sas.SendAnnouncements([]*Announcement{ann})
	if err == nil {
		t.Error("SendAnnouncements() on closed stream should return error")
	}
}

func TestSendAnnounceStreamSet(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	// Test setting active announcement
	err := sas.set("stream1", true)
	if err != nil {
		t.Errorf("set(active) error = %v", err)
	}

	if len(sas.pendings) != 1 {
		t.Errorf("pendings length = %v, want 1", len(sas.pendings))
	}

	pending := sas.pendings["stream1"]
	if pending.AnnounceStatus != message.ACTIVE {
		t.Errorf("announce status = %v, want %v", pending.AnnounceStatus, message.ACTIVE)
	}

	// Test setting ended announcement
	err = sas.set("stream1", false)
	if err != nil {
		t.Errorf("set(ended) error = %v", err)
	}

	pending = sas.pendings["stream1"]
	if pending.AnnounceStatus != message.ENDED {
		t.Errorf("announce status = %v, want %v", pending.AnnounceStatus, message.ENDED)
	}
}

func TestSendAnnounceStreamSetClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	err := sas.set("stream1", true)
	if err == nil {
		t.Error("set() on closed stream should return error")
	}
}

func TestSendAnnounceStreamSend(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	// Add some pending messages
	sas.pendings["stream1"] = message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    "stream1",
	}

	err := sas.send()
	if err != nil {
		t.Errorf("send() error = %v", err)
	}

	// Pendings should be cleared after sending
	if len(sas.pendings) != 0 {
		t.Errorf("pendings length after send = %v, want 0", len(sas.pendings))
	}

}

func TestSendAnnounceStreamSendNoPendings(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	// No pending messages
	err := sas.send()
	if err != nil {
		t.Errorf("send() with no pendings error = %v", err)
	}

}

func TestSendAnnounceStreamSendClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	err := sas.send()
	if err == nil {
		t.Error("send() on closed stream should return error")
	}
}

func TestSendAnnounceStreamClose(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	err := sas.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !sas.closed {
		t.Error("stream should be marked as closed")
	}
}

func TestSendAnnounceStreamCloseAlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))
	sas.closed = true

	err := sas.Close()
	if err != nil {
		t.Errorf("Close() on already closed stream error = %v", err)
	}
}

func TestSendAnnounceStreamCloseWithError(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	testErr := errors.New("test error")
	err := sas.CloseWithError(testErr)
	if err != nil {
		t.Errorf("CloseWithError() error = %v", err)
	}

	if !sas.closed {
		t.Error("stream should be marked as closed")
	}

	if sas.closeErr != testErr {
		t.Errorf("closeErr = %v, want %v", sas.closeErr, testErr)
	}
}

func TestSendAnnounceStreamCloseWithErrorAlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))
	sas.closed = true
	existingErr := errors.New("existing error")
	sas.closeErr = existingErr

	newErr := errors.New("new error")
	err := sas.CloseWithError(newErr)
	if err != existingErr {
		t.Errorf("CloseWithError() on already closed stream error = %v, want %v", err, existingErr)
	}
}

func TestSendAnnounceStreamInterface(t *testing.T) {
	// Test that sendAnnounceStream implements AnnouncementWriter interface
	var _ AnnouncementWriter = (*sendAnnounceStream)(nil)
}

func TestSendAnnounceStreamConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	// Test concurrent access to set and send
	go func() {
		for i := 0; i < 10; i++ {
			sas.set("stream1", true)
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			sas.send()
			time.Sleep(time.Millisecond)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	// Test should complete without race conditions
}

func TestSendAnnounceStreamAnnouncementLifecycle(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newSendAnnounceStream(mockStream, "/test", moqtrace.DefaultQUICStreamAccepted(0))

	// Create announcement
	ctx, cancel := context.WithCancel(context.Background())
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream"))

	// Send announcement
	err := sas.SendAnnouncements([]*Announcement{ann})
	if err != nil {
		t.Errorf("SendAnnouncements() error = %v", err)
	}

	// Verify active announcement is stored
	if len(sas.actives) != 1 {
		t.Errorf("actives length = %v, want 1", len(sas.actives))
	}

	// End the announcement
	cancel()

	// Give time for announcement end to be processed
	time.Sleep(20 * time.Millisecond)
}

func TestSendAnnounceStreamPrefixMatching(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		broadcastPath string
		shouldMatch   bool
	}{
		{
			name:          "exact prefix match",
			prefix:        "/test",
			broadcastPath: "/test/stream",
			shouldMatch:   true,
		},
		{
			name:          "nested prefix match",
			prefix:        "/test/sub",
			broadcastPath: "/test/sub/stream",
			shouldMatch:   true,
		},
		{
			name:          "no match - different prefix",
			prefix:        "/test",
			broadcastPath: "/other/stream",
			shouldMatch:   false,
		},
		{
			name:          "no match - prefix is substring but not path prefix",
			prefix:        "/test",
			broadcastPath: "/testing/stream",
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &MockQUICStream{}

			sas := newSendAnnounceStream(mockStream, tt.prefix, moqtrace.DefaultQUICStreamAccepted(0))

			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			initialActiveCount := len(sas.actives)

			err := sas.SendAnnouncements([]*Announcement{ann})
			if err != nil {
				t.Errorf("SendAnnouncements() error = %v", err)
			}

			expectedActiveCount := initialActiveCount
			if tt.shouldMatch {
				expectedActiveCount++
			}

			if len(sas.actives) != expectedActiveCount {
				t.Errorf("actives length = %v, want %v", len(sas.actives), expectedActiveCount)
			}
		})
	}
}

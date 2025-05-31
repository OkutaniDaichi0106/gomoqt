package moqt

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewReceiveAnnounceStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test/prefix"

	ras := newReceiveAnnounceStream(mockStream, prefix)

	if ras == nil {
		t.Fatal("newReceiveAnnounceStream returned nil")
	}

	if ras.prefix != prefix {
		t.Errorf("prefix = %v, want %v", ras.prefix, prefix)
	}

	if ras.stream != mockStream {
		t.Error("stream not set correctly")
	}

	if ras.announcements == nil {
		t.Error("announcements map should not be nil")
	}

	if ras.next == nil {
		t.Error("next slice should not be nil")
	}

	if ras.liveCh == nil {
		t.Error("liveCh should not be nil")
	}

	// Give time for goroutine to start
	time.Sleep(10 * time.Millisecond)
}

func TestReceiveAnnounceStreamReceiveAnnouncementsEmpty(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	announcements, err := ras.ReceiveAnnouncements(ctx)

	// Should timeout since no announcements are available
	if err == nil {
		t.Error("expected timeout error")
	}

	if announcements != nil {
		t.Errorf("announcements should be nil on timeout, got %v", announcements)
	}
}

func TestReceiveAnnounceStreamReceiveAnnouncementsClosed(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Close the stream
	testErr := errors.New("stream closed")
	ras.closed = true
	ras.closeErr = testErr

	ctx := context.Background()
	announcements, err := ras.ReceiveAnnouncements(ctx)

	if err != testErr {
		t.Errorf("ReceiveAnnouncements() error = %v, want %v", err, testErr)
	}

	if announcements != nil {
		t.Errorf("announcements should be nil on closed stream, got %v", announcements)
	}
}

func TestReceiveAnnounceStreamClose(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	err := ras.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !ras.closed {
		t.Error("stream should be marked as closed")
	}
}

func TestReceiveAnnounceStream_CloseAlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")
	ras.closed = true

	err := ras.Close()
	if err == nil {
		t.Error("Close() on already closed stream should return error")
	}
}

func TestReceiveAnnounceStream_CloseWithError(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	testErr := ErrProtocolViolation
	err := ras.CloseWithError(testErr)
	if err != nil {
		t.Errorf("CloseWithError() error = %v", err)
	}

	if !ras.closed {
		t.Error("stream should be marked as closed")
	}

	if ras.closeErr != testErr {
		t.Errorf("closeErr = %v, want %v", ras.closeErr, testErr)
	}

	// Should cancel read and write on underlying stream
	if !mockStream.AssertCalled(t, "CancelRead") {
		t.Error("underlying stream should be cancelled")
	}
}

func TestReceiveAnnounceStream_CloseWithNilError(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	err := ras.CloseWithError(nil)
	if err != nil {
		t.Errorf("CloseWithError(nil) error = %v", err)
	}

	if !ras.closed {
		t.Error("stream should be marked as closed")
	}

	// Should use default error when nil is passed
	if ras.closeErr != ErrInternalError {
		t.Errorf("closeErr = %v, want %v", ras.closeErr, ErrInternalError)
	}
}

func TestReceiveAnnounceStreamCloseWithErrorAlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")
	ras.closed = true
	existingErr := errors.New("existing error")
	ras.closeErr = existingErr

	newErr := errors.New("new error")
	err := ras.CloseWithError(newErr)
	if err != existingErr {
		t.Errorf("CloseWithError() on already closed stream error = %v, want %v", err, existingErr)
	}
}

func TestReceiveAnnounceStreamInterface(t *testing.T) {
	// Test that receiveAnnounceStream implements AnnouncementReader interface
	var _ AnnouncementReader = (*receiveAnnounceStream)(nil)
}

func TestReceiveAnnounceStreamAnnouncementTracking(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Test internal announcement tracking
	ctx := context.Background()
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

	// Manually add announcements to test tracking
	ras.announcements["stream1"] = ann1
	ras.announcements["stream2"] = ann2

	if len(ras.announcements) != 2 {
		t.Errorf("announcements length = %v, want 2", len(ras.announcements))
	}

	// Test ending announcement
	ann1.End()

	// Announcement should still be in map until processed by listenAnnouncements
	if len(ras.announcements) != 2 {
		t.Errorf("announcements length after end = %v, want 2", len(ras.announcements))
	}
}

func TestReceiveAnnounceStreamConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Test concurrent access to ReceiveAnnouncements and Close
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		ras.ReceiveAnnouncements(ctx)
	}()

	go func() {
		time.Sleep(25 * time.Millisecond)
		ras.Close()
	}()

	time.Sleep(100 * time.Millisecond)

	// Test should complete without race conditions
}

func TestReceiveAnnounceStreamContextCancellation(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	announcements, err := ras.ReceiveAnnouncements(ctx)

	if err != context.Canceled {
		t.Errorf("ReceiveAnnouncements() error = %v, want %v", err, context.Canceled)
	}

	if announcements != nil {
		t.Errorf("announcements should be nil on context cancellation, got %v", announcements)
	}
}

func TestReceiveAnnounceStreamPrefixHandling(t *testing.T) {
	tests := []struct {
		name         string
		prefix       string
		suffix       string
		expectedPath string
	}{
		{
			name:         "simple prefix and suffix",
			prefix:       "/test",
			suffix:       "/stream",
			expectedPath: "/test/stream",
		},
		{
			name:         "nested prefix",
			prefix:       "/test/sub",
			suffix:       "/stream",
			expectedPath: "/test/sub/stream",
		},
		{
			name:         "root prefix",
			prefix:       "/",
			suffix:       "stream",
			expectedPath: "/stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &MockQUICStream{}

			ras := newReceiveAnnounceStream(mockStream, tt.prefix)

			// Manually simulate announcement creation
			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.expectedPath))
			ras.next = append(ras.next, ann)

			announcements, err := ras.ReceiveAnnouncements(ctx)
			if err != nil {
				t.Errorf("ReceiveAnnouncements() error = %v", err)
			}

			if len(announcements) != 1 {
				t.Errorf("announcements length = %v, want 1", len(announcements))
			}

			if string(announcements[0].BroadcastPath()) != tt.expectedPath {
				t.Errorf("announcement path = %v, want %v", announcements[0].BroadcastPath(), tt.expectedPath)
			}
		})
	}
}

func TestReceiveAnnounceStreamListenAnnouncementsInvalidMessage(t *testing.T) {
	// Create invalid message data

	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer([]byte{0xFF, 0xFF, 0xFF}),
	}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Give time for listenAnnouncements to process invalid data
	time.Sleep(50 * time.Millisecond)

	// Stream should be closed due to invalid message
	if !ras.closed {
		t.Error("stream should be closed after invalid message")
	}
}

func TestReceiveAnnounceStreamDoubleClose(t *testing.T) {
	mockStream := &MockQUICStream{}

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// First close
	err1 := ras.Close()
	if err1 != nil {
		t.Errorf("first Close() error = %v", err1)
	}

	// Second close should return error
	err2 := ras.Close()
	if err2 == nil {
		t.Error("second Close() should return error")
	}
}

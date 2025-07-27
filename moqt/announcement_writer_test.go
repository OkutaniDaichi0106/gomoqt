package moqt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper functions for testing

// createMockStreamWithBehavior creates a mock stream with specific behavior patterns
func createMockStreamWithBehavior(behavior string) *MockQUICStream {
	mockStream := &MockQUICStream{}

	switch behavior {
	case "success":
		mockStream.On("Write", mock.Anything).Return(0, nil)
	case "write_error":
		writeError := errors.New("write failed")
		mockStream.On("Write", mock.Anything).Return(0, writeError)
	case "stream_error":
		streamError := &quic.StreamError{
			StreamID:  quic.StreamID(123),
			ErrorCode: quic.StreamErrorCode(42),
		}
		mockStream.On("Write", mock.Anything).Return(0, streamError)
	case "intermittent":
		// First call succeeds, second fails
		mockStream.On("Write", mock.Anything).Return(0, nil).Once()
		mockStream.On("Write", mock.Anything).Return(0, errors.New("second write failed")).Once()
	case "close_error":
		mockStream.On("Close").Return(errors.New("close failed"))
	case "close_success":
		mockStream.On("Close").Return(nil)
	default:
		// Default successful behavior
		mockStream.On("Write", mock.Anything).Return(0, nil)
	}

	return mockStream
}

// createTestAnnouncement creates a test announcement with specified parameters
func createTestAnnouncement(ctx context.Context, path BroadcastPath) *Announcement {
	if ctx == nil {
		ctx = context.Background()
	}
	return NewAnnouncement(ctx, path)
}

// verifyAnnouncementState verifies the expected state of announcements in SendAnnounceStream
func verifyAnnouncementState(t *testing.T, sas *AnnouncementWriter, expectedCount int, expectedPaths []string) {
	t.Helper()

	assert.Len(t, sas.actives, expectedCount, "Expected %d active announcements, got %d", expectedCount, len(sas.actives))

	for _, path := range expectedPaths {
		assert.Contains(t, sas.actives, path, "Expected path %s to be in actives", path)
	}

	// Verify no unexpected paths
	for path := range sas.actives {
		found := false
		for _, expected := range expectedPaths {
			if path == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Unexpected path %s found in actives", path)
	}
}

// Helper function to wait for background goroutine processing
func waitForBackgroundProcessing() {
	time.Sleep(150 * time.Millisecond)
}

// createMultipleAnnouncements creates multiple test announcements with different paths
func createMultipleAnnouncements(ctx context.Context, prefix string, count int) []*Announcement {
	announcements := make([]*Announcement, count)
	for i := 0; i < count; i++ {
		path := fmt.Sprintf("%s/stream%d", prefix, i+1)
		announcements[i] = createTestAnnouncement(ctx, BroadcastPath(path))
	}
	return announcements
}

// assertStreamState verifies the state of a SendAnnounceStream
func assertStreamState(t *testing.T, sas *AnnouncementWriter, expectClosed bool, expectError bool) {
	t.Helper()

	assert.Equal(t, expectClosed, sas.closed, "Stream closed state mismatch")

	if expectError {
		assert.NotNil(t, sas.closeErr, "Expected close error but got nil")
	} else {
		assert.Nil(t, sas.closeErr, "Expected no close error but got: %v", sas.closeErr)
	}
}

func TestNewSendAnnounceStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test/prefix"

	sas := newAnnouncementWriter(mockStream, prefix)

	require.NotNil(t, sas)
	assert.Equal(t, prefix, sas.prefix)
	assert.Equal(t, mockStream, sas.stream)
	assert.NotNil(t, sas.actives)
}

func TestSendAnnounceStream_SendAnnouncement(t *testing.T) {
	tests := map[string]struct {
		prefix         string
		broadcastPath  string
		expectError    bool
		shouldBeActive bool
	}{
		"valid path": {
			prefix:         "/test",
			broadcastPath:  "/test/stream1",
			expectError:    false,
			shouldBeActive: true,
		},
		"invalid path": {
			prefix:         "/test",
			broadcastPath:  "/other/stream1",
			expectError:    true,
			shouldBeActive: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}
			// For valid path, expect Write calls for ACTIVE message
			if !tt.expectError {
				mockStream.On("Write", mock.Anything).Return(0, nil)
			}

			sas := newAnnouncementWriter(mockStream, tt.prefix)

			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			err := sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			time.Sleep(50 * time.Millisecond) // Allow goroutine to process

			if tt.shouldBeActive {
				assert.Len(t, sas.actives, 1)
			} else {
				assert.Len(t, sas.actives, 0)
			}

			if !tt.expectError {
				mockStream.AssertExpectations(t)
			}
		})
	}
}

func TestSendAnnounceStream_SendAnnouncement_ClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	prefix := "/test"

	sas := newAnnouncementWriter(mockStream, prefix)
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream"))

	err := sas.SendAnnouncement(ann)
	assert.Error(t, err)
}

func TestSendAnnounceStream_SendAnnouncement_WriteError(t *testing.T) {
	tests := map[string]struct {
		name         string
		setupError   func() error
		expectClosed bool
		expectAnnErr bool
	}{
		"stream_error": {
			setupError: func() error {
				return &quic.StreamError{
					StreamID:  quic.StreamID(123),
					ErrorCode: quic.StreamErrorCode(42),
				}
			},
			expectClosed: true,
			expectAnnErr: true,
		},
		"generic_error": {
			setupError: func() error {
				return errors.New("generic write error")
			},
			expectClosed: false,
			expectAnnErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			writeError := tt.setupError()
			mockStream.On("Write", mock.Anything).Return(0, writeError)

			sas := newAnnouncementWriter(mockStream, "/test")

			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

			err := sas.SendAnnouncement(ann)

			assert.Error(t, err)
			assert.Equal(t, tt.expectClosed, sas.closed)

			if tt.expectAnnErr {
				var announceErr *AnnounceError
				assert.ErrorAs(t, err, &announceErr)
				assert.NotNil(t, sas.closeErr)
				assert.ErrorAs(t, sas.closeErr, &announceErr)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_Close(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	mockStream.On("Close").Return(nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	err := sas.Close()
	assert.NoError(t, err)
	assert.True(t, sas.closed)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_Close_AlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sas := newAnnouncementWriter(mockStream, "/test")
	sas.closed = true

	err := sas.Close()
	assert.NoError(t, err)
}

func TestSendAnnounceStream_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		errorCode    AnnounceErrorCode
		expectClosed bool
	}{
		"internal error": {
			errorCode:    InternalAnnounceErrorCode,
			expectClosed: true,
		},
		"duplicated announce error": {
			errorCode:    DuplicatedAnnounceErrorCode,
			expectClosed: true,
		},
		"uninterested error": {
			errorCode:    UninterestedErrorCode,
			expectClosed: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}
			mockStream.On("StreamID").Return(quic.StreamID(123))
			mockStream.On("CancelWrite", quic.StreamErrorCode(tt.errorCode)).Return()
			mockStream.On("CancelRead", quic.StreamErrorCode(tt.errorCode)).Return()

			sas := newAnnouncementWriter(mockStream, "/test")

			err := sas.CloseWithError(tt.errorCode)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectClosed, sas.closed)
			assert.NotNil(t, sas.closeErr)

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_CloseWithError_AlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sas := newAnnouncementWriter(mockStream, "/test")
	sas.closed = true
	existingErr := &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  quic.StreamID(123),
			ErrorCode: quic.StreamErrorCode(InternalAnnounceErrorCode),
		},
	}
	sas.closeErr = existingErr

	err := sas.CloseWithError(DuplicatedAnnounceErrorCode)
	// エラーが返されることを期待する（既存のエラー）
	assert.Equal(t, existingErr, err)
	assert.Equal(t, existingErr, sas.closeErr) // Should keep existing error
}

func TestSendAnnounceStreamInterface(t *testing.T) {
	// Test that sendAnnounceStream implements AnnouncementWriter interface
	var _ AnnouncementWriter = (*AnnouncementWriter)(nil)
}

func TestSendAnnounceStream_ConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	// Test concurrent access to SendAnnouncement
	go func() {
		for i := 0; i < 10; i++ {
			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
			sas.SendAnnouncement(ann)
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))
			sas.SendAnnouncement(ann)
			time.Sleep(time.Millisecond)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	// Test should complete without race conditions
}

func TestSendAnnounceStream_SendAnnouncement_MultipleAnnouncements(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

	err1 := sas.SendAnnouncement(ann1)
	err2 := sas.SendAnnouncement(ann2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 2)
	// Verify both announcements are tracked
	assert.Contains(t, sas.actives, "/stream1")
	assert.Contains(t, sas.actives, "/stream2")
	assert.Equal(t, ann1, sas.actives["/stream1"])
	assert.Equal(t, ann2, sas.actives["/stream2"])

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_SendAnnouncement_ReplaceExisting(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream1")) // Same suffix

	err1 := sas.SendAnnouncement(ann1)
	err2 := sas.SendAnnouncement(ann2)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 1)
	// Verify the second announcement replaced the first
	assert.Equal(t, ann2, sas.actives["/stream1"])
	assert.False(t, ann1.IsActive()) // First announcement should be ended

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_SendAnnouncement_ReplaceExisting_Debug(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream1")) // Same suffix

	// Debug: Check GetSuffix behavior
	suffix1, ok1 := ann1.BroadcastPath().GetSuffix(sas.prefix)
	t.Logf("Ann1 - Path: %s, Prefix: %s, Suffix: '%s', OK: %t", ann1.BroadcastPath(), sas.prefix, suffix1, ok1)

	suffix2, ok2 := ann2.BroadcastPath().GetSuffix(sas.prefix)
	t.Logf("Ann2 - Path: %s, Prefix: %s, Suffix: '%s', OK: %t", ann2.BroadcastPath(), sas.prefix, suffix2, ok2)

	// Debug: Initial state
	t.Logf("Initial actives count: %d", len(sas.actives))

	err1 := sas.SendAnnouncement(ann1)
	t.Logf("After ann1 - Error: %v, Actives count: %d", err1, len(sas.actives))
	if len(sas.actives) > 0 {
		for k, v := range sas.actives {
			t.Logf("Active key: '%s', value: %p", k, v)
		}
	}

	err2 := sas.SendAnnouncement(ann2)
	t.Logf("After ann2 - Error: %v, Actives count: %d", err2, len(sas.actives))
	if len(sas.actives) > 0 {
		for k, v := range sas.actives {
			t.Logf("Active key: '%s', value: %p", k, v)
		}
	}

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 1)
	// Verify the second announcement replaced the first
	assert.Equal(t, ann2, sas.actives["/stream1"])
	assert.False(t, ann1.IsActive()) // First announcement should be ended

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_SendAnnouncement_SameInstance(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(2) // Two calls expected

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	// Send the same announcement twice
	err1 := sas.SendAnnouncement(ann)
	err2 := sas.SendAnnouncement(ann)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 1)
	assert.True(t, ann.IsActive()) // Should still be active

	// Write should be called twice since SendAnnouncement always sends a message
	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_AnnouncementEnd_BackgroundProcessing(t *testing.T) {
	mockStream := &MockQUICStream{}
	// Expect Write calls for both ACTIVE and ENDED messages
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(2)

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	err := sas.SendAnnouncement(ann)
	assert.NoError(t, err)
	assert.Len(t, sas.actives, 1)

	// End the announcement
	ann.End()

	// Allow time for background goroutine to process
	time.Sleep(100 * time.Millisecond)

	// Announcement should be removed from actives
	assert.Len(t, sas.actives, 0)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_AnnouncementEnd_ClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil).Once() // Only ACTIVE message

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	err := sas.SendAnnouncement(ann)
	assert.NoError(t, err)

	// Close the stream
	sas.closed = true

	// End the announcement
	ann.End()

	// Allow time for background goroutine to process
	time.Sleep(100 * time.Millisecond)

	// Announcement should be removed from actives even if stream is closed
	assert.Len(t, sas.actives, 0)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_AnnouncementEnd_WriteError(t *testing.T) {
	mockStream := &MockQUICStream{}
	// First call succeeds (ACTIVE), second call fails (ENDED)
	mockStream.On("Write", mock.Anything).Return(0, nil).Once()
	writeError := &quic.StreamError{
		StreamID:  quic.StreamID(123),
		ErrorCode: quic.StreamErrorCode(42),
	}
	mockStream.On("Write", mock.Anything).Return(0, writeError).Once()

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	err := sas.SendAnnouncement(ann)
	assert.NoError(t, err)

	// End the announcement
	ann.End()

	// Allow time for background goroutine to process
	time.Sleep(100 * time.Millisecond)

	// Stream should be closed due to write error in background
	assert.True(t, sas.closed)
	assert.NotNil(t, sas.closeErr)

	var announceErr *AnnounceError
	assert.ErrorAs(t, sas.closeErr, &announceErr)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_BoundaryValues(t *testing.T) {
	tests := map[string]struct {
		prefix        string
		broadcastPath string
		expectError   bool
	}{
		"empty_prefix": {
			prefix:        "",
			broadcastPath: "/stream1",
			expectError:   false,
		},
		"root_prefix": {
			prefix:        "/",
			broadcastPath: "/stream1",
			expectError:   false,
		},
		"long_prefix": {
			prefix:        "/very/long/nested/prefix/path",
			broadcastPath: "/very/long/nested/prefix/path/stream1",
			expectError:   false,
		}, "matching_prefix_path": {
			prefix:        "/test",
			broadcastPath: "/test",
			expectError:   false, // Empty suffix is valid
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			// Always expect Write calls since GetSuffix might succeed
			mockStream.On("Write", mock.Anything).Return(0, nil).Maybe()

			sas := newAnnouncementWriter(mockStream, tt.prefix)

			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			err := sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_Close_StreamError(t *testing.T) {
	closeError := errors.New("close failed")
	mockStream := &MockQUICStream{}
	mockStream.On("Close").Return(closeError)

	sas := newAnnouncementWriter(mockStream, "/test")

	err := sas.Close()
	assert.Error(t, err)
	assert.Equal(t, closeError, err)
	assert.True(t, sas.closed)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_CloseWithError_StreamIDAccess(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("StreamID").Return(quic.StreamID(456))
	mockStream.On("CancelWrite", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return()
	mockStream.On("CancelRead", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return()

	sas := newAnnouncementWriter(mockStream, "/test")

	err := sas.CloseWithError(InternalAnnounceErrorCode)
	assert.NoError(t, err)
	assert.True(t, sas.closed)
	assert.NotNil(t, sas.closeErr)

	// Verify the StreamID in the error
	var announceErr *AnnounceError
	assert.ErrorAs(t, sas.closeErr, &announceErr)
	assert.Equal(t, quic.StreamID(456), announceErr.StreamError.StreamID)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_SendAnnouncement_InvalidPath_EdgeCases(t *testing.T) {
	tests := map[string]struct {
		prefix        string
		broadcastPath string
		expectError   bool
		description   string
	}{
		"different_root": {
			prefix:        "/test",
			broadcastPath: "/other/stream1",
			expectError:   true,
			description:   "Completely different path root",
		},
		"partial_match": {
			prefix:        "/test/sub",
			broadcastPath: "/test/other",
			expectError:   true,
			description:   "Partial path match but different branch",
		},
		"empty_path": {
			prefix:        "/test",
			broadcastPath: "",
			expectError:   true,
			description:   "Empty broadcast path",
		},
		"path_shorter_than_prefix": {
			prefix:        "/test/long/prefix",
			broadcastPath: "/test",
			expectError:   true,
			description:   "Broadcast path shorter than prefix",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			// No Write expectations since these should all fail

			sas := newAnnouncementWriter(mockStream, tt.prefix)

			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			err := sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "invalid broadcast path")
			} else {
				assert.NoError(t, err, tt.description)
			}

			// No announcements should be added to actives for invalid paths
			assert.Len(t, sas.actives, 0)

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_Concurrency_SafeAccess(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	// Test concurrent access to the same suffix
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	ctx := context.Background()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 50; i++ {
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
			if err := sas.SendAnnouncement(ann); err != nil {
				errors <- err
				return
			}
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 50; i++ {
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
			if err := sas.SendAnnouncement(ann); err != nil {
				errors <- err
				return
			}
			time.Sleep(time.Microsecond)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	close(errors)
	for err := range errors {
		t.Errorf("Unexpected error during concurrent access: %v", err)
	}
	// Should have exactly one active announcement
	assert.Len(t, sas.actives, 1)
	assert.Contains(t, sas.actives, "/stream1")

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_StateConsistency(t *testing.T) {
	mockStream := &MockQUICStream{}

	sas := newAnnouncementWriter(mockStream, "/test")

	// Test initial state
	assert.False(t, sas.closed)
	assert.Nil(t, sas.closeErr)
	assert.NotNil(t, sas.actives)
	assert.Len(t, sas.actives, 0)
	assert.Equal(t, "/test", sas.prefix)
	assert.Equal(t, mockStream, sas.stream)
}

func TestSendAnnounceStream_SendAnnouncement_NilAnnouncement(t *testing.T) {
	mockStream := &MockQUICStream{}
	sas := newAnnouncementWriter(mockStream, "/test")

	// Test with nil announcement - should cause panic or error
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil dereference
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	err := sas.SendAnnouncement(nil)
	if err != nil {
		t.Logf("Error with nil announcement: %v", err)
	}
}

func TestSendAnnounceStream_MultipleClose(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Close").Return(nil).Once()

	sas := newAnnouncementWriter(mockStream, "/test")

	// First close should succeed
	err1 := sas.Close()
	assert.NoError(t, err1)
	assert.True(t, sas.closed)

	// Second close should return without error (idempotent)
	err2 := sas.Close()
	assert.NoError(t, err2)

	// Third close should also return without error
	err3 := sas.Close()
	assert.NoError(t, err3)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_CloseWithError_Multiple(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("StreamID").Return(quic.StreamID(123)).Once()
	mockStream.On("CancelWrite", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return().Once()
	mockStream.On("CancelRead", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return().Once()

	sas := newAnnouncementWriter(mockStream, "/test")

	// First CloseWithError should succeed
	err1 := sas.CloseWithError(InternalAnnounceErrorCode)
	assert.NoError(t, err1)
	assert.True(t, sas.closed)
	assert.NotNil(t, sas.closeErr)

	originalErr := sas.closeErr

	// Second CloseWithError should return existing error
	err2 := sas.CloseWithError(DuplicatedAnnounceErrorCode)
	assert.Equal(t, originalErr, err2)
	assert.Equal(t, originalErr, sas.closeErr) // Should not change

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_InterfaceCompliance(t *testing.T) {
	// Test that sendAnnounceStream fully implements AnnouncementWriter interface
	mockStream := &MockQUICStream{}
	mockStream.On("Close").Return(nil).Maybe()
	mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()

	sas := newAnnouncementWriter(mockStream, "/test")

	// Test all interface methods are callable
	var writer AnnouncementWriter = sas

	// Method existence test
	assert.NotNil(t, writer.SendAnnouncement)
	assert.NotNil(t, writer.Close)
	assert.NotNil(t, writer.CloseWithError)

	// These should not panic
	_ = writer.Close()
	_ = writer.CloseWithError(InternalAnnounceErrorCode)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_Performance_ManyAnnouncements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	ctx := context.Background()
	const numAnnouncements = 1000

	start := time.Now()

	// Create many different announcements
	for i := 0; i < numAnnouncements; i++ {
		ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream%d", i)))
		err := sas.SendAnnouncement(ann)
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	t.Logf("Sent %d announcements in %v (%.2f announcements/sec)",
		numAnnouncements, duration, float64(numAnnouncements)/duration.Seconds())

	assert.Len(t, sas.actives, numAnnouncements)

	mockStream.AssertExpectations(t)
}

// Performance and cleanup tests

func TestSendAnnounceStream_Performance_LargeNumberOfAnnouncements(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	const numAnnouncements = 1000
	announcements := make([]*Announcement, numAnnouncements)

	// Measure time to send many announcements
	start := time.Now()
	for i := 0; i < numAnnouncements; i++ {
		announcements[i] = createTestAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream%d", i)))
		err := sas.SendAnnouncement(announcements[i])
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	t.Logf("Time to send %d announcements: %v", numAnnouncements, duration)
	t.Logf("Average time per announcement: %v", duration/numAnnouncements)

	// Verify all announcements are tracked
	assert.Len(t, sas.actives, numAnnouncements)

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_Cleanup_ResourceLeaks(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	// Create and end many announcements to test cleanup
	for i := 0; i < 100; i++ {
		ann := createTestAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream%d", i)))

		err := sas.SendAnnouncement(ann)
		assert.NoError(t, err)

		// Immediately end the announcement
		ann.End()
	}

	// Allow time for all background processing
	waitForBackgroundProcessing()

	// All announcements should be cleaned up
	assert.Len(t, sas.actives, 0, "Expected all announcements to be cleaned up")

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_Cleanup_PartialCleanup(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	// Create multiple announcements
	announcements := createMultipleAnnouncements(ctx, "/test", 5)

	for _, ann := range announcements {
		err := sas.SendAnnouncement(ann)
		assert.NoError(t, err)
	}

	// End only some announcements
	announcements[1].End()
	announcements[3].End()

	waitForBackgroundProcessing()

	// Should have 3 remaining announcements
	verifyAnnouncementState(t, sas, 3, []string{"/stream1", "/stream3", "/stream5"})

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_Memory_AnnouncementLifecycle(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")

	// Test announcement lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	ann := createTestAnnouncement(ctx, BroadcastPath("/test/stream1"))

	err := sas.SendAnnouncement(ann)
	assert.NoError(t, err)
	verifyAnnouncementState(t, sas, 1, []string{"/stream1"})

	// Cancel context (simulating cancellation)
	cancel()

	// End announcement
	ann.End()
	waitForBackgroundProcessing()

	// Should be cleaned up
	verifyAnnouncementState(t, sas, 0, []string{})

	mockStream.AssertExpectations(t)
}

// Benchmark for concurrent access (now that race conditions are fixed)
func BenchmarkSendAnnounceStream_ConcurrentAccess(b *testing.B) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ann := createTestAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream%d", i)))
			err := sas.SendAnnouncement(ann)
			if err != nil {
				b.Fatalf("Unexpected error: %v", err)
			}
			i++
		}
	})
}

// Additional comprehensive test cases

func TestSendAnnounceStream_AnnouncementLifecycle_CompleteFlow(t *testing.T) {
	mockStream := &MockQUICStream{}
	// Expected sequence: ACTIVE message, then ENDED message
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(2)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	// Create and send announcement
	ann := createTestAnnouncement(ctx, BroadcastPath("/test/stream1"))
	err := sas.SendAnnouncement(ann)
	assert.NoError(t, err)

	// Verify announcement is active
	assert.True(t, ann.IsActive())
	verifyAnnouncementState(t, sas, 1, []string{"/stream1"})

	// End the announcement
	ann.End()
	assert.False(t, ann.IsActive())

	// Wait for background processing
	waitForBackgroundProcessing()

	// Verify cleanup
	verifyAnnouncementState(t, sas, 0, []string{})

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_MultipleLifecycles_Interleaved(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	// Create multiple announcements with different lifecycles
	ann1 := createTestAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := createTestAnnouncement(ctx, BroadcastPath("/test/stream2"))
	ann3 := createTestAnnouncement(ctx, BroadcastPath("/test/stream3"))

	// Send all announcements
	assert.NoError(t, sas.SendAnnouncement(ann1))
	assert.NoError(t, sas.SendAnnouncement(ann2))
	assert.NoError(t, sas.SendAnnouncement(ann3))

	verifyAnnouncementState(t, sas, 3, []string{"/stream1", "/stream2", "/stream3"})

	// End them in different order
	ann2.End()
	waitForBackgroundProcessing()
	verifyAnnouncementState(t, sas, 2, []string{"/stream1", "/stream3"})

	ann1.End()
	waitForBackgroundProcessing()
	verifyAnnouncementState(t, sas, 1, []string{"/stream3"})

	ann3.End()
	waitForBackgroundProcessing()
	verifyAnnouncementState(t, sas, 0, []string{})

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_WriteError_Types(t *testing.T) {
	tests := map[string]struct {
		createError  func() error
		expectClosed bool
		expectAnnErr bool
	}{
		"quic_stream_error": {
			createError: func() error {
				return &quic.StreamError{
					StreamID:  quic.StreamID(123),
					ErrorCode: quic.StreamErrorCode(42),
				}
			},
			expectClosed: true,
			expectAnnErr: true,
		}, "connection_error": {
			createError: func() error {
				return errors.New("connection closed")
			},
			expectClosed: false,
			expectAnnErr: false,
		},
		"io_error": {
			createError: func() error {
				return io.ErrUnexpectedEOF
			},
			expectClosed: false,
			expectAnnErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := createMockStreamWithBehavior("write_error")
			// Override with specific error
			mockStream.ExpectedCalls = nil
			mockStream.On("Write", mock.Anything).Return(0, tt.createError())

			sas := newAnnouncementWriter(mockStream, "/test")
			ctx := context.Background()
			ann := createTestAnnouncement(ctx, BroadcastPath("/test/stream1"))

			err := sas.SendAnnouncement(ann)
			assert.Error(t, err)

			assertStreamState(t, sas, tt.expectClosed, tt.expectAnnErr)

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_BackgroundProcessing_EdgeCases(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	// Test rapid succession of announcement creation and ending
	for i := 0; i < 10; i++ {
		ann := createTestAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/rapid%d", i)))
		assert.NoError(t, sas.SendAnnouncement(ann))

		// End immediately after sending
		ann.End()
	}

	// Wait for all background processing
	waitForBackgroundProcessing()

	// All should be cleaned up
	verifyAnnouncementState(t, sas, 0, []string{})

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_PathValidation_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		broadcastPath string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid_nested_path",
			prefix:        "/app/v1",
			broadcastPath: "/app/v1/stream/video",
			expectError:   false,
		},
		{
			name:          "valid_empty_suffix",
			prefix:        "/test",
			broadcastPath: "/test",
			expectError:   false,
		},
		{
			name:          "invalid_wrong_prefix",
			prefix:        "/app/v1",
			broadcastPath: "/app/v2/stream",
			expectError:   true,
			errorContains: "invalid broadcast path",
		},
		{
			name:          "invalid_shorter_path",
			prefix:        "/app/v1/long",
			broadcastPath: "/app/v1",
			expectError:   true,
			errorContains: "invalid broadcast path",
		},
		{
			name:          "invalid_empty_path",
			prefix:        "/test",
			broadcastPath: "",
			expectError:   true,
			errorContains: "invalid broadcast path",
		},
		{
			name:          "valid_slash_prefix",
			prefix:        "/",
			broadcastPath: "/anything",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			if !tt.expectError {
				mockStream.On("Write", mock.Anything).Return(0, nil)
			}

			sas := newAnnouncementWriter(mockStream, tt.prefix)
			ctx := context.Background()
			ann := createTestAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			err := sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				verifyAnnouncementState(t, sas, 0, []string{})
			} else {
				assert.NoError(t, err)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_StreamState_Transitions(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("StreamID").Return(quic.StreamID(123))
	mockStream.On("CancelWrite", mock.Anything).Return()
	mockStream.On("CancelRead", mock.Anything).Return()

	sas := newAnnouncementWriter(mockStream, "/test")

	// Initial state
	assertStreamState(t, sas, false, false)

	// Close with error
	err := sas.CloseWithError(InternalAnnounceErrorCode)
	assert.NoError(t, err)
	assertStreamState(t, sas, true, true)

	// Verify error type
	var announceErr *AnnounceError
	assert.ErrorAs(t, sas.closeErr, &announceErr)
	assert.Equal(t, InternalAnnounceErrorCode, AnnounceErrorCode(announceErr.StreamError.ErrorCode))

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name         string
		setupError   func() error
		expectType   string
		expectClosed bool
	}{
		{
			name: "stream_error_propagation",
			setupError: func() error {
				return &quic.StreamError{
					StreamID:  quic.StreamID(456),
					ErrorCode: quic.StreamErrorCode(99),
				}
			},
			expectType:   "*moqt.AnnounceError",
			expectClosed: true,
		},
		{
			name: "generic_error_propagation",
			setupError: func() error {
				return errors.New("network error")
			},
			expectType:   "*errors.errorString",
			expectClosed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			writeError := tt.setupError()
			mockStream.On("Write", mock.Anything).Return(0, writeError)

			sas := newAnnouncementWriter(mockStream, "/test")
			ctx := context.Background()
			ann := createTestAnnouncement(ctx, BroadcastPath("/test/stream1"))

			err := sas.SendAnnouncement(ann)
			assert.Error(t, err)

			// Verify error type
			assert.Equal(t, tt.expectType, fmt.Sprintf("%T", err))
			assert.Equal(t, tt.expectClosed, sas.closed)

			mockStream.AssertExpectations(t)
		})
	}
}

// Test for multiple announcements lifecycle management
func TestSendAnnounceStream_MultipleAnnouncements_LifecycleManagement(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test")
	ctx := context.Background()

	// Create and send multiple announcements
	announcements := createMultipleAnnouncements(ctx, "/test", 5)

	for _, ann := range announcements {
		err := sas.SendAnnouncement(ann)
		assert.NoError(t, err)
	}

	// Verify all are active
	verifyAnnouncementState(t, sas, 5, []string{"/stream1", "/stream2", "/stream3", "/stream4", "/stream5"})

	// End one announcement
	announcements[2].End() // stream3
	waitForBackgroundProcessing()

	// Should have 4 remaining
	verifyAnnouncementState(t, sas, 4, []string{"/stream1", "/stream2", "/stream4", "/stream5"})

	// Replace one announcement (same path)
	newAnn := createTestAnnouncement(ctx, BroadcastPath("/test/stream2"))
	err := sas.SendAnnouncement(newAnn)
	assert.NoError(t, err)

	// Should still have 4, but stream2 should be replaced
	verifyAnnouncementState(t, sas, 4, []string{"/stream1", "/stream2", "/stream4", "/stream5"})
	assert.Equal(t, newAnn, sas.actives["/stream2"])

	// End all remaining
	for _, ann := range announcements {
		if ann.IsActive() {
			ann.End()
		}
	}
	newAnn.End()

	waitForBackgroundProcessing()
	verifyAnnouncementState(t, sas, 0, []string{})

	mockStream.AssertExpectations(t)
}

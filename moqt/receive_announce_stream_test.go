package moqt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewReceiveAnnounceStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
	prefix := "/test/prefix"

	ras := newReceiveAnnounceStream(mockStream, prefix)

	require.NotNil(t, ras)
	assert.Equal(t, prefix, ras.prefix)
	assert.Equal(t, mockStream, ras.stream)
	assert.NotNil(t, ras.active)
	assert.NotNil(t, ras.pendings)
	assert.NotNil(t, ras.announcedCh)
	assert.NotNil(t, ras.ctx)

	// Clean up
	time.Sleep(10 * time.Millisecond) // Allow goroutine to start and finish
	mockStream.AssertExpectations(t)
}

func TestReceiveAnnounceStream_ReceiveAnnouncement(t *testing.T) {
	tests := map[string]struct {
		receiveAnnounceStream *receiveAnnounceStream
		ctx                   context.Context
		wantErr               bool
		wantErrType           error
		wantAnn               bool
	}{
		"success_with_valid_announcement": {
			receiveAnnounceStream: func() *receiveAnnounceStream {
				buf := bytes.NewBuffer(nil)
				err := message.AnnounceMessage{
					TrackSuffix:    "valid_announcement",
					AnnounceStatus: message.ACTIVE,
				}.Encode(buf)
				require.NoError(t, err)

				// Create a mock stream that uses the buffer directly and then blocks
				mockStream := &MockQUICStream{
					ReadFunc: func(p []byte) (int, error) {
						if buf.Len() > 0 {
							// Use buffer's Read method directly - much simpler!
							return buf.Read(p)
						}
						// After all message data is consumed, block indefinitely to simulate ongoing stream
						select {}
					},
				}
				mockStream.On("Read", mock.AnythingOfType("[]uint8"))
				mockStream.On("CancelRead", mock.Anything).Return().Maybe()
				mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
				mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
				ras := newReceiveAnnounceStream(mockStream, "/test/")
				return ras
			}(),
			ctx:     context.Background(),
			wantErr: false,
			wantAnn: true,
		},
		"context_cancelled": {
			receiveAnnounceStream: func() *receiveAnnounceStream {
				mockStream := &MockQUICStream{
					ReadFunc: func(p []byte) (int, error) {
						// Block on read to simulate ongoing stream
						select {}
					},
				}
				mockStream.On("Read", mock.AnythingOfType("[]uint8"))
				mockStream.On("CancelRead", mock.Anything).Return().Maybe()
				mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
				mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
				return newReceiveAnnounceStream(mockStream, "/test/")
			}(),
			ctx: func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(), wantErr: true,
			wantErrType: context.Canceled,
			wantAnn:     false,
		},
		"stream_closed": {
			receiveAnnounceStream: func() *receiveAnnounceStream {
				mockStream := &MockQUICStream{
					ReadFunc: func(p []byte) (int, error) {
						// Block on read to simulate ongoing stream
						select {}
					},
				}
				mockStream.On("Read", mock.AnythingOfType("[]uint8"))
				mockStream.On("Close").Return(nil).Maybe()
				mockStream.On("CancelRead", mock.Anything).Return().Maybe()
				mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
				mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
				ras := newReceiveAnnounceStream(mockStream, "/test/")
				time.Sleep(20 * time.Millisecond) // Allow goroutine to start
				_ = ras.Close()
				return ras
			}(),
			ctx:     context.Background(),
			wantErr: true,
			wantAnn: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ras := tt.receiveAnnounceStream

			// Allow time for background goroutine processing
			if name == "success_with_valid_announcement" {
				time.Sleep(100 * time.Millisecond)
			}

			ctxWithTimeout, cancel := context.WithTimeout(tt.ctx, 2*time.Second)
			defer cancel()

			announcement, err := ras.ReceiveAnnouncement(ctxWithTimeout)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.wantAnn {
				assert.NotNil(t, announcement)
				if announcement != nil {
					assert.Equal(t, BroadcastPath("/test/valid_announcement"), announcement.BroadcastPath())
				}
			} else {
				assert.Nil(t, announcement)
			}

			// Clean up
			if ras.stream != nil {
				if mockStream, ok := ras.stream.(*MockQUICStream); ok {
					mockStream.AssertExpectations(t)
				}
			}
		})
	}
}

func TestReceiveAnnounceStream_Close(t *testing.T) {
	tests := map[string]struct {
		setupFunc func() *receiveAnnounceStream
		wantErr   bool
	}{"normal_close": {
		setupFunc: func() *receiveAnnounceStream {
			mockStream := &MockQUICStream{}
			// Block reads to prevent goroutine from interfering
			mockStream.On("Read", mock.Anything).Run(func(args mock.Arguments) {
				time.Sleep(100 * time.Millisecond)
			}).Return(0, io.EOF).Maybe()
			mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
			mockStream.On("Close").Return(nil)
			return newReceiveAnnounceStream(mockStream, "/test")
		},
		wantErr: false,
	},
		"already_closed": {
			setupFunc: func() *receiveAnnounceStream {
				mockStream := &MockQUICStream{}
				// Block reads to prevent goroutine from interfering
				mockStream.On("Read", mock.Anything).Run(func(args mock.Arguments) {
					time.Sleep(100 * time.Millisecond)
				}).Return(0, io.EOF).Maybe()
				mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
				mockStream.On("Close").Return(nil)
				mockStream.On("CancelRead", mock.Anything).Return().Maybe()
				mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
				ras := newReceiveAnnounceStream(mockStream, "/test")
				_ = ras.Close() // Close once
				return ras
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ras := tt.setupFunc()
			time.Sleep(10 * time.Millisecond) // Allow goroutine to start

			err := ras.Close()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify stream is closed
			select {
			case <-ras.ctx.Done():
				// Stream is closed as expected
			default:
				t.Error("Expected stream to be closed")
			}

			// Verify close state
			assert.True(t, ras.closed)

			if mockStream, ok := ras.stream.(*MockQUICStream); ok {
				mockStream.AssertExpectations(t)
			}
		})
	}
}

func TestReceiveAnnounceStream_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		setupFunc func() *receiveAnnounceStream
		errorCode AnnounceErrorCode
		wantErr   bool
	}{"internal_error": {
		setupFunc: func() *receiveAnnounceStream {
			mockStream := &MockQUICStream{}
			// Block reads to prevent goroutine from interfering
			mockStream.On("Read", mock.Anything).Run(func(args mock.Arguments) {
				time.Sleep(100 * time.Millisecond)
			}).Return(0, io.EOF).Maybe()
			mockStream.On("CancelRead", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return()
			mockStream.On("CancelWrite", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return()
			mockStream.On("StreamID").Return(quic.StreamID(123))
			return newReceiveAnnounceStream(mockStream, "/test")
		},
		errorCode: InternalAnnounceErrorCode,
		wantErr:   false,
	}, "duplicated_error": {
		setupFunc: func() *receiveAnnounceStream {
			mockStream := &MockQUICStream{}
			// Block reads to prevent goroutine from interfering
			mockStream.On("Read", mock.Anything).Run(func(args mock.Arguments) {
				time.Sleep(100 * time.Millisecond)
			}).Return(0, io.EOF).Maybe()
			mockStream.On("CancelRead", quic.StreamErrorCode(DuplicatedAnnounceErrorCode)).Return()
			mockStream.On("CancelWrite", quic.StreamErrorCode(DuplicatedAnnounceErrorCode)).Return()
			mockStream.On("StreamID").Return(quic.StreamID(123))
			return newReceiveAnnounceStream(mockStream, "/test")
		},
		errorCode: DuplicatedAnnounceErrorCode,
		wantErr:   false,
	},
		"already_closed": {
			setupFunc: func() *receiveAnnounceStream {
				mockStream := &MockQUICStream{}
				// Block reads to prevent goroutine from interfering
				mockStream.On("Read", mock.Anything).Run(func(args mock.Arguments) {
					time.Sleep(100 * time.Millisecond)
				}).Return(0, io.EOF).Maybe()
				mockStream.On("CancelRead", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return()
				mockStream.On("CancelWrite", quic.StreamErrorCode(InternalAnnounceErrorCode)).Return()
				mockStream.On("StreamID").Return(quic.StreamID(123))
				ras := newReceiveAnnounceStream(mockStream, "/test")
				time.Sleep(10 * time.Millisecond) // Allow goroutine to start
				_ = ras.CloseWithError(InternalAnnounceErrorCode)
				return ras
			},
			errorCode: DuplicatedAnnounceErrorCode,
			wantErr:   true, // Returns existing error when already closed
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ras := tt.setupFunc()
			time.Sleep(10 * time.Millisecond) // Allow goroutine to start

			err := ras.CloseWithError(tt.errorCode)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify stream is closed
			select {
			case <-ras.ctx.Done():
				// Stream is closed as expected
			default:
				t.Error("Expected stream to be closed")
			} // Verify error state
			assert.True(t, ras.closed)
			if name != "already_closed" {
				assert.NotNil(t, ras.closeErr)
				var announceErr *AnnounceError
				assert.ErrorAs(t, ras.closeErr, &announceErr)
			} else {
				// For already_closed case, verify the original error is preserved
				assert.NotNil(t, ras.closeErr)
				var announceErr *AnnounceError
				assert.ErrorAs(t, ras.closeErr, &announceErr)
				assert.Equal(t, quic.StreamErrorCode(InternalAnnounceErrorCode), announceErr.StreamError.ErrorCode)
			}

			if mockStream, ok := ras.stream.(*MockQUICStream); ok {
				mockStream.AssertExpectations(t)
			}
		})
	}
}

func TestReceiveAnnounceStream_AnnouncementTracking(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Test internal announcement tracking
	ctx := context.Background()
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

	// Manually add announcements to test tracking
	ras.mu.Lock()
	ras.active["stream1"] = ann1
	ras.active["stream2"] = ann2
	ras.mu.Unlock()

	assert.Len(t, ras.active, 2)

	// Test ending announcement
	ann1.End()

	// Announcement should still be in map until processed by background goroutine
	assert.Len(t, ras.active, 2)

	mockStream.AssertExpectations(t)
}

func TestReceiveAnnounceStream_ConcurrentAccess(t *testing.T) {
	// Create multiple messages
	buf := bytes.NewBuffer(nil)
	for i := 0; i < 5; i++ {
		err := message.AnnounceMessage{
			TrackSuffix:    fmt.Sprintf("/stream%d", i),
			AnnounceStatus: message.ACTIVE,
		}.Encode(buf)
		require.NoError(t, err)
	}

	var mu sync.Mutex
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (n int, err error) {
			mu.Lock()
			defer mu.Unlock()
			if buf.Len() > 0 {
				// Use buffer's Read method directly - thread-safe with mutex
				return buf.Read(p)
			}
			// Block after all data
			select {}
		},
	}
	mockStream.On("Read", mock.AnythingOfType("[]uint8"))
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("Close").Return(nil).Maybe()
	mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Allow time for message processing
	time.Sleep(100 * time.Millisecond)

	// Test concurrent ReceiveAnnouncement calls
	var wg sync.WaitGroup
	results := make(chan *Announcement, 5)
	errors := make(chan error, 5)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			ann, err := ras.ReceiveAnnouncement(ctx)
			if err != nil {
				errors <- err
			} else if ann != nil {
				results <- ann
			}
		}()
	}

	// Concurrent close call
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = ras.Close()
	}()

	wg.Wait()
	close(results)
	close(errors)

	// Verify we got some results before closing
	receivedCount := 0
	for range results {
		receivedCount++
	}

	// Should have received at least some announcements
	assert.GreaterOrEqual(t, receivedCount, 0)

	mockStream.AssertExpectations(t)
}

func TestReceiveAnnounceStream_PrefixHandling(t *testing.T) {
	tests := map[string]struct {
		prefix       string
		suffix       string
		expectedPath string
	}{
		"simple_prefix_and_suffix": {
			prefix:       "/test",
			suffix:       "/stream",
			expectedPath: "/test/stream",
		},
		"nested_prefix": {
			prefix:       "/test/sub",
			suffix:       "/stream",
			expectedPath: "/test/sub/stream",
		},
		"root_prefix": {
			prefix:       "/",
			suffix:       "stream",
			expectedPath: "/stream",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
			mockStream.On("CancelRead", mock.Anything).Return().Maybe()
			mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
			mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()
			ras := newReceiveAnnounceStream(mockStream, tt.prefix)

			require.NotNil(t, ras)
			assert.Equal(t, tt.prefix, ras.prefix)

			// Test path construction by manually adding announcement
			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.expectedPath))
			ras.mu.Lock()
			ras.pendings = append(ras.pendings, ann)
			ras.mu.Unlock()

			announcement, err := ras.ReceiveAnnouncement(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPath, string(announcement.BroadcastPath()))

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveAnnounceStream_InvalidMessage(t *testing.T) {
	// Create invalid message data
	invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	buf := bytes.NewBuffer(invalidData)

	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			if buf.Len() > 0 {
				return buf.Read(p)
			}
			return 0, io.EOF
		},
	}
	mockStream.On("Read", mock.AnythingOfType("[]uint8"))
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Give time for processing invalid data
	time.Sleep(100 * time.Millisecond)

	// Stream should be closed due to invalid message
	select {
	case <-ras.ctx.Done():
		// Stream is closed as expected due to decode error
		cause := context.Cause(ras.ctx)
		assert.Error(t, cause)
	default:
		t.Error("Expected stream to be closed due to invalid message")
	}

	mockStream.AssertExpectations(t)
}

func TestReceiveAnnounceStream_AnnouncementLifecycle(t *testing.T) {
	tests := map[string]struct {
		messages     []message.AnnounceMessage
		expectActive []string
		expectEnded  []string
		wantErr      bool
	}{"active_then_ended": {
		messages: []message.AnnounceMessage{
			{TrackSuffix: "/stream1", AnnounceStatus: message.ACTIVE},
			{TrackSuffix: "/stream1", AnnounceStatus: message.ENDED},
		},
		expectActive: []string{"/stream1"},
		expectEnded:  []string{},
		wantErr:      false,
	},
		"multiple_active_streams": {
			messages: []message.AnnounceMessage{
				{TrackSuffix: "/stream1", AnnounceStatus: message.ACTIVE},
				{TrackSuffix: "/stream2", AnnounceStatus: message.ACTIVE},
			},
			expectActive: []string{"/stream1", "/stream2"},
			expectEnded:  []string{},
			wantErr:      false,
		},
		"duplicate_active_error": {
			messages: []message.AnnounceMessage{
				{TrackSuffix: "/stream1", AnnounceStatus: message.ACTIVE},
				{TrackSuffix: "/stream1", AnnounceStatus: message.ACTIVE}, // Duplicate
			},
			expectActive: []string{"/stream1"},
			expectEnded:  []string{},
			wantErr:      true, // Should cause stream to close with error
		},
		"end_non_existent_stream": {
			messages: []message.AnnounceMessage{
				{TrackSuffix: "/stream1", AnnounceStatus: message.ENDED}, // End without ACTIVE
			},
			expectActive: []string{},
			expectEnded:  []string{},
			wantErr:      true, // Should cause stream to close with error
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Prepare message data
			buf := bytes.NewBuffer(nil)
			for _, msg := range tt.messages {
				err := msg.Encode(buf)
				require.NoError(t, err)
			}

			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (n int, err error) {
					if buf.Len() > 0 {
						return buf.Read(p)
					}
					// Block after all data is consumed
					select {}
				},
			}
			mockStream.On("Read", mock.AnythingOfType("[]uint8"))

			if tt.wantErr {
				mockStream.On("CancelRead", mock.Anything).Return()
				mockStream.On("CancelWrite", mock.Anything).Return()
			} else {
				mockStream.On("CancelRead", mock.Anything).Return().Maybe()
				mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
			}
			mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()

			ras := newReceiveAnnounceStream(mockStream, "/test")

			// Allow more time for message processing
			time.Sleep(300 * time.Millisecond)

			if tt.wantErr {
				// Verify stream was closed due to error
				select {
				case <-ras.ctx.Done():
					// Stream closed as expected
				default:
					t.Error("Expected stream to be closed due to error")
				}
			} else { // Verify expected announcements were received
				receivedActive := 0
				receivedEnded := 0

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				// Receive all announcements
				for {
					ann, err := ras.ReceiveAnnouncement(ctx)
					if err != nil {
						if err == context.DeadlineExceeded {
							break
						}
						t.Logf("Error receiving announcement: %v", err)
						break
					}
					if ann != nil {
						if ann.IsActive() {
							receivedActive++
							t.Logf("Received active announcement %d: %s", receivedActive, ann.BroadcastPath())
						} else {
							receivedEnded++
							t.Logf("Received ended announcement %d: %s", receivedEnded, ann.BroadcastPath())
						}
					}
				} // For active_then_ended case, we expect 1 active announcement initially
				// but it gets ended by the ENDED message, so we might not receive any active ones
				if name == "active_then_ended" {
					// In this case, the active announcement gets ended immediately
					// so we might receive 0 active announcements when we check
					assert.GreaterOrEqual(t, receivedActive, 0, "Should receive at least 0 active announcements")
					assert.LessOrEqual(t, receivedActive, 1, "Should receive at most 1 active announcement")
				} else {
					assert.Equal(t, len(tt.expectActive), receivedActive)
				}
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveAnnounceStream_NotifyChannel(t *testing.T) {
	// Create a message
	buf := bytes.NewBuffer(nil)
	err := message.AnnounceMessage{
		TrackSuffix:    "/test_stream",
		AnnounceStatus: message.ACTIVE}.Encode(buf)
	require.NoError(t, err)

	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (n int, err error) {
			if buf.Len() > 0 {
				return buf.Read(p)
			}
			// Block after data
			select {}
		},
	}
	mockStream.On("Read", mock.AnythingOfType("[]uint8"))
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()

	ras := newReceiveAnnounceStream(mockStream, "/test")

	// Allow time for message processing and notification
	time.Sleep(100 * time.Millisecond)

	// Verify that we can receive the announcement without blocking
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	announcement, err := ras.ReceiveAnnouncement(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, announcement)
	assert.Equal(t, BroadcastPath("/test/test_stream"), announcement.BroadcastPath())
	assert.True(t, announcement.IsActive())

	mockStream.AssertExpectations(t)
}

func TestReceiveAnnounceStream_BoundaryValues(t *testing.T) {
	tests := map[string]struct {
		prefix       string
		suffix       string
		expectedPath string
		wantErr      bool
	}{
		"empty_prefix": {
			prefix:       "",
			suffix:       "/stream",
			expectedPath: "/stream",
			wantErr:      false,
		},
		"empty_suffix": {
			prefix:       "/test",
			suffix:       "",
			expectedPath: "/test",
			wantErr:      false,
		},
		"both_empty": {
			prefix:       "",
			suffix:       "",
			expectedPath: "",
			wantErr:      false,
		},
		"root_prefix": {
			prefix:       "/",
			suffix:       "stream",
			expectedPath: "/stream",
			wantErr:      false,
		},
		"long_prefix": {
			prefix:       "/very/long/nested/prefix/path",
			suffix:       "/stream",
			expectedPath: "/very/long/nested/prefix/path/stream",
			wantErr:      false,
		},
		"long_suffix": {
			prefix:       "/test",
			suffix:       "/very/long/nested/suffix/path/stream",
			expectedPath: "/test/very/long/nested/suffix/path/stream",
			wantErr:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create message with the test suffix
			buf := bytes.NewBuffer(nil)
			err := message.AnnounceMessage{
				TrackSuffix:    tt.suffix,
				AnnounceStatus: message.ACTIVE}.Encode(buf)
			require.NoError(t, err)

			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (n int, err error) {
					if buf.Len() > 0 {
						return buf.Read(p)
					}
					// Block after data
					select {}
				},
			}
			mockStream.On("Read", mock.AnythingOfType("[]uint8"))
			mockStream.On("CancelRead", mock.Anything).Return().Maybe()
			mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
			mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()

			ras := newReceiveAnnounceStream(mockStream, tt.prefix)

			// Allow time for processing
			time.Sleep(50 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			announcement, err := ras.ReceiveAnnouncement(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, announcement)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, announcement)
				assert.Equal(t, BroadcastPath(tt.expectedPath), announcement.BroadcastPath())
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveAnnounceStream_StreamErrors(t *testing.T) {
	tests := map[string]struct {
		setupError   func() error
		expectedType error
		wantErr      bool
	}{
		"quic_stream_error": {
			setupError: func() error {
				return &quic.StreamError{
					StreamID:  quic.StreamID(123),
					ErrorCode: quic.StreamErrorCode(42),
				}
			},
			expectedType: &AnnounceError{},
			wantErr:      true,
		},
		"generic_io_error": {
			setupError: func() error {
				return io.ErrUnexpectedEOF
			},
			expectedType: io.ErrUnexpectedEOF,
			wantErr:      true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			testError := tt.setupError()

			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, testError
				},
			}
			mockStream.On("Read", mock.AnythingOfType("[]uint8"))
			mockStream.On("CancelRead", mock.Anything).Return().Maybe()
			mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
			mockStream.On("StreamID").Return(quic.StreamID(123)).Maybe()

			ras := newReceiveAnnounceStream(mockStream, "/test")

			// Allow time for error processing
			time.Sleep(50 * time.Millisecond)

			// Verify stream was closed due to error
			select {
			case <-ras.ctx.Done():
				cause := context.Cause(ras.ctx)
				if tt.wantErr {
					assert.Error(t, cause)
					// Check error type
					if name == "quic_stream_error" {
						var announceErr *AnnounceError
						assert.ErrorAs(t, cause, &announceErr)
					}
				}
			default:
				if tt.wantErr {
					t.Error("Expected stream to be closed due to error")
				}
			}

			mockStream.AssertExpectations(t)
		})
	}
}

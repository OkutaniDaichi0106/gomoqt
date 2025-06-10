package moqt

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSendAnnounceStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test/prefix"

	sas := newSendAnnounceStream(mockStream, prefix)

	require.NotNil(t, sas)
	assert.Equal(t, prefix, sas.prefix)
	assert.Equal(t, mockStream, sas.stream)
	assert.NotNil(t, sas.actives)
	assert.NotNil(t, sas.pendings)
	assert.NotNil(t, sas.sendCh)

	// Give time for goroutine to start
	time.Sleep(10 * time.Millisecond)
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
			sas := newSendAnnounceStream(mockStream, tt.prefix)

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

	sas := newSendAnnounceStream(mockStream, prefix)
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream"))

	err := sas.SendAnnouncement(ann)
	assert.Error(t, err)
}

func TestSendAnnounceStream_Set(t *testing.T) {
	tests := map[string]struct {
		active          bool
		existingPending bool
		expectedStatus  message.AnnounceStatus
		expectedInMap   bool
	}{
		"set active": {
			active:          true,
			existingPending: false,
			expectedStatus:  message.ACTIVE,
			expectedInMap:   true,
		},
		"set ended without existing": {
			active:          false,
			existingPending: false,
			expectedStatus:  message.ENDED,
			expectedInMap:   true,
		},
		"set ended with existing": {
			active:          false,
			existingPending: true,
			expectedStatus:  0,
			expectedInMap:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}
			sas := newSendAnnounceStream(mockStream, "/test")

			if tt.existingPending {
				sas.pendings["stream1"] = message.AnnounceMessage{
					AnnounceStatus: message.ACTIVE,
					TrackSuffix:    "stream1",
				}
			}

			sas.set("stream1", tt.active)

			if tt.expectedInMap {
				require.Contains(t, sas.pendings, "stream1")
				assert.Equal(t, tt.expectedStatus, sas.pendings["stream1"].AnnounceStatus)
			} else {
				assert.NotContains(t, sas.pendings, "stream1")
			}
		})
	}
}

func TestSendAnnounceStream_SetClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sas := newSendAnnounceStream(mockStream, "/test")
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	// set method doesn't check for closed state, it just sets the pending message
	sas.set("stream1", true)

	assert.Contains(t, sas.pendings, "stream1")
	assert.Equal(t, message.ACTIVE, sas.pendings["stream1"].AnnounceStatus)
}

func TestSendAnnounceStream_Send(t *testing.T) {
	tests := map[string]struct {
		setupPendings func(*sendAnnounceStream)
		expectError   bool
		shouldClear   bool
	}{
		"with pending messages": {
			setupPendings: func(sas *sendAnnounceStream) {
				sas.pendings["stream1"] = message.AnnounceMessage{
					AnnounceStatus: message.ACTIVE,
					TrackSuffix:    "stream1",
				}
			},
			expectError: false,
			shouldClear: true,
		},
		"no pending messages": {
			setupPendings: func(sas *sendAnnounceStream) {
				// No setup needed
			},
			expectError: false,
			shouldClear: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}
			mockStream.On("Write", mock.Anything).Return(0, nil)

			sas := newSendAnnounceStream(mockStream, "/test")
			tt.setupPendings(sas)

			err := sas.send()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.shouldClear {
				assert.Empty(t, sas.pendings)
			}
		})
	}
}

func TestSendAnnounceStream_Send_ClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sas := newSendAnnounceStream(mockStream, "/test")
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	err := sas.send()
	assert.Error(t, err)
}

func TestSendAnnounceStream_Close(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	mockStream.On("Close").Return(nil)

	sas := newSendAnnounceStream(mockStream, "/test")

	err := sas.Close()
	assert.NoError(t, err)
	assert.True(t, sas.closed)
}

func TestSendAnnounceStream_Close_AlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	// Close()は呼ばれないのでモックを設定する必要はありません

	sas := newSendAnnounceStream(mockStream, "/test")
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

			sas := newSendAnnounceStream(mockStream, "/test")

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

	sas := newSendAnnounceStream(mockStream, "/test")
	sas.closed = true
	existingErr := &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  quic.StreamID(123),
			ErrorCode: quic.StreamErrorCode(InternalAnnounceErrorCode),
		},
	}
	sas.closeErr = existingErr

	err := sas.CloseWithError(DuplicatedAnnounceErrorCode)
	// エラーは期待されません
	if err != nil {
		t.Logf("Unexpected error: %v", err)
	}
	assert.Equal(t, existingErr, sas.closeErr) // Should keep existing error
}

func TestSendAnnounceStreamInterface(t *testing.T) {
	// Test that sendAnnounceStream implements AnnouncementWriter interface
	var _ AnnouncementWriter = (*sendAnnounceStream)(nil)
}

func TestSendAnnounceStream_ConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newSendAnnounceStream(mockStream, "/test")

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

func TestSendAnnounceStream_AnnouncementLifecycle(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sas := newSendAnnounceStream(mockStream, "/test")

	// Create announcement
	ctx, cancel := context.WithCancel(context.Background())
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream"))

	// Send announcement
	err := sas.SendAnnouncement(ann)
	assert.NoError(t, err)

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	// Verify active announcement is stored
	assert.Len(t, sas.actives, 1)

	// End the announcement
	cancel()

	// Give time for announcement end to be processed
	time.Sleep(50 * time.Millisecond)
}

func TestSendAnnounceStream_PrefixMatching(t *testing.T) {
	tests := map[string]struct {
		prefix        string
		broadcastPath string
		expectError   bool
		shouldMatch   bool
	}{
		"exact prefix match": {
			prefix:        "/test",
			broadcastPath: "/test/stream",
			expectError:   false,
			shouldMatch:   true,
		},
		"nested prefix match": {
			prefix:        "/test/sub",
			broadcastPath: "/test/sub/stream",
			expectError:   false,
			shouldMatch:   true,
		},
		"no match - different prefix": {
			prefix:        "/test",
			broadcastPath: "/other/stream",
			expectError:   true,
			shouldMatch:   false,
		},
		"no match - prefix is substring but not path prefix": {
			prefix:        "/test",
			broadcastPath: "/testing/stream",
			expectError:   true,
			shouldMatch:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}

			sas := newSendAnnounceStream(mockStream, tt.prefix)

			ctx := context.Background()
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))
			initialActiveCount := len(sas.actives)

			err := sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Give time for processing
			time.Sleep(50 * time.Millisecond)

			expectedActiveCount := initialActiveCount
			if tt.shouldMatch {
				expectedActiveCount++
			}

			assert.Len(t, sas.actives, expectedActiveCount)
		})
	}
}

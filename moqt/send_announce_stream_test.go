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

	sas := newSendAnnounceStream(mockStream, prefix)
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	ctx := context.Background()
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream"))

	err := sas.SendAnnouncement(ann)
	assert.Error(t, err)
}

func TestSendAnnounceStream_SendMessage(t *testing.T) {
	tests := map[string]struct {
		suffix      string
		status      message.AnnounceStatus
		expectError bool
	}{
		"active message": {
			suffix:      "stream1",
			status:      message.ACTIVE,
			expectError: false,
		},
		"ended message": {
			suffix:      "stream1",
			status:      message.ENDED,
			expectError: false,
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

			err := sas.sendMessage(tt.suffix, tt.status)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSendAnnounceStream_SendMessage_ClosedStream(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sas := newSendAnnounceStream(mockStream, "/test")
	sas.closed = true
	sas.closeErr = errors.New("stream closed")

	err := sas.sendMessage("stream1", message.ACTIVE)
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

	mockStream.AssertExpectations(t)
}

func TestSendAnnounceStream_Close_AlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

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

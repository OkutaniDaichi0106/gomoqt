package moqt

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSessionStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	// Setup Read mock - will be used by the background goroutine in newSessionStream
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	ss := newSessionStream(mockStream)

	assert.NotNil(t, ss, "newSessionStream should not return nil")
	assert.NotNil(t, ss.SessionUpdated(), "SessionUpdated should return a valid channel")
	assert.Equal(t, mockStream, ss.stream, "stream should be set correctly")
	assert.Equal(t, uint64(0), ss.localBitrate, "local bitrate should be initialized to 0")
	assert.Equal(t, uint64(0), ss.remoteBitrate, "remote bitrate should be initialized to 0")

	// Give time for background goroutines to complete
	time.Sleep(50 * time.Millisecond)

	// Verify the session stream was created properly
	mockStream.AssertExpectations(t)
}

func TestSessionStream_updateSession(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	bitrate := uint64(1000000)
	// Create expected message for verification
	sum := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	sum.Encode(buf)
	mockStream.WriteFunc = buf.Write

	ss := newSessionStream(mockStream)

	err := ss.updateSession(bitrate)

	assert.NoError(t, err)
	assert.Equal(t, bitrate, ss.localBitrate, "local bitrate should be updated")

	// Give time for background goroutines to complete
	time.Sleep(50 * time.Millisecond)
	mockStream.AssertExpectations(t)
}

func TestSessionStream_updateSession_WriteError(t *testing.T) {
	mockStream := &MockQUICStream{}
	writeError := errors.New("write error")
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, writeError)

	ss := newSessionStream(mockStream)

	err := ss.updateSession(uint64(1000000))

	assert.Error(t, err)
	assert.Equal(t, writeError, err)

	// Give time for background goroutines to complete
	time.Sleep(50 * time.Millisecond)
	mockStream.AssertExpectations(t)
}

func TestSessionStream_SessionUpdated(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	ss := newSessionStream(mockStream)

	ch := ss.SessionUpdated()
	assert.NotNil(t, ch, "SessionUpdated should return a valid channel")
	// SessionUpdated() returns <-chan struct{}, not chan struct{}
	assert.IsType(t, (<-chan struct{})(nil), ch, "SessionUpdated should return a receive-only channel")

	// Give time for background goroutines to complete
	time.Sleep(50 * time.Millisecond)
	mockStream.AssertExpectations(t)
}

func TestSessionStream_updateSession_ZeroBitrate(t *testing.T) {
	mockStream := &MockQUICStream{}

	// Setup Read mock - EOF will trigger close from background goroutine
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(2, nil) // 2 bytes for zero bitrate message

	ss := newSessionStream(mockStream)

	// Give time for background goroutine to start
	time.Sleep(10 * time.Millisecond)

	err := ss.updateSession(0)
	assert.NoError(t, err, "updateSession(0) should not error")
	assert.Equal(t, uint64(0), ss.localBitrate, "local bitrate should be set to 0")

	mockStream.AssertCalled(t, "Write", mock.Anything)
	mockStream.AssertExpectations(t)
}

func TestSessionStream_updateSession_LargeBitrate(t *testing.T) {
	mockStream := &MockQUICStream{}

	// Setup Read mock - EOF will trigger close from background goroutine
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(10, nil) // 10 bytes for large bitrate message

	ss := newSessionStream(mockStream)

	// Give time for background goroutine to start
	time.Sleep(10 * time.Millisecond)

	largeBitrate := uint64(1<<62 - 1) // Large but valid value
	err := ss.updateSession(largeBitrate)
	assert.NoError(t, err, "updateSession with large bitrate should not error")
	assert.Equal(t, largeBitrate, ss.localBitrate, "local bitrate should be set correctly")

	mockStream.AssertCalled(t, "Write", mock.Anything)
	mockStream.AssertExpectations(t)
}

func TestSessionStream_listenUpdates(t *testing.T) {
	tests := map[string]struct {
		buffer        func() *bytes.Buffer
		expectBitrate uint64
	}{
		"valid message": {
			buffer: func() *bytes.Buffer {
				// Create a valid SessionUpdateMessage with a bitrate
				bitrate := uint64(1000000)
				sessionUpdate := message.SessionUpdateMessage{
					Bitrate: bitrate,
				}
				var buf bytes.Buffer
				err := sessionUpdate.Encode(&buf)
				if err != nil {
					panic("failed to encode SessionUpdateMessage: " + err.Error())
				}
				return &buf
			},
			expectBitrate: 1000000,
		},
		"empty message": {
			buffer: func() *bytes.Buffer {
				// Create an empty buffer to simulate no data
				return &bytes.Buffer{}
			},
			expectBitrate: 0,
		},
		"zero bitrate": {
			buffer: func() *bytes.Buffer {
				// Create a SessionUpdateMessage with a zero bitrate
				bitrate := uint64(0)
				sessionUpdate := message.SessionUpdateMessage{
					Bitrate: bitrate,
				}
				var buf bytes.Buffer
				err := sessionUpdate.Encode(&buf)
				if err != nil {
					panic("failed to encode SessionUpdateMessage: " + err.Error())
				}
				return &buf
			},
			expectBitrate: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := tt.buffer()
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (n int, err error) {
					return buf.Read(p)
				},
			}

			mockStream.On("Read", mock.Anything)

			ss := newSessionStream(mockStream)

			// Give time for listenUpdates to process the message
			time.Sleep(100 * time.Millisecond)
			// Check if we get the update notification
			select {
			case <-ss.SessionUpdated():
				if name == "valid message" {
					assert.Equal(t, tt.expectBitrate, ss.remoteBitrate, "remote bitrate should be updated correctly")
				}
			case <-time.After(200 * time.Millisecond):
				if name == "valid message" {
					t.Error("timed out waiting for session update")
				}
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestSessionStream_listenUpdates_StreamClosed(t *testing.T) {
	mockStream := &MockQUICStream{}

	// Set up mock to return EOF immediately
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Once()

	ss := newSessionStream(mockStream)

	// Give time for listenUpdates to encounter the EOF
	time.Sleep(50 * time.Millisecond)

	// Verify context was cancelled due to EOF (channel should be closed)
	select {
	case <-ss.SessionUpdated():
		// Channel is closed, which is expected
	case <-time.After(100 * time.Millisecond):
		t.Error("expected channel to be closed due to EOF")
	}

	mockStream.AssertExpectations(t)
}

func TestSessionStream_listenUpdates_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &MockQUICStream{}

	// Mock Read to potentially be called
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(ctx)

	ss := newSessionStream(mockStream)

	// Let listenUpdates start
	time.Sleep(10 * time.Millisecond)

	// Cancel the context
	cancel()

	// Give time for listenUpdates to exit
	time.Sleep(50 * time.Millisecond)

	// Verify the stream detects closure (channel should be closed)
	select {
	case <-ss.SessionUpdated():
		// Channel is closed, which is expected
	case <-time.After(100 * time.Millisecond):
		t.Error("expected channel to be closed due to context cancellation")
	}

	mockStream.AssertExpectations(t)
}

func TestSessionStream_ConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{}

	// Setup mocks to allow concurrent operations
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(8, nil)
	mockStream.On("Close").Return(nil)

	ss := newSessionStream(mockStream)

	// Test concurrent access to various methods
	done := make(chan struct{})
	var operations int

	// Concurrent updateSession calls
	go func() {
		for i := range 5 {
			ss.updateSession(uint64(i * 1000))
			time.Sleep(time.Millisecond)
		}
		operations++
		if operations == 2 {
			close(done)
		}
	}()

	// // Concurrent close attempts
	// go func() {
	// 	time.Sleep(10 * time.Millisecond)
	// 	ss.close()
	// 	operations++
	// 	if operations == 2 {
	// 		close(done)
	// 	}
	// }()

	select {
	case <-done:
		// Test should complete without race conditions or panics
	case <-time.After(1 * time.Second):
		t.Error("concurrent operations timed out")
	}

	mockStream.AssertExpectations(t)
}

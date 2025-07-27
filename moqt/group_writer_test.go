package moqt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSendGroupStream(t *testing.T) {
	tests := map[string]struct {
		setupMock func() *MockQUICSendStream
		sequence  GroupSequence
	}{
		"valid stream and sequence": {
			setupMock: func() *MockQUICSendStream {
				mockStream := &MockQUICSendStream{}
				return mockStream
			},
			sequence: GroupSequence(123),
		},
		"different sequence": {
			setupMock: func() *MockQUICSendStream {
				mockStream := &MockQUICSendStream{}
				return mockStream
			},
			sequence: GroupSequence(456),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.setupMock()
			onClose := func() {}

			sgs := newGroupWriter(mockStream, tt.sequence, onClose)

			assert.NotNil(t, sgs)
			assert.Equal(t, tt.sequence, sgs.sequence)
			assert.Equal(t, uint64(0), sgs.frameCount)
			assert.NotNil(t, sgs.ctx)
			assert.Equal(t, mockStream, sgs.stream)
			assert.NotNil(t, sgs.onClose)
		})
	}
}

func TestSendGroupStream_GroupSequence(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	sequence := GroupSequence(789)
	sgs := newGroupWriter(mockStream, sequence, func() {})

	result := sgs.GroupSequence()
	assert.Equal(t, sequence, result)
}

func TestSendGroupStream_WriteFrame(t *testing.T) {
	tests := map[string]struct {
		frame       *Frame
		expectError bool
		setupMock   func() *MockQUICSendStream
	}{
		"valid frame": {
			frame: &Frame{message: &message.FrameMessage{Payload: []byte("test")}},
			setupMock: func() *MockQUICSendStream {
				mockStream := &MockQUICSendStream{}
				mockStream.On("Write", mock.Anything).Return(4, nil)
				return mockStream
			},
			expectError: false,
		},
		"nil frame": {
			frame: nil,
			setupMock: func() *MockQUICSendStream {
				return &MockQUICSendStream{}
			},
			expectError: true,
		},
		"frame with nil message": {
			frame: &Frame{message: nil},
			setupMock: func() *MockQUICSendStream {
				return &MockQUICSendStream{}
			},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.setupMock()
			sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

			err := sgs.WriteFrame(tt.frame)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, uint64(1), sgs.frameCount)
				mockStream.AssertExpectations(t)
			}
		})
	}
}

func TestSendGroupStream_SetWriteDeadline(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	deadline := time.Now().Add(time.Minute)
	mockStream.On("SetWriteDeadline", deadline).Return(nil)
	sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

	err := sgs.SetWriteDeadline(deadline)
	assert.NoError(t, err)
	mockStream.AssertExpectations(t)
}

func TestSendGroupStream_Close(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	mockStream.On("Close").Return(nil)

	sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

	err := sgs.Close()
	assert.NoError(t, err)
	mockStream.AssertExpectations(t)

	// Verify the context is cancelled after Close()
	assert.Error(t, sgs.ctx.Err(), "context should be cancelled after Close()")
}

func TestSendGroupStream_CloseWithError(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	errorCode := GroupErrorCode(42)
	streamID := quic.StreamID(123)

	mockStream.On("CancelWrite", quic.StreamErrorCode(errorCode)).Return()
	mockStream.On("StreamID").Return(streamID)
	sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

	sgs.CancelWrite(errorCode)
	mockStream.AssertExpectations(t)

	// Verify the context is cancelled after CancelWrite()
	assert.Error(t, sgs.ctx.Err(), "context should be cancelled after CancelWrite()")

	// Verify the cause is a GroupError
	causeErr := context.Cause(sgs.ctx)
	assert.NotNil(t, causeErr, "context cause should not be nil")

	var groupErr *GroupError
	assert.True(t, errors.As(causeErr, &groupErr), "context cause should be a GroupError")
}

func TestSendGroupStream_ContextCancellation(t *testing.T) {
	t.Run("operations fail when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		mockStream := &MockQUICSendStream{}
		mockStream.On("Context").Return(ctx)

		sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

		// Cancel the context
		cancel()

		// Test that operations fail when context is cancelled
		frame := &Frame{message: &message.FrameMessage{Payload: []byte("test")}}
		err := sgs.WriteFrame(frame)
		assert.Error(t, err)
	})
}

func TestSendGroupStream_CloseWithStreamError(t *testing.T) {
	t.Run("close returns stream error", func(t *testing.T) {
		mockStream := &MockQUICSendStream{}
		streamID := quic.StreamID(123)
		streamErr := &quic.StreamError{
			StreamID:  streamID,
			ErrorCode: quic.StreamErrorCode(42),
		}

		mockStream.On("Close").Return(streamErr)

		sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

		err := sgs.Close()
		assert.Error(t, err)

		// Verify it returns a GroupError
		var groupErr *GroupError
		assert.True(t, errors.As(err, &groupErr), "error should be a GroupError")
		assert.Equal(t, streamErr, groupErr.StreamError)

		// Verify the context is cancelled
		assert.Error(t, sgs.ctx.Err(), "context should be cancelled after Close() error")
		mockStream.AssertExpectations(t)
	})

	t.Run("close returns non-stream error", func(t *testing.T) {
		mockStream := &MockQUICSendStream{}
		otherErr := errors.New("some other error")

		mockStream.On("Close").Return(otherErr)

		sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

		err := sgs.Close()
		assert.Error(t, err)
		assert.Equal(t, otherErr, err)

		// Context should not be cancelled for non-stream errors in current implementation
		mockStream.AssertExpectations(t)
	})

	t.Run("close when context already cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		mockStream := &MockQUICSendStream{}
		mockStream.On("Context").Return(ctx)

		sgs := newGroupWriter(mockStream, GroupSequence(1), func() {})

		// Cancel the context first
		cancel()

		err := sgs.Close()
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)

		// Stream.Close should not be called if context is already cancelled
		mockStream.AssertNotCalled(t, "Close")
	})
}

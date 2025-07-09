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
			ctx := context.Background()

			sgs := newSendGroupStream(ctx, mockStream, tt.sequence)

			assert.NotNil(t, sgs)
			assert.Equal(t, tt.sequence, sgs.sequence)
			assert.Equal(t, uint64(0), sgs.frameCount)
			assert.NotNil(t, sgs.ctx)
			assert.NotNil(t, sgs.cancel)
			assert.Equal(t, mockStream, sgs.stream)
		})
	}
}

func TestSendGroupStream_GroupSequence(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	sequence := GroupSequence(789)
	ctx := context.Background()

	sgs := newSendGroupStream(ctx, mockStream, sequence)

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
			ctx := context.Background()
			sgs := newSendGroupStream(ctx, mockStream, GroupSequence(1))

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
	ctx := context.Background()

	sgs := newSendGroupStream(ctx, mockStream, GroupSequence(1))

	err := sgs.SetWriteDeadline(deadline)
	assert.NoError(t, err)
	mockStream.AssertExpectations(t)
}

func TestSendGroupStream_Close(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	mockStream.On("Close").Return(nil)
	ctx := context.Background()

	sgs := newSendGroupStream(ctx, mockStream, GroupSequence(1))

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
	ctx := context.Background()

	sgs := newSendGroupStream(ctx, mockStream, GroupSequence(1))

	err := sgs.CancelWrite(errorCode)
	assert.NoError(t, err)
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
		mockStream := &MockQUICSendStream{}
		ctx, cancel := context.WithCancel(context.Background())

		sgs := newSendGroupStream(ctx, mockStream, GroupSequence(1))

		// Cancel the context
		cancel()

		// Test that operations fail when context is cancelled
		frame := &Frame{message: &message.FrameMessage{Payload: []byte("test")}}
		err := sgs.WriteFrame(frame)
		assert.Error(t, err)
	})
}

package moqt

import (
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

			sgs := newSendGroupStream(mockStream, tt.sequence)

			assert.NotNil(t, sgs)
			assert.Equal(t, mockStream, sgs.stream)
			assert.Equal(t, tt.sequence, sgs.sequence)
			assert.Equal(t, uint64(0), sgs.frameCount)
			assert.False(t, sgs.closed)
		})
	}
}

func TestSendGroupStream_GroupSequence(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	sequence := GroupSequence(789)

	sgs := newSendGroupStream(mockStream, sequence)

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
			sgs := newSendGroupStream(mockStream, GroupSequence(1))

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

	sgs := newSendGroupStream(mockStream, GroupSequence(1))

	err := sgs.SetWriteDeadline(deadline)
	assert.NoError(t, err)
	mockStream.AssertExpectations(t)
}

func TestSendGroupStream_Close(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	mockStream.On("Close").Return(nil)

	sgs := newSendGroupStream(mockStream, GroupSequence(1))

	err := sgs.Close()
	assert.NoError(t, err)
	assert.True(t, sgs.closed)
	mockStream.AssertExpectations(t)

	// Verify the closed channel is closed
	select {
	case <-sgs.closedCh:
		// Good - should be closed
	default:
		t.Error("closedCh should be closed after Close()")
	}
}

func TestSendGroupStream_CloseWithError(t *testing.T) {
	mockStream := &MockQUICSendStream{}
	errorCode := GroupErrorCode(42)
	streamID := quic.StreamID(123)

	mockStream.On("CancelWrite", quic.StreamErrorCode(errorCode)).Return()
	mockStream.On("StreamID").Return(streamID)

	sgs := newSendGroupStream(mockStream, GroupSequence(1))

	err := sgs.CancelWrite(errorCode)
	assert.NoError(t, err)
	assert.True(t, sgs.closed)
	assert.NotNil(t, sgs.closeErr)
	mockStream.AssertExpectations(t)

	// Verify the closed channel is closed
	select {
	case <-sgs.closedCh:
		// Good - should be closed
	default:
		t.Error("closedCh should be closed after CloseWithError()")
	}
}

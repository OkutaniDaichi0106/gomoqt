package moqt

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewReceiveGroupStream(t *testing.T) {
	tests := map[string]struct {
		sequence    GroupSequence
		expectValid bool
	}{
		"valid creation": {
			sequence:    GroupSequence(123),
			expectValid: true,
		},
		"zero sequence": {
			sequence:    GroupSequence(0),
			expectValid: true,
		},
		"large sequence": {
			sequence:    GroupSequence(4294967295),
			expectValid: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICReceiveStream{}

			rgs := newReceiveGroupStream(context.Background(), tt.sequence, mockStream)

			assert.NotNil(t, rgs)
			assert.Equal(t, tt.sequence, rgs.sequence)
			assert.Equal(t, mockStream, rgs.stream)
			assert.Equal(t, int64(0), rgs.frameCount)
			assert.NotNil(t, rgs.ctx)
		})
	}
}

func TestReceiveGroupStream_GroupSequence(t *testing.T) {
	tests := map[string]struct {
		sequence GroupSequence
	}{
		"minimum value": {
			sequence: GroupSequence(0),
		},
		"small value": {
			sequence: GroupSequence(1),
		},
		"medium value": {
			sequence: GroupSequence(1000),
		},
		"large value": {
			sequence: GroupSequence(1000000),
		},
		"maximum uint64": {
			sequence: GroupSequence(^uint64(0)),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICReceiveStream{}
			rgs := newReceiveGroupStream(context.Background(), tt.sequence, mockStream)

			result := rgs.GroupSequence()
			assert.Equal(t, tt.sequence, result)
		})
	}
}

func TestReceiveGroupStream_ReadFrame_EOF(t *testing.T) {
	mockStream := &MockQUICReceiveStream{}
	buf := bytes.NewBuffer(nil) // Empty buffer will return EOF
	mockStream.ReadFunc = buf.Read

	rgs := newReceiveGroupStream(context.Background(), GroupSequence(123), mockStream)

	frame, err := rgs.ReadFrame()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, frame)
}

func TestReceiveGroupStream_CancelRead(t *testing.T) {
	tests := map[string]struct {
		errorCode GroupErrorCode
	}{
		"internal group error": {
			errorCode: InternalGroupErrorCode,
		},
		"out of range error": {
			errorCode: OutOfRangeErrorCode,
		},
		"expired group error": {
			errorCode: ExpiredGroupErrorCode,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICReceiveStream{}
			mockStream.On("StreamID").Return(quic.StreamID(123))
			mockStream.On("CancelRead", quic.StreamErrorCode(tt.errorCode)).Return()

			rgs := newReceiveGroupStream(context.Background(), GroupSequence(123), mockStream)

			rgs.CancelRead(tt.errorCode)

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveGroupStream_CancelRead_MultipleCalls(t *testing.T) {
	mockStream := &MockQUICReceiveStream{}
	mockStream.On("StreamID").Return(quic.StreamID(123))
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()

	rgs := newReceiveGroupStream(context.Background(), GroupSequence(123), mockStream)

	// Cancel multiple times
	rgs.CancelRead(InternalGroupErrorCode)
	rgs.CancelRead(OutOfRangeErrorCode)

	// Should be called for each CancelRead invocation
	mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(InternalGroupErrorCode))
	mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(OutOfRangeErrorCode))
	mockStream.AssertExpectations(t)
}

func TestReceiveGroupStream_SetReadDeadline(t *testing.T) {
	tests := map[string]struct {
		setupMock func() *MockQUICReceiveStream
		deadline  time.Time
		wantErr   bool
	}{
		"successful set deadline": {
			setupMock: func() *MockQUICReceiveStream {
				mockStream := &MockQUICReceiveStream{}
				mockStream.On("SetReadDeadline", mock.Anything).Return(nil)
				return mockStream
			},
			deadline: time.Now().Add(time.Hour),
			wantErr:  false,
		},
		"set deadline with error": {
			setupMock: func() *MockQUICReceiveStream {
				mockStream := &MockQUICReceiveStream{}
				mockStream.On("SetReadDeadline", mock.Anything).Return(assert.AnError)
				return mockStream
			},
			deadline: time.Now().Add(time.Hour),
			wantErr:  true,
		},
		"zero time deadline": {
			setupMock: func() *MockQUICReceiveStream {
				mockStream := &MockQUICReceiveStream{}
				mockStream.On("SetReadDeadline", time.Time{}).Return(nil)
				return mockStream
			},
			deadline: time.Time{},
			wantErr:  false,
		},
		"deadline in the past": {
			setupMock: func() *MockQUICReceiveStream {
				mockStream := &MockQUICReceiveStream{}
				mockStream.On("SetReadDeadline", mock.Anything).Return(nil)
				return mockStream
			},
			deadline: time.Now().Add(-time.Hour),
			wantErr:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.setupMock()
			rgs := newReceiveGroupStream(context.Background(), GroupSequence(123), mockStream)

			err := rgs.SetReadDeadline(tt.deadline)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveGroupStream_ReadFrame_StreamError(t *testing.T) {
	mockStream := &MockQUICReceiveStream{}
	mockStream.On("StreamID").Return(quic.StreamID(123))

	// Simulate a stream error
	streamErr := &quic.StreamError{
		StreamID:  quic.StreamID(123),
		ErrorCode: quic.StreamErrorCode(1),
	}

	mockStream.ReadFunc = func(p []byte) (int, error) {
		return 0, streamErr
	}

	rgs := newReceiveGroupStream(context.Background(), GroupSequence(123), mockStream)

	frame, err := rgs.ReadFrame()
	assert.Error(t, err)
	assert.Nil(t, frame)

	// Should be a GroupError
	var groupErr *GroupError
	assert.True(t, errors.As(err, &groupErr))
}

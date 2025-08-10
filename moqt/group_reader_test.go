package moqt

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
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

			rgs := newGroupReader(tt.sequence, mockStream, func() {})

			assert.NotNil(t, rgs)
			assert.Equal(t, tt.sequence, rgs.sequence)
			assert.Equal(t, mockStream, rgs.stream)
			assert.Equal(t, int64(0), rgs.frameCount)
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
			rgs := newGroupReader(tt.sequence, mockStream, func() {})

			result := rgs.GroupSequence()
			assert.Equal(t, tt.sequence, result)
		})
	}
}

func TestReceiveGroupStream_ReadFrame_EOF(t *testing.T) {
	mockStream := &MockQUICReceiveStream{}
	buf := bytes.NewBuffer(nil) // Empty buffer will return EOF
	mockStream.ReadFunc = buf.Read

	rgs := newGroupReader(GroupSequence(123), mockStream, func() {})

	frame := NewFrame(nil)
	err := rgs.ReadFrame(frame)
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	// Note: frame is modified even on error, so we don't assert it's nil
	assert.NotNil(t, frame) // Frame is modified with internal message
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
			mockStream.On("CancelRead", quic.StreamErrorCode(tt.errorCode)).Return()

			rgs := newGroupReader(GroupSequence(123), mockStream, func() {})

			rgs.CancelRead(tt.errorCode)

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveGroupStream_CancelRead_MultipleCalls(t *testing.T) {
	mockStream := &MockQUICReceiveStream{}
	mockStream.On("CancelRead", quic.StreamErrorCode(InternalGroupErrorCode)).Return()

	rgs := newGroupReader(GroupSequence(123), mockStream, func() {})

	// Cancel multiple times with the same error code
	rgs.CancelRead(InternalGroupErrorCode)
	rgs.CancelRead(InternalGroupErrorCode)

	// Should be called for each CancelRead invocation
	mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(InternalGroupErrorCode))
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
			rgs := newGroupReader(123, mockStream, func() {})

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
	mockStream := &MockQUICReceiveStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, &quic.StreamError{
				StreamID:  quic.StreamID(123),
				ErrorCode: quic.StreamErrorCode(1),
			}
		},
	}
	mockStream.On("StreamID").Return(quic.StreamID(123))

	rgs := newGroupReader(123, mockStream, func() {})

	frame := NewFrame(nil)
	err := rgs.ReadFrame(frame)
	assert.Error(t, err)
	// Note: frame is modified even on error, so we don't assert it's nil
	assert.NotNil(t, frame) // Frame is modified with internal message

	// Should be a GroupError
	var groupErr *GroupError
	assert.True(t, errors.As(err, &groupErr))
}

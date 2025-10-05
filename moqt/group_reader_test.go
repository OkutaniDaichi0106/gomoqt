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
			sequence: GroupSequence(1<<(64-2) - 1), // maxVarInt8
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
	frame, err := rgs.ReadFrame()
	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
	// On EOF, frame should be nil
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
	frame, err := rgs.ReadFrame()
	assert.Error(t, err)
	// On stream error, frame should be nil
	assert.Nil(t, frame)

	// Should be a GroupError
	var groupErr *GroupError
	assert.True(t, errors.As(err, &groupErr))
}

func TestGroupReader_ReadFrame(t *testing.T) {
	tests := map[string]struct {
		setupStream func() *MockQUICReceiveStream
		expectError bool
		expectFrame bool
	}{
		"successful read": {
			setupStream: func() *MockQUICReceiveStream {
				// Create a frame with some data
				builder := NewFrameBuilder(10)
				builder.Append([]byte("test data"))
				frame := builder.Frame()
				var buf bytes.Buffer
				err := frame.encode(&buf)
				if err != nil {
					panic(err)
				}
				data := buf.Bytes()

				mockStream := &MockQUICReceiveStream{
					ReadFunc: func(p []byte) (int, error) {
						if len(data) == 0 {
							return 0, io.EOF
						}
						n := copy(p, data)
						data = data[n:]
						return n, nil
					},
				}
				return mockStream
			},
			expectError: false,
			expectFrame: true,
		},
		"EOF": {
			setupStream: func() *MockQUICReceiveStream {
				mockStream := &MockQUICReceiveStream{
					ReadFunc: func(p []byte) (int, error) {
						return 0, io.EOF
					},
				}
				return mockStream
			},
			expectError: true,
			expectFrame: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.setupStream()
			rgs := newGroupReader(123, mockStream, func() {})

			frame, err := rgs.ReadFrame()
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectFrame {
					assert.NotNil(t, frame)
				} else {
					assert.Nil(t, frame)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, frame)
			}
		})
	}
}

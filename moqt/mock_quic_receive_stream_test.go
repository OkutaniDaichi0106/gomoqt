package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.ReceiveStream = (*MockQUICReceiveStream)(nil)

func newMockQUICReceiveStream(streamID quic.StreamID) *MockQUICReceiveStream {
	mockStream := &MockQUICReceiveStream{}
	mockStream.On("StreamID").Return(streamID)
	return mockStream
}

// MockQUICReceiveStream is a mock implementation of quic.ReceiveStream using testify/mock
type MockQUICReceiveStream struct {
	mock.Mock
}

func (m *MockQUICReceiveStream) StreamID() quic.StreamID {
	args := m.Called()
	return args.Get(0).(quic.StreamID)
}

func (m *MockQUICReceiveStream) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICReceiveStream) CancelRead(code quic.StreamErrorCode) {
	m.Called(code)
}

func (m *MockQUICReceiveStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

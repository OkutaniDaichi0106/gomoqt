package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.ReceiveStream = (*MockQUICReceiveStream)(nil)

// MockQUICReceiveStream is a mock implementation of quic.ReceiveStream using testify/mock
type MockQUICReceiveStream struct {
	mock.Mock
	ReadFunc func(p []byte) (n int, err error)
}

func (m *MockQUICReceiveStream) StreamID() quic.StreamID {
	args := m.Called()
	return args.Get(0).(quic.StreamID)
}

func (m *MockQUICReceiveStream) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
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

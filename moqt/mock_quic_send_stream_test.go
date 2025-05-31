package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.SendStream = (*MockQUICSendStream)(nil)

// MockQUICSendStream is a mock implementation of quic.SendStream using testify/mock
type MockQUICSendStream struct {
	mock.Mock
	StreamIDValue quic.StreamID
	Cancelled     bool
	CancelCode    quic.StreamErrorCode
	Deadline      time.Time
	WriteFunc     func(p []byte) (int, error)
}

func (m *MockQUICSendStream) StreamID() quic.StreamID {
	if args := m.Called(); args.Get(0) != nil {
		return args.Get(0).(quic.StreamID)
	}
	return m.StreamIDValue
}

func (m *MockQUICSendStream) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICSendStream) CancelWrite(code quic.StreamErrorCode) {
	m.Called(code)
	m.Cancelled = true
	m.CancelCode = code
}

func (m *MockQUICSendStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	m.Deadline = t
	return args.Error(0)
}

func (m *MockQUICSendStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

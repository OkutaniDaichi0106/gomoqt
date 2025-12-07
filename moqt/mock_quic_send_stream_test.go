package moqt

import (
	"context"
	"time"

	"github.com/okdaichi/gomoqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.SendStream = (*MockQUICSendStream)(nil)

// MockQUICSendStream is a mock implementation of quic.SendStream using testify/mock
type MockQUICSendStream struct {
	mock.Mock
	WriteFunc func(p []byte) (n int, err error)
}

func (m *MockQUICSendStream) StreamID() quic.StreamID {
	// Recover from testify/mock panic when no expectation is provided and
	// default to zero StreamID which is safe for logging purposes.
	defer func() {
		if r := recover(); r != nil {
			_ = r // ignore the panic and return zero value below
		}
	}()
	args := m.Called()
	if len(args) == 0 || args.Get(0) == nil {
		return quic.StreamID(0)
	}
	return args.Get(0).(quic.StreamID)
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
}

func (m *MockQUICSendStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICSendStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockQUICSendStream) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

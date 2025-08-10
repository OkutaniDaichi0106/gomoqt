package moqt

import (
	"context"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.Stream = (*MockQUICStream)(nil)

// MockQUICStream is a mock implementation of quic.Stream using testify/mock
type MockQUICStream struct {
	mock.Mock
	ReadFunc  func(p []byte) (n int, err error)
	WriteFunc func(p []byte) (n int, err error)
}

func (m *MockQUICStream) StreamID() quic.StreamID {
	args := m.Called()
	return args.Get(0).(quic.StreamID)
}

func (m *MockQUICStream) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICStream) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockQUICStream) CancelRead(code quic.StreamErrorCode) {
	m.Called(code)
}

func (m *MockQUICStream) CancelWrite(code quic.StreamErrorCode) {
	m.Called(code)
}

func (m *MockQUICStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockQUICStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockQUICStream) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

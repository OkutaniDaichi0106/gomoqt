package moqt

import (
	"context"
	"sync"
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
	// internal cancellable context returned by Context()
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
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
	// Cancel the context to simulate stream cancellation
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.mu.Unlock()
}

func (m *MockQUICStream) CancelWrite(code quic.StreamErrorCode) {
	m.Called(code)
	// Cancel the context to simulate stream cancellation
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.mu.Unlock()
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
	// Cancel the context to simulate stream close
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.mu.Unlock()
	return args.Error(0)
}

func (m *MockQUICStream) Context() context.Context {
	args := m.Called()
	// If caller (test) provided a parent context via mock return, wrap it
	// so we can cancel a derived context when Close/Cancel* is called.
	var parent context.Context
	if args.Get(0) != nil {
		parent = args.Get(0).(context.Context)
	} else {
		parent = context.Background()
	}

	m.mu.Lock()
	if m.ctx == nil {
		m.ctx, m.cancel = context.WithCancel(parent)
	}
	ctx := m.ctx
	m.mu.Unlock()
	return ctx
}

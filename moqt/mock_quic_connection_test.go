package moqt

import (
	"context"
	"net"

	"github.com/okdaichi/gomoqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.Connection = (*MockQUICConnection)(nil)

// MockQUICConnection is a mock implementation of quic.Connection using testify/mock
type MockQUICConnection struct {
	mock.Mock
	AcceptStreamFunc      func(ctx context.Context) (quic.Stream, error)
	AcceptUniStreamFunc   func(ctx context.Context) (quic.ReceiveStream, error)
	OpenStreamFunc        func() (quic.Stream, error)
	OpenUniStreamFunc     func() (quic.SendStream, error)
	OpenStreamSyncFunc    func(ctx context.Context) (quic.Stream, error)
	OpenUniStreamSyncFunc func(ctx context.Context) (quic.SendStream, error)
}

func (m *MockQUICConnection) AcceptStream(ctx context.Context) (quic.Stream, error) {
	if m.AcceptStreamFunc != nil {
		return m.AcceptStreamFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (m *MockQUICConnection) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	if m.AcceptUniStreamFunc != nil {
		return m.AcceptUniStreamFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.ReceiveStream), args.Error(1)
}

func (m *MockQUICConnection) OpenStream() (quic.Stream, error) {
	if m.OpenStreamFunc != nil {
		return m.OpenStreamFunc()
	}
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (m *MockQUICConnection) OpenUniStream() (quic.SendStream, error) {
	if m.OpenUniStreamFunc != nil {
		return m.OpenUniStreamFunc()
	}
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.SendStream), args.Error(1)
}

func (m *MockQUICConnection) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	if m.OpenStreamSyncFunc != nil {
		return m.OpenStreamSyncFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (m *MockQUICConnection) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	if m.OpenUniStreamSyncFunc != nil {
		return m.OpenUniStreamSyncFunc(ctx)
	}
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.SendStream), args.Error(1)
}

func (m *MockQUICConnection) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockQUICConnection) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockQUICConnection) CloseWithError(code quic.ApplicationErrorCode, reason string) error {
	args := m.Called(code, reason)
	return args.Error(0)
}

func (m *MockQUICConnection) ConnectionState() quic.ConnectionState {
	args := m.Called()
	return args.Get(0).(quic.ConnectionState)
}

func (m *MockQUICConnection) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

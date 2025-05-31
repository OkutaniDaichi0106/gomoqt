package moqt

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/mock"
)

var _ quic.Connection = (*MockQUICConnection)(nil)

// MockQUICConnection is a mock implementation of quic.Connection using testify/mock
type MockQUICConnection struct {
	mock.Mock
}

func (m *MockQUICConnection) AcceptStream(ctx context.Context) (quic.Stream, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (m *MockQUICConnection) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.ReceiveStream), args.Error(1)
}

func (m *MockQUICConnection) OpenStream() (quic.Stream, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (m *MockQUICConnection) OpenUniStream() (quic.SendStream, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.SendStream), args.Error(1)
}

func (m *MockQUICConnection) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (m *MockQUICConnection) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
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

func (m *MockQUICConnection) CloseWithError(code quic.ConnectionErrorCode, reason string) error {
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

package moqt

import (
	"context"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
	"github.com/stretchr/testify/mock"
)

var _ webtransport.Server = (*MockWebTransportServer)(nil)

// MockWebTransportServer is a mock implementation of the webtransport.Server interface
type MockWebTransportServer struct {
	mock.Mock
}

// ServeHTTP mocks the ServeHTTP method
func (m *MockWebTransportServer) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	args := m.Called(w, r)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockWebTransportServer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Serve mocks the Serve method for net.Listener (from webtransport.Server)
func (m *MockWebTransportServer) Serve(conn net.PacketConn) error {
	args := m.Called(conn)
	return args.Error(0)
}

// Upgrade mocks the Upgrade method
func (m *MockWebTransportServer) Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error) {
	args := m.Called(w, r)
	if conn, ok := args.Get(0).(quic.Connection); ok {
		return conn, args.Error(1)
	}
	return nil, args.Error(1)
}

// ServeQUICConn mocks the ServeQUICConn method
func (m *MockWebTransportServer) ServeQUICConn(conn quic.Connection) error {
	args := m.Called(conn)
	return args.Error(0)
}

// ServePacketConn mocks the Serve method for net.PacketConn
func (m *MockWebTransportServer) ServePacketConn(conn net.PacketConn) error {
	args := m.Called(conn)
	return args.Error(0)
}

// Shutdown mocks the Shutdown method
func (m *MockWebTransportServer) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

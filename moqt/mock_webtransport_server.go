package moqt

import (
	"context"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport"
	"github.com/stretchr/testify/mock"
)

var _ webtransport.Server = (*MockWebTransportServer)(nil)

// MockWebTransportServer is a mock implementation of the webtransport.Server interface
type MockWebTransportServer struct {
	mock.Mock
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

// Close mocks the Close method
func (m *MockWebTransportServer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Shutdown mocks the Shutdown method
func (m *MockWebTransportServer) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

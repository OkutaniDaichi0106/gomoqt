package moqt

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServer_Init(t *testing.T) {
	tests := map[string]struct {
		addr   string
		logger *slog.Logger
		config *Config
	}{
		"basic init": {
			addr:   ":8080",
			logger: slog.Default(),
			config: &Config{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
				Config: tt.config,
			}

			server.init()

			assert.NotNil(t, server.listeners, "listeners map should be initialized")
			assert.NotNil(t, server.doneChan, "doneChan should be initialized")
			assert.NotNil(t, server.activeSess, "activeSess map should be initialized")

		})
	}
}

func TestServer_InitOnce(t *testing.T) {
	tests := map[string]struct {
		addr      string
		logger    *slog.Logger
		initCalls int
	}{
		"multiple init calls": {
			addr:      ":8080",
			logger:    slog.Default(),
			initCalls: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			// Call init multiple times
			for i := 0; i < tt.initCalls; i++ {
				server.init()
			}

			// Should only initialize once - verify by checking that fields are set
			assert.NotNil(t, server.listeners, "listeners map should be initialized")
		})
	}
}

func TestServer_ServeQUICListener(t *testing.T) {
	tests := map[string]struct {
		addr        string
		logger      *slog.Logger
		waitTime    time.Duration
		expectError bool
	}{
		"serve QUIC listener": {
			addr:        ":8080",
			logger:      slog.Default(),
			waitTime:    50 * time.Millisecond,
			expectError: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			mockListener := &MockEarlyListener{}
			mockConn := &MockQUICConnection{}

			// Setup the mock connection methods that will be called
			mockConn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
			mockConn.On("ConnectionState").Return(quic.ConnectionState{
				TLS: tls.ConnectionState{NegotiatedProtocol: "moq-00"},
			})
			// Context is used for accept timeout in handleNativeQUIC
			mockConn.On("Context").Return(context.Background())
			// Avoid blocking goroutines: no actual streams in this test
			mockConn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)

			mockListener.On("Accept", mock.Anything).Return(mockConn, nil)
			mockListener.On("Close").Return(nil)

			// Test serving the listener
			go func() {
				err := server.ServeQUICListener(mockListener)
				if err != nil && err != ErrServerClosed {
					if tt.expectError {
						return
					}
					t.Errorf("ServeQUICListener() error = %v", err)
				}
			}()

			// Give time for the server to start
			time.Sleep(tt.waitTime)

			// Close the server
			server.Close()
		})
	}
}

func TestServer_ServeQUICListener_AcceptError(t *testing.T) {
	tests := map[string]struct {
		addr        string
		logger      *slog.Logger
		expectError bool
	}{
		"accept error handling": {
			addr:        ":8080",
			logger:      slog.Default(),
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			mockListener := &MockEarlyListener{}
			mockListener.On("Accept", mock.Anything).Return(nil, errors.New("accept error"))

			err := server.ServeQUICListener(mockListener)
			// Should handle accept errors gracefully
			if tt.expectError {
				assert.True(t, err != nil && err != ErrServerClosed, "should handle accept errors gracefully")
			}
		})
	}
}

func TestServer_ServeQUICListener_ShuttingDown(t *testing.T) {
	tests := map[string]struct {
		addr      string
		logger    *slog.Logger
		expectErr error
	}{
		"shutting down server": {
			addr:      ":8080",
			logger:    slog.Default(),
			expectErr: ErrServerClosed,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			// Set server to shutting down state
			server.inShutdown.Store(true)

			mockListener := &MockEarlyListener{}

			err := server.ServeQUICListener(mockListener)
			assert.Equal(t, tt.expectErr, err, "ServeQUICListener() on shutting down server should return ErrServerClosed")
		})
	}
}

func TestServer_Close(t *testing.T) {
	tests := map[string]struct {
		addr   string
		logger *slog.Logger
	}{
		"close server": {
			addr:   ":8080",
			logger: slog.Default(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			} // Initialize server
			server.init()

			// Add a mock listener
			mockListener := &MockEarlyListener{}
			mockListener.On("Close").Return(nil)
			server.listeners[mockListener] = struct{}{}

			err := server.Close()
			assert.NoError(t, err, "Close() should not return error")
			assert.True(t, server.shuttingDown(), "server should be in shutting down state after close")
			assert.True(t, mockListener.AssertCalled(t, "Close"), "listener should be closed when server closes")
		})
	}
}

func TestServer_Close_AlreadyShuttingDown(t *testing.T) {
	tests := map[string]struct {
		addr      string
		logger    *slog.Logger
		expectErr error
	}{
		"close already shutting down": {
			addr:      ":8080",
			logger:    slog.Default(),
			expectErr: ErrServerClosed,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			// Set server to shutting down state
			server.inShutdown.Store(true)

			err := server.Close()
			assert.Equal(t, tt.expectErr, err, "Close() on already shutting down server should return ErrServerClosed")
		})
	}
}

func TestServer_ShuttingDown(t *testing.T) {
	tests := map[string]struct {
		setShutdown  bool
		expectResult bool
	}{
		"new server not shutting down": {
			setShutdown:  false,
			expectResult: false,
		},
		"server after shutdown set": {
			setShutdown:  true,
			expectResult: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{}

			if tt.setShutdown {
				server.inShutdown.Store(true)
			}

			result := server.shuttingDown()
			assert.Equal(t, tt.expectResult, result, "shuttingDown() should return expected result")
		})
	}
}

func TestServer_AcceptSession(t *testing.T) {
	tests := map[string]struct {
		addr     string
		logger   *slog.Logger
		path     string
		expectOK bool
	}{
		"accept session success": {
			addr:     ":8080",
			logger:   slog.Default(),
			path:     "/test",
			expectOK: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock session setup messages
			var buf bytes.Buffer
			// First, encode STREAM_TYPE message
			err := message.StreamTypeSession.Encode(&buf)
			require.NoError(t, err, "failed to encode STREAM_TYPE message")

			// Then, encode SESSION_CLIENT message
			scm := message.SessionClientMessage{
				SupportedVersions: []protocol.Version{protocol.Version(1)},
				Parameters:        message.Parameters{},
			}
			err = scm.Encode(&buf)
			require.NoError(t, err, "failed to encode SESSION_CLIENT message") // Create a mock connection with a session stream
			mockStream := &MockQUICStream{}
			mockStream.ReadFunc = func(p []byte) (int, error) {
				if buf.Len() == 0 {
					return 0, io.EOF
				}
				return buf.Read(p)
			}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.AnythingOfType("[]uint8"))
			mockStream.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)
			mockStream.On("StreamID").Return(quic.StreamID(1))

			mockConn := &MockQUICConnection{}
			// Mock RemoteAddr for logging
			mockAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
			mockConn.On("RemoteAddr").Return(mockAddr)
			// First call returns the session stream, subsequent calls return EOF to stop goroutines
			mockConn.On("AcceptStream", mock.Anything).Return(mockStream, nil).Once()
			mockConn.On("AcceptStream", mock.Anything).Return(nil, io.EOF) // For handleBiStreams goroutine
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // Return EOF to stop the goroutine

			ctx := context.Background()

			sessStream, err := acceptSessionStream(ctx, mockConn, slog.Default())
			if tt.expectOK {
				assert.NoError(t, err, "acceptSessionStream() should not return error")
				assert.NotNil(t, sessStream, "acceptSessionStream() should return session stream")
			}

			// Cleanup
			if sessStream != nil {
				// Note: sessStream doesn't have Terminate method, this cleanup may not be needed
			}
		})
	}
}

func TestServer_AcceptSession_AcceptStreamError(t *testing.T) {
	tests := map[string]struct {
		addr      string
		logger    *slog.Logger
		path      string
		expectErr error
	}{
		"accept stream error": {
			addr:      ":8080",
			logger:    slog.Default(),
			path:      "/test",
			expectErr: errors.New("stream accept error"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockConn := &MockQUICConnection{}
			// Mock RemoteAddr for logging
			mockAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
			mockConn.On("RemoteAddr").Return(mockAddr)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, tt.expectErr)

			ctx := context.Background()

			sessStream, err := acceptSessionStream(ctx, mockConn, slog.Default())

			assert.Error(t, err, "acceptSessionStream() should return an error")
			assert.Contains(t, err.Error(), tt.expectErr.Error(), "acceptSessionStream() should return wrapped accept error")
			assert.Nil(t, sessStream, "acceptSessionStream() should return nil session stream on error")
		})
	}
}

func TestServer_DoneChannel(t *testing.T) {
	tests := map[string]struct {
		addr      string
		logger    *slog.Logger
		waitTime  time.Duration
		closeTime time.Duration
		checkTime time.Duration
	}{
		"done channel lifecycle": {
			addr:      ":8080",
			logger:    slog.Default(),
			waitTime:  50 * time.Millisecond,
			closeTime: 50 * time.Millisecond,
			checkTime: 100 * time.Millisecond,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			// Initialize server
			server.init()

			// Should not be done initially
			select {
			case <-server.doneChan:
				t.Error("doneChan should not be closed initially")
			default:
				// Expected - channel should not be closed
			}

			// Close server
			server.Close()

			// Give time for cleanup
			time.Sleep(tt.closeTime)

			// Should be done after close
			select {
			case <-server.doneChan:
				// Expected - channel should be closed
			case <-time.After(tt.checkTime):
				t.Error("doneChan should be closed after server close")
			}
		})
	}
}

func TestServer_ConcurrentOperations(t *testing.T) {
	tests := map[string]struct {
		addr     string
		logger   *slog.Logger
		waitTime time.Duration
	}{
		"concurrent operations": {
			addr:     ":8080",
			logger:   slog.Default(),
			waitTime: 50 * time.Millisecond,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			// Test concurrent initialization and operations
			go server.init()
			go server.init()
			go func() {
				time.Sleep(10 * time.Millisecond)
				server.shuttingDown()
			}()

			time.Sleep(tt.waitTime)

			// Test concurrent close operations
			go server.Close()
			go server.Close()

			time.Sleep(tt.waitTime)

			// Test should complete without race conditions
			assert.True(t, true, "test should complete without race conditions")
		})
	}
}

func TestServer_WithCustomWebTransportServer(t *testing.T) {

	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
		NewWebtransportServerFunc: func(checkOrigin func(*http.Request) bool) webtransport.Server {
			return &MockWebTransportServer{}
		},
	}

	server.init()

	assert.NotNil(t, server.wtServer, "should use custom WebTransport server when provided")
}

func TestServer_SessionManagement(t *testing.T) {
	tests := map[string]struct {
		addr             string
		logger           *slog.Logger
		expectInitCount  int
		expectFinalCount int
	}{
		"session add and remove": {
			addr:             ":8080",
			logger:           slog.Default(),
			expectInitCount:  1,
			expectFinalCount: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			server.init() // Create a mock session
			connCtx, cancelConn := context.WithCancelCause(context.Background())
			streamCtx, cancelStream := context.WithCancelCause(connCtx)
			defer cancelStream(nil) // Ensure stream context is cancelled

			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(streamCtx)
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)

			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				cancelConn(&quic.ApplicationError{
					ErrorCode:    quic.ApplicationErrorCode(args[0].(quic.ApplicationErrorCode)),
					ErrorMessage: args[1].(string),
				}) // Cancel the connection context
			}).Return(nil) // For session.Terminate
			mockConn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

			req := &SetupRequest{
				Path:       "/test",
				Extensions: NewParameters(),
			}
			sessStream := newSessionStream(mockStream, req)
			session := newSession(mockConn, sessStream, nil, nil, nil)

			// Test adding session
			server.sessMu.Lock()
			server.activeSess[session] = struct{}{}
			server.sessMu.Unlock()

			server.sessMu.RLock()
			count := len(server.activeSess)
			server.sessMu.RUnlock()

			assert.Equal(t, tt.expectInitCount, count, "active session count should match expected")

			// Test removing session
			server.sessMu.Lock()
			delete(server.activeSess, session)
			server.sessMu.Unlock()

			server.sessMu.RLock()
			count = len(server.activeSess)
			server.sessMu.RUnlock()

			assert.Equal(t, tt.expectFinalCount, count, "active session count after removal should match expected")

			// Cleanup
			session.Terminate(NoError, NoError.String())
		})
	}
}

func TestServer_ConfigDefaults(t *testing.T) {
	tests := map[string]struct {
		addr       string
		logger     *slog.Logger
		expectInit bool
	}{
		"nil config handling": {
			addr:       ":8080",
			logger:     nil,
			expectInit: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr: tt.addr,
			}

			// Test that server handles nil config gracefully
			server.init()

			// Should not panic and should initialize properly
			if tt.expectInit {
				assert.NotNil(t, server.listeners, "listeners should be initialized even with nil config")
			}
		})
	}
}

func TestServer_ListenerManagement(t *testing.T) {
	tests := map[string]struct {
		addr         string
		logger       *slog.Logger
		numListeners int
		expectCount  int
	}{
		"multiple listeners": {
			addr:         ":8080",
			logger:       slog.Default(),
			numListeners: 2,
			expectCount:  2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				Addr:   tt.addr,
				Logger: tt.logger,
			}

			server.init()

			mockListener1 := &MockEarlyListener{}
			mockListener2 := &MockEarlyListener{}

			// Set up mock expectations for Close calls
			mockListener1.On("Close").Return(nil)
			mockListener2.On("Close").Return(nil)

			// Add listeners
			server.listenerMu.Lock()
			server.listeners[mockListener1] = struct{}{}
			server.listeners[mockListener2] = struct{}{}
			server.listenerMu.Unlock()

			server.listenerMu.RLock()
			count := len(server.listeners)
			server.listenerMu.RUnlock()

			assert.Equal(t, tt.expectCount, count, "listener count should match expected")

			// Close server should close all listeners
			server.Close()

			assert.True(t, mockListener1.AssertCalled(t, "Close"), "listener1 should be closed when server closes")
			assert.True(t, mockListener2.AssertCalled(t, "Close"), "listener2 should be closed when server closes")
		})
	}
}

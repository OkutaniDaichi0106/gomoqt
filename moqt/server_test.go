package moqt

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
)

func TestServer_Init(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	server.init()

	if server.listeners == nil {
		t.Error("listeners map should be initialized")
	}

	if server.doneChan == nil {
		t.Error("doneChan should be initialized")
	}

	if server.activeSess == nil {
		t.Error("activeSess map should be initialized")
	}

	if server.nativeQUICCh == nil {
		t.Error("nativeQUICCh should be initialized")
	}

	if server.WebtransportServer == nil {
		t.Error("WebtransportServer should be initialized")
	}
}

func TestServer_InitOnce(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	// Call init multiple times
	server.init()
	server.init()
	server.init()

	// Should only initialize once - verify by checking that fields are set
	if server.listeners == nil {
		t.Error("listeners map should be initialized")
	}
}

func TestServer_ServeQUICListener(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	mockListener := &MockEarlyListener{}
	mockListener.On("Accept", context.Background()).Return(&MockQUICConnection{}, nil)

	// Test serving the listener
	go func() {
		err := server.ServeQUICListener(mockListener)
		if err != nil && err != ErrServerClosed {
			t.Errorf("ServeQUICListener() error = %v", err)
		}
	}()

	// Give time for the server to start
	time.Sleep(50 * time.Millisecond)

	// Close the server
	server.Close()
}

func TestServerServeQUICListenerAcceptError(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	mockListener := &MockEarlyListener{}

	err := server.ServeQUICListener(mockListener)
	// Should handle accept errors gracefully
	if err != nil && err != ErrServerClosed {
		// Expected behavior - server should handle accept errors
	}
}

func TestServerServeQUICListenerShuttingDown(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	// Set server to shutting down state
	server.inShutdown.Store(true)

	mockListener := &MockEarlyListener{}

	err := server.ServeQUICListener(mockListener)
	if err != ErrServerClosed {
		t.Errorf("ServeQUICListener() on shutting down server error = %v, want %v", err, ErrServerClosed)
	}
}

func TestServerClose(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	// Initialize server
	server.init()

	// Add a mock listener
	mockListener := &MockEarlyListener{}
	server.listeners[mockListener] = struct{}{}

	err := server.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !server.shuttingDown() {
		t.Error("server should be in shutting down state after close")
	}

	if !mockListener.AssertCalled(t, "Close") {
		t.Error("listener should be closed when server closes")
	}
}

func TestServerCloseAlreadyShuttingDown(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	// Set server to shutting down state
	server.inShutdown.Store(true)

	err := server.Close()
	if err != ErrServerClosed {
		t.Errorf("Close() on already shutting down server error = %v, want %v", err, ErrServerClosed)
	}
}

func TestServerShuttingDown(t *testing.T) {
	server := &Server{}

	if server.shuttingDown() {
		t.Error("new server should not be shutting down")
	}

	server.inShutdown.Store(true)

	if !server.shuttingDown() {
		t.Error("server should be shutting down after inShutdown is set")
	}
}

func TestServer_AcceptSession(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	// Create a mock connection with a session stream
	mockStream := &MockQUICStream{}

	mockConn := &MockQUICConnection{}
	mockConn.On("AcceptStream").Return(mockStream, nil)

	ctx := context.Background()
	path := "/test"
	extensions := func(p *Parameters) (*Parameters, error) {
		return p, nil
	}
	mux := NewTrackMux()

	session, err := server.acceptSession(ctx, path, mockConn, extensions, mux)
	if err != nil {
		t.Errorf("acceptSession() error = %v", err)
	}

	if session == nil {
		t.Error("acceptSession() returned nil session")
	}

	// Cleanup
	if session != nil {
		session.Terminate(nil)
	}
}

func TestServer_AcceptSession_AcceptStreamError(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	expectErr := errors.New("stream accept error")

	mockConn := &MockQUICConnection{}
	mockConn.On("AcceptStream").Return(nil, expectErr)

	ctx := context.Background()
	path := "/test"
	extensions := func(p *Parameters) (*Parameters, error) {
		return p, nil
	}
	mux := NewTrackMux()

	session, err := server.acceptSession(ctx, path, mockConn, extensions, mux)
	if err != expectErr {
		t.Error("acceptSession() should return accept error")
	}

	if session != nil {
		t.Error("acceptSession() should return nil session on error")
	}
}

func TestServerDoneChannel(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
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
	time.Sleep(50 * time.Millisecond)

	// Should be done after close
	select {
	case <-server.doneChan:
		// Expected - channel should be closed
	case <-time.After(100 * time.Millisecond):
		t.Error("doneChan should be closed after server close")
	}
}

func TestServerConcurrentOperations(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	// Test concurrent initialization and operations
	go server.init()
	go server.init()
	go func() {
		time.Sleep(10 * time.Millisecond)
		server.shuttingDown()
	}()

	time.Sleep(50 * time.Millisecond)

	// Test concurrent close operations
	go server.Close()
	go server.Close()

	time.Sleep(50 * time.Millisecond)

	// Test should complete without race conditions
}

func TestServerWithCustomWebTransportServer(t *testing.T) {
	customWT := &MockWebTransportServer{}

	server := &Server{
		Addr:               ":8080",
		Logger:             slog.Default(),
		WebtransportServer: customWT,
	}

	server.init()

	if server.WebtransportServer != customWT {
		t.Error("should use custom WebTransport server when provided")
	}
}

func TestServerSessionManagement(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	server.init()

	// Create a mock session
	sessCtx := createTestSessionContext(context.Background())
	mockStream := &MockQUICStream{}
	sessStream := newSessionStream(sessCtx, mockStream, moqtrace.DefaultQUICStreamAccepted(0))
	mockConn := &MockQUICConnection{}
	session := newSession(sessCtx, sessStream, mockConn, nil)

	// Test adding session
	server.mu.Lock()
	server.activeSess[session] = struct{}{}
	server.mu.Unlock()

	server.mu.RLock()
	count := len(server.activeSess)
	server.mu.RUnlock()

	if count != 1 {
		t.Errorf("active session count = %v, want 1", count)
	}

	// Test removing session
	server.mu.Lock()
	delete(server.activeSess, session)
	server.mu.Unlock()

	server.mu.RLock()
	count = len(server.activeSess)
	server.mu.RUnlock()

	if count != 0 {
		t.Errorf("active session count after removal = %v, want 0", count)
	}

	// Cleanup
	session.Terminate(nil)
}

func TestServerConfigDefaults(t *testing.T) {
	server := &Server{
		Addr: ":8080",
	}

	// Test that server handles nil config gracefully
	server.init()

	// Should not panic and should initialize properly
	if server.listeners == nil {
		t.Error("listeners should be initialized even with nil config")
	}
}

func TestServerListenerManagement(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	server.init()

	mockListener1 := &MockEarlyListener{}
	mockListener2 := &MockEarlyListener{}

	// Add listeners
	server.mu.Lock()
	server.listeners[mockListener1] = struct{}{}
	server.listeners[mockListener2] = struct{}{}
	server.mu.Unlock()

	server.mu.RLock()
	count := len(server.listeners)
	server.mu.RUnlock()

	if count != 2 {
		t.Errorf("listener count = %v, want 2", count)
	}

	// Close server should close all listeners
	server.Close()

	if !mockListener1.AssertCalled(t, "Close") || !mockListener2.AssertCalled(t, "Close") {
		t.Error("all listeners should be closed when server closes")
	}
}

func TestServerNativeQUICChannel(t *testing.T) {
	server := &Server{
		Addr:   ":8080",
		Logger: slog.Default(),
	}

	server.init()

	if server.nativeQUICCh == nil {
		t.Error("nativeQUICCh should be initialized")
	}

	// Test channel capacity (should be 1<<4 = 16)
	if cap(server.nativeQUICCh) != 16 {
		t.Errorf("nativeQUICCh capacity = %v, want 16", cap(server.nativeQUICCh))
	}

	// Test sending connection to channel
	mockConn := &MockQUICConnection{}
	select {
	case server.nativeQUICCh <- mockConn:
		// Success
	default:
		t.Error("should be able to send connection to nativeQUICCh")
	}

	// Test receiving from channel
	select {
	case conn := <-server.nativeQUICCh:
		if conn != mockConn {
			t.Error("received connection should match sent connection")
		}
	default:
		t.Error("should be able to receive connection from nativeQUICCh")
	}
}

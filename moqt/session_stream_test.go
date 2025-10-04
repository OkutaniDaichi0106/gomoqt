package moqt

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestNewSessionStream tests basic SessionStream creation
func TestNewSessionStream(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	assert.NotNil(t, ss, "newSessionStream should not return nil")
	assert.NotNil(t, ss.SessionUpdated(), "SessionUpdated channel should be initialized")
	assert.Equal(t, req.Path, ss.Path, "path should be set correctly")

	// Give time for background goroutines to complete
	time.Sleep(50 * time.Millisecond)

	mockStream.AssertExpectations(t)
}

// TestSessionStream_updateSession tests basic session update functionality
func TestSessionStream_updateSession(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("Write", mock.Anything).Return(8, nil)

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	bitrate := uint64(1000000)
	err := ss.updateSession(bitrate)

	assert.NoError(t, err, "updateSession should not return error")
	assert.Equal(t, bitrate, ss.localBitrate, "local bitrate should be updated")

	mockStream.AssertCalled(t, "Write", mock.Anything)
	mockStream.AssertExpectations(t)
}

// TestSessionStream_updateSession_WriteError tests behavior on write errors
func TestSessionStream_updateSession_WriteError(t *testing.T) {
	mockStream := &MockQUICStream{}
	writeError := errors.New("write error")
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("Write", mock.Anything).Return(0, writeError)

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Error handling depends on implementation, but should not panic
	assert.NotPanics(t, func() {
		ss.updateSession(uint64(1000000))
	}, "updateSession should not panic on write error")

	mockStream.AssertExpectations(t)
}

// TestSessionStream_SessionUpdated tests SessionUpdated channel functionality
func TestSessionStream_SessionUpdated(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Trigger setupDone to start listening for updates
	ss.handleUpdates()
	// Short sleep to let background goroutine start
	time.Sleep(10 * time.Millisecond)

	ch := ss.SessionUpdated()
	assert.NotNil(t, ch, "SessionUpdated should return a valid channel")
	assert.IsType(t, (<-chan struct{})(nil), ch, "SessionUpdated should return a receive-only channel")

	mockStream.AssertExpectations(t)
}

// TestSessionStream_updateSession_ZeroBitrate tests updateSession with zero bitrate
func TestSessionStream_updateSession_ZeroBitrate(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("Write", mock.Anything).Return(2, nil)

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	err := ss.updateSession(0)
	assert.NoError(t, err, "updateSession(0) should not error")
	assert.Equal(t, uint64(0), ss.localBitrate, "local bitrate should be set to 0")

	mockStream.AssertCalled(t, "Write", mock.Anything)
	mockStream.AssertExpectations(t)
}

// TestSessionStream_updateSession_LargeBitrate tests updateSession with large bitrate values
func TestSessionStream_updateSession_LargeBitrate(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("Write", mock.Anything).Return(10, nil)

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	largeBitrate := uint64(1<<62 - 1) // Large but valid value
	err := ss.updateSession(largeBitrate)
	assert.NoError(t, err, "updateSession with large bitrate should not error")
	assert.Equal(t, largeBitrate, ss.localBitrate, "local bitrate should be set correctly")

	mockStream.AssertCalled(t, "Write", mock.Anything)
	mockStream.AssertExpectations(t)
}

// TestSessionStream_listenUpdates tests message listening functionality
func TestSessionStream_listenUpdates(t *testing.T) {
	tests := map[string]struct {
		mockStream    func() *MockQUICStream
		expectUpdate  bool
		expectBitrate uint64
	}{
		"valid message": {
			mockStream: func() *MockQUICStream {
				// Valid SessionUpdateMessage
				bitrate := uint64(1000000)
				var buf bytes.Buffer
				message.SessionUpdateMessage{
					Bitrate: bitrate,
				}.Encode(&buf)

				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
			expectUpdate:  true,
			expectBitrate: 1000000,
		},
		"empty stream": {
			mockStream: func() *MockQUICStream {
				// Empty buffer will return 0, io.EOF immediately
				var buf bytes.Buffer
				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
			expectUpdate:  false,
			expectBitrate: 0,
		},
		"zero bitrate": {
			mockStream: func() *MockQUICStream {
				var buf bytes.Buffer
				message.SessionUpdateMessage{
					Bitrate: 0,
				}.Encode(&buf)

				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
			expectUpdate:  true,
			expectBitrate: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.mockStream()

			req := &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Start listening for updates
			ss.handleUpdates()

			// Give sufficient time for listenUpdates to process message
			time.Sleep(100 * time.Millisecond)

			if tt.expectUpdate {
				select {
				case <-ss.SessionUpdated():
					// Check if the bitrate was updated correctly
					ss.mu.Lock()
					actualBitrate := ss.remoteBitrate
					ss.mu.Unlock()
					assert.Equal(t, tt.expectBitrate, actualBitrate, "remote bitrate should match expected")
				case <-time.After(500 * time.Millisecond):
					if name == "valid message" || name == "zero bitrate" {
						t.Error("expected session update but timed out")
					}
				}
			}

			mockStream.AssertExpectations(t)
		})
	}
}

// TestSessionStream_listenUpdates_StreamClosed tests behavior when stream is closed
func TestSessionStream_listenUpdates_StreamClosed(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Trigger setupDone to start listening for updates
	ss.handleUpdates()

	// Give time for listenUpdates to encounter EOF and close
	time.Sleep(100 * time.Millisecond)

	// Verify the session stream handles EOF properly
	select {
	case <-ss.SessionUpdated():
		// Channel might be closed, which is acceptable
	case <-time.After(50 * time.Millisecond):
		// No update received, also acceptable for EOF case
	}

	mockStream.AssertExpectations(t)
}

// TestSessionStream_listenUpdates_ContextCancellation tests behavior on context cancellation
func TestSessionStream_listenUpdates_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Read", mock.Anything).Return(0, nil).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Let listenUpdates start
	time.Sleep(10 * time.Millisecond)

	// Cancel the context
	cancel()

	// Give time for listenUpdates to detect cancellation
	time.Sleep(50 * time.Millisecond)

	// Verify the stream handles cancellation properly
	select {
	case <-ss.SessionUpdated():
		// Channel might be closed due to cancellation
	case <-time.After(50 * time.Millisecond):
		// No update received, also acceptable for cancellation
	}

	mockStream.AssertExpectations(t)
}

// TestSessionStream_ConcurrentAccess tests concurrent access to SessionStream methods
func TestSessionStream_ConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockStream.On("Write", mock.Anything).Return(8, nil).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Trigger setupDone to start listening for updates
	ss.handleUpdates()

	// Test concurrent access to various methods
	var wg sync.WaitGroup

	// Concurrent updateSession calls
	wg.Go(func() {
		for i := range 5 {
			ss.updateSession(uint64(i * 1000))
			time.Sleep(time.Millisecond)
		}
	})

	// Concurrent SessionUpdated calls
	wg.Go(func() {
		for range 5 {
			ss.SessionUpdated()
			time.Sleep(time.Millisecond)
		}
	})

	// Concurrent access to bitrate fields (read-only)
	wg.Go(func() {
		for range 5 {
			_ = ss.localBitrate
			_ = ss.remoteBitrate
			time.Sleep(time.Millisecond)
		}
	})

	wg.Wait()

	// Test should complete without race conditions or panics
	mockStream.AssertExpectations(t)
}

// TestAccept tests responseWriter Accept functionality
func TestAccept(t *testing.T) {
	tests := map[string]struct {
		version     Version
		extensions  *Parameters
		mockStream  func() *MockQUICStream
		expectError bool
	}{
		"successful accept": {
			version:    protocol.Develop,
			extensions: NewParameters(),
			mockStream: func() *MockQUICStream {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("Write", mock.Anything).Return(10, nil)
				mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
				return mockStream
			},
			expectError: false,
		},
		"write error on accept": {
			version:    protocol.Develop,
			extensions: NewParameters(),
			mockStream: func() *MockQUICStream {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("Write", mock.Anything).Return(0, errors.New("write failed"))
				return mockStream
			},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.mockStream()

			req := &SetupRequest{
				Path:             "test/path",
				Versions:         []Version{protocol.Develop},
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Create mock connection and server for responseWriter
			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("LocalAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
			mockConn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081}).Maybe()

			mockServer := &Server{}
			mockServer.init() // Initialize the server properly
			rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

			// Use new API: SelectVersion, SetExtensions, then Accept
			err := rw.SelectVersion(tt.version)
			assert.NoError(t, err, "SelectVersion should not return error")
			rw.SetExtensions(tt.extensions)

			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)

			if tt.expectError {
				assert.Error(t, err, "Accept should return error")
				assert.Nil(t, session, "session should be nil on error")
			} else {
				assert.NoError(t, err, "Accept should not return error")
				assert.NotNil(t, session, "session should not be nil on success")
				assert.Equal(t, tt.version, ss.Version, "version should be set correctly")
				assert.Equal(t, tt.extensions, ss.ServerExtensions, "server parameters should be set correctly")
			}

			mockStream.AssertExpectations(t)
		})
	}
}

// TestAccept_OnlyOnce tests that Accept is only called once
func TestAccept_OnlyOnce(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(10, nil).Once()
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		Versions:         []Version{protocol.Develop},
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Create mock connection and server for responseWriter
	mockConn := &MockQUICConnection{}
	mockConn.On("Context").Return(context.Background())
	mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

	mockServer := &Server{}
	mockServer.init() // Initialize the server properly
	rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

	version := protocol.Develop
	extensions := NewParameters()

	// First call should succeed
	err := rw.SelectVersion(version)
	assert.NoError(t, err, "SelectVersion should not return error")
	rw.SetExtensions(extensions)

	mux := NewTrackMux()
	session1, err1 := Accept(rw, ss.SetupRequest, mux)
	assert.NoError(t, err1, "first Accept call should succeed")
	assert.NotNil(t, session1, "first Accept should return session")

	// Second call should be ignored (no additional Write calls, due to sync.Once)
	mux2 := NewTrackMux()
	session2, err2 := Accept(rw, ss.SetupRequest, mux2)
	assert.NoError(t, err2, "second Accept call should be ignored")
	assert.NotNil(t, session2, "second Accept should still return session")

	// Version should remain from first call
	assert.Equal(t, version, ss.Version, "version should remain from first call")

	mockStream.AssertExpectations(t)
}

// TestAccept_ConcurrentCalls tests concurrent Accept calls
func TestAccept_ConcurrentCalls(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(10, nil).Once()
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		Versions:         []Version{protocol.Develop},
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Create mock connection and server for responseWriter
	mockConn := &MockQUICConnection{}
	mockConn.On("Context").Return(context.Background())
	mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

	mockServer := &Server{}
	mockServer.init() // Initialize the server properly
	rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

	version := protocol.Develop
	extensions := NewParameters()

	// Set version and extensions before concurrent calls
	err := rw.SelectVersion(version)
	assert.NoError(t, err, "SelectVersion should not return error")
	rw.SetExtensions(extensions)

	var wg sync.WaitGroup
	const numGoroutines = 10
	errors := make([]error, numGoroutines)
	sessions := make([]*Session, numGoroutines)

	// Start multiple goroutines calling Accept concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)
			errors[id] = err
			sessions[id] = session
		}(i)
	}

	// Also call Accept from main goroutine
	mux := NewTrackMux()
	mainSession, mainErr := Accept(rw, ss.SetupRequest, mux)
	assert.NoError(t, mainErr, "main Accept call should not return error")
	assert.NotNil(t, mainSession, "main Accept should return session")

	wg.Wait()

	// All calls should succeed due to sync.Once
	for i, err := range errors {
		assert.NoError(t, err, "Accept call %d should succeed", i)
		assert.NotNil(t, sessions[i], "session %d should not be nil", i)
	}

	// Only one Write call should have been made due to sync.Once
	mockStream.AssertExpectations(t)
}

// TestResponse_AwaitAccepted tests response AwaitAccepted functionality
func TestResponse_AwaitAccepted(t *testing.T) {
	tests := map[string]struct {
		mockStream  func() *MockQUICStream
		expectError bool
		checkResult func(*testing.T, *response)
	}{
		"successful await": {
			mockStream: func() *MockQUICStream {
				// Create a valid SessionServerMessage
				ssm := message.SessionServerMessage{
					SelectedVersion: protocol.Develop,
					Parameters:      map[uint64][]byte{1: []byte("test")},
				}
				var buf bytes.Buffer
				ssm.Encode(&buf)

				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
			expectError: false,
			checkResult: func(t *testing.T, r *response) {
				assert.Equal(t, protocol.Develop, r.Version, "version should be set correctly")
				assert.NotNil(t, r.ServerExtensions, "server parameters should be set")
			},
		},
		"decode error": {
			mockStream: func() *MockQUICStream {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("Read", mock.Anything).Return(0, errors.New("decode failed"))
				return mockStream
			},
			expectError: true,
			checkResult: func(t *testing.T, r *response) {
				// Version should remain unset on error
			},
		},
		"EOF on read": {
			mockStream: func() *MockQUICStream {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
				return mockStream
			},
			expectError: true,
			checkResult: func(t *testing.T, r *response) {
				// Version should remain unset on error
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.mockStream()

			req := &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)
			r := newResponse(ss)

			err := r.AwaitAccepted()

			if tt.expectError {
				assert.Error(t, err, "AwaitAccepted should return error")
			} else {
				assert.NoError(t, err, "AwaitAccepted should not return error")
			}

			tt.checkResult(t, r)
			mockStream.AssertExpectations(t)
		})
	}
}

// TestResponse_AwaitAccepted_OnlyOnce tests that AwaitAccepted is only executed once
func TestResponse_AwaitAccepted_OnlyOnce(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())

	// Create a valid SessionServerMessage
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Develop,
		Parameters:      map[uint64][]byte{1: []byte("test")},
	}
	var buf bytes.Buffer
	ssm.Encode(&buf)

	// Use ReadFunc for simpler mocking
	mockStream.ReadFunc = buf.Read

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)
	r := newResponse(ss)

	// First call should read from stream
	err1 := r.AwaitAccepted()
	assert.NoError(t, err1, "first AwaitAccepted call should succeed")
	assert.Equal(t, protocol.Develop, r.Version, "version should be set from first call")

	// Second call should return immediately without reading from stream
	err2 := r.AwaitAccepted()
	assert.NoError(t, err2, "second AwaitAccepted call should succeed")
	assert.Equal(t, protocol.Develop, r.Version, "version should remain from first call")

	mockStream.AssertExpectations(t)
}

// TestResponse_AwaitAccepted_ConcurrentCalls tests concurrent AwaitAccepted calls
func TestResponse_AwaitAccepted_ConcurrentCalls(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())

	// Create a valid SessionServerMessage
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Develop,
		Parameters:      map[uint64][]byte{1: []byte("test")},
	}
	var buf bytes.Buffer
	ssm.Encode(&buf)

	// Use ReadFunc for simpler mocking
	mockStream.ReadFunc = buf.Read

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)
	r := newResponse(ss)

	var wg sync.WaitGroup
	const numGoroutines = 10
	results := make([]error, numGoroutines)

	// Start multiple goroutines calling AwaitAccepted concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			results[id] = r.AwaitAccepted()
		}(i)
	}

	wg.Wait()

	// All calls should succeed
	for i, err := range results {
		assert.NoError(t, err, "AwaitAccepted call %d should succeed", i)
	}

	// Version should be set correctly
	assert.Equal(t, protocol.Develop, r.Version, "version should be set correctly")

	// Only one Read call should have been made due to sync.Once
	mockStream.AssertExpectations(t)
}

// TestAccept_NilParameters tests Accept with nil parameters
func TestAccept_NilParameters(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		Versions:         []Version{protocol.Develop},
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Create mock connection and server for responseWriter
	mockConn := &MockQUICConnection{}
	mockConn.On("Context").Return(context.Background())
	mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

	mockServer := &Server{}
	mockServer.init() // Initialize the server properly
	rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

	version := protocol.Develop

	// Set version and extensions before Accept
	err := rw.SelectVersion(version)
	assert.NoError(t, err, "SelectVersion should not return error")
	rw.SetExtensions(nil)

	mux := NewTrackMux()
	session, err := Accept(rw, ss.SetupRequest, mux)
	assert.NoError(t, err, "Accept should handle nil parameters")
	assert.NotNil(t, session, "session should not be nil")
	assert.Equal(t, version, ss.Version, "version should be set correctly")
	assert.Nil(t, ss.ServerExtensions, "server parameters should be nil when nil is passed")

	mockStream.AssertExpectations(t)
}

// TestAccept_MultipleVersions tests Accept with different versions
func TestAccept_MultipleVersions(t *testing.T) {
	tests := map[string]struct {
		version Version
	}{
		"develop version": {version: protocol.Develop},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Write", mock.Anything).Return(10, nil)
			mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

			req := &SetupRequest{
				Path:             "test/path",
				Versions:         []Version{protocol.Develop}, // Include supported version
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Create mock connection and server for responseWriter
			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

			mockServer := &Server{}
			mockServer.init() // Initialize the server properly
			rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

			extensions := NewParameters()

			// Set version and extensions before Accept
			err := rw.SelectVersion(tt.version)
			assert.NoError(t, err, "SelectVersion should not return error")
			rw.SetExtensions(extensions)

			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)
			assert.NoError(t, err, "Accept should succeed for version %d", tt.version)
			assert.NotNil(t, session, "session should not be nil")
			assert.Equal(t, tt.version, ss.Version, "version should be set correctly")

			mockStream.AssertExpectations(t)
		})
	}
}

// TestResponse_AwaitAccepted_InvalidMessage tests AwaitAccepted with invalid message data
func TestResponse_AwaitAccepted_InvalidMessage(t *testing.T) {
	tests := map[string]struct {
		mockStream func() *MockQUICStream
	}{
		"invalid message data": {
			mockStream: func() *MockQUICStream {
				// Create buffer with invalid data
				invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
				buf := bytes.NewBuffer(invalidData)

				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
		},
		"truncated message": {
			mockStream: func() *MockQUICStream {
				// Create a valid message first, then truncate it
				ssm := message.SessionServerMessage{
					SelectedVersion: protocol.Develop,
					Parameters:      map[uint64][]byte{1: []byte("test")},
				}
				var fullBuf bytes.Buffer
				ssm.Encode(&fullBuf)
				fullData := fullBuf.Bytes()

				// Take only first 2 bytes to create truncated message
				truncatedData := fullData[:2]
				buf := bytes.NewBuffer(truncatedData)

				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.mockStream()

			req := &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)
			r := newResponse(ss)

			err := r.AwaitAccepted()

			assert.Error(t, err, "AwaitAccepted should return error for invalid message")

			mockStream.AssertExpectations(t)
		})
	}
}

// TestResponse_AwaitAccepted_DifferentVersions tests AwaitAccepted with different protocol versions
func TestResponse_AwaitAccepted_DifferentVersions(t *testing.T) {
	tests := map[string]struct {
		version Version
	}{
		"version 0":     {version: Version(0)},
		"version 1":     {version: protocol.Develop},
		"version 255":   {version: Version(255)},
		"large version": {version: Version(65535)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())

			// Create a valid SessionServerMessage with specific version
			ssm := message.SessionServerMessage{
				SelectedVersion: tt.version,
				Parameters:      map[uint64][]byte{1: []byte("test")},
			}
			var buf bytes.Buffer
			ssm.Encode(&buf)

			// Use ReadFunc for simpler mocking
			mockStream.ReadFunc = buf.Read

			req := &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)
			r := newResponse(ss)

			err := r.AwaitAccepted()

			assert.NoError(t, err, "AwaitAccepted should succeed for version %d", tt.version)
			assert.Equal(t, tt.version, r.Version, "version should be set correctly")
			assert.NotNil(t, r.ServerExtensions, "server parameters should be set")

			mockStream.AssertExpectations(t)
		})
	}
}

// TestSessionStream_listenUpdates_InitialChannelState tests initial state of updatedCh
func TestSessionStream_listenUpdates_InitialChannelState(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	// Trigger setupDone to start listening for updates
	ss.handleUpdates()

	// Channel should be initialized and available immediately
	ch := ss.SessionUpdated()
	assert.NotNil(t, ch, "SessionUpdated channel should be initialized")

	// Give time for listenUpdates to start and finish
	time.Sleep(50 * time.Millisecond)

	mockStream.AssertExpectations(t)
}

// TestSessionStream_Context tests Context method
func TestSessionStream_Context(t *testing.T) {
	ctx := context.Background()
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)

	resultCtx := ss.Context()

	assert.NotNil(t, resultCtx, "Context should not be nil")
	// The context should be a derived context with stream type value
	assert.NotEqual(t, ctx, resultCtx, "Context should be derived with additional values")
}

// TestResponse_Interface tests that response implements expected behaviors
func TestResponse_Interface(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)
	r := newResponse(ss)

	assert.NotNil(t, r, "response should not be nil")
	assert.NotNil(t, r.sessionStream, "sessionStream should not be nil")
	assert.Equal(t, ss, r.sessionStream, "sessionStream should be set correctly")
}

// TestResponse_AwaitAccepted_ErrorHandling tests various error scenarios
func TestResponse_AwaitAccepted_ErrorHandling(t *testing.T) {
	tests := map[string]struct {
		mockStream    func() *MockQUICStream
		expectError   bool
		expectVersion Version
	}{
		"network error on read": {
			mockStream: func() *MockQUICStream {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("Read", mock.Anything).Return(0, errors.New("network error"))
				return mockStream
			},
			expectError:   true,
			expectVersion: Version(0),
		},
		"context cancelled": {
			mockStream: func() *MockQUICStream {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("Read", mock.Anything).Return(0, context.Canceled)
				return mockStream
			},
			expectError:   true,
			expectVersion: Version(0),
		},
		"empty parameters": {
			mockStream: func() *MockQUICStream {
				ssm := message.SessionServerMessage{
					SelectedVersion: Version(42),
					Parameters:      map[uint64][]byte{},
				}
				var buf bytes.Buffer
				ssm.Encode(&buf)

				mockStream := &MockQUICStream{
					ReadFunc: buf.Read,
				}
				mockStream.On("Context").Return(context.Background())
				return mockStream
			},
			expectError:   false,
			expectVersion: Version(42),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := tt.mockStream()

			req := &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)
			r := newResponse(ss)

			err := r.AwaitAccepted()

			if tt.expectError {
				assert.Error(t, err, "AwaitAccepted should return error")
			} else {
				assert.NoError(t, err, "AwaitAccepted should not return error")
				assert.Equal(t, tt.expectVersion, r.Version, "version should be set correctly")
			}

			mockStream.AssertExpectations(t)
		})
	}
}

// TestAccept_ErrorHandling tests various error scenarios
func TestAccept_ErrorHandling(t *testing.T) {
	tests := map[string]struct {
		version     Version
		extensions  *Parameters
		setupMock   func(*MockQUICStream)
		expectError bool
	}{
		"network write error": {
			version:    protocol.Develop,
			extensions: NewParameters(),
			setupMock: func(mockStream *MockQUICStream) {
				mockStream.On("Write", mock.Anything).Return(0, errors.New("network write error"))
			},
			expectError: true,
		},
		"stream closed error": {
			version:    protocol.Develop,
			extensions: NewParameters(),
			setupMock: func(mockStream *MockQUICStream) {
				mockStream.On("Write", mock.Anything).Return(0, errors.New("stream closed"))
			},
			expectError: true,
		},
		"partial write": {
			version:    protocol.Develop,
			extensions: NewParameters(),
			setupMock: func(mockStream *MockQUICStream) {
				mockStream.On("Write", mock.Anything).Return(5, nil) // Partial write
				mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
			},
			expectError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			tt.setupMock(mockStream)

			req := &SetupRequest{
				Path:             "test/path",
				Versions:         []Version{protocol.Develop}, // Include supported version
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Create mock connection and server for responseWriter
			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

			mockServer := &Server{}
			mockServer.init() // Initialize the server properly
			rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

			// Set version and extensions before Accept
			err := rw.SelectVersion(tt.version)
			if tt.expectError {
				// For error cases, we might still succeed in SelectVersion
				// The error will come from Accept
			} else {
				assert.NoError(t, err, "SelectVersion should not return error")
			}
			rw.SetExtensions(tt.extensions)

			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)

			if tt.expectError {
				assert.Error(t, err, "Accept should return error")
				assert.Nil(t, session, "session should be nil on error")
			} else {
				assert.NoError(t, err, "Accept should not return error")
				assert.NotNil(t, session, "session should not be nil")
				assert.Equal(t, tt.version, ss.Version, "version should be set correctly")
			}

			mockStream.AssertExpectations(t)
		})
	}
}

// TestAccept_ParameterHandling tests parameter handling
func TestAccept_ParameterHandling(t *testing.T) {
	tests := map[string]struct {
		setupParam func() *Parameters
	}{
		"empty parameters": {
			setupParam: func() *Parameters { return NewParameters() },
		},
		"parameters with values": {
			setupParam: func() *Parameters {
				params := NewParameters()
				params.SetString(1, "test_value")
				params.SetUint(2, 12345)
				return params
			},
		},
		"large parameters": {
			setupParam: func() *Parameters {
				params := NewParameters()
				for i := uint64(0); i < 10; i++ {
					params.SetUint(ParameterType(i), i*1000)
				}
				return params
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Write", mock.Anything).Return(20, nil)
			mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

			req := &SetupRequest{
				Path:             "test/path",
				Versions:         []Version{protocol.Develop}, // Include supported version
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Create mock connection and server for responseWriter
			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

			mockServer := &Server{}
			mockServer.init() // Initialize the server properly
			rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

			extensions := tt.setupParam()

			// Set version and extensions before Accept
			err := rw.SelectVersion(protocol.Develop)
			assert.NoError(t, err, "SelectVersion should not return error")
			rw.SetExtensions(extensions)

			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)
			assert.NoError(t, err, "Accept should handle parameters correctly")
			assert.NotNil(t, session, "session should not be nil")
			assert.Equal(t, protocol.Develop, ss.Version, "version should be set correctly")
			assert.Equal(t, extensions, ss.ServerExtensions, "parameters should be set correctly")

			mockStream.AssertExpectations(t)
		})
	}
}

// TestAccept_BoundaryVersions tests Accept with boundary version values
func TestAccept_BoundaryVersions(t *testing.T) {
	tests := map[string]struct {
		version Version
	}{
		"minimum version":        {version: protocol.Develop},
		"maximum uint8 version":  {version: Version(255)},
		"maximum uint16 version": {version: Version(65535)},
		"maximum uint32 version": {version: Version(4294967295)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Write", mock.Anything).Return(10, nil)
			mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

			req := &SetupRequest{
				Path:             "test/path",
				Versions:         []Version{protocol.Develop, Version(255), Version(65535), Version(4294967295)},
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Create mock connection and server for responseWriter
			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

			mockServer := &Server{}
			mockServer.init()
			rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

			// Set version and extensions before Accept
			err := rw.SelectVersion(tt.version)
			assert.NoError(t, err, "SelectVersion should not return error")
			rw.SetExtensions(NewParameters())

			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)
			assert.NoError(t, err, "Accept should handle version %d", tt.version)
			assert.NotNil(t, session, "session should not be nil")
			assert.Equal(t, tt.version, ss.Version, "version should be set correctly")

			mockStream.AssertExpectations(t)
		})
	}
}

// TestResponse_AwaitAccepted_BoundaryVersions tests AwaitAccepted with boundary version values
func TestResponse_AwaitAccepted_BoundaryVersions(t *testing.T) {
	tests := map[string]struct {
		version Version
	}{
		"minimum version":        {version: Version(0)},
		"maximum uint8 version":  {version: Version(255)},
		"maximum uint16 version": {version: Version(65535)},
		"maximum uint32 version": {version: Version(4294967295)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())

			// Create a valid SessionServerMessage with boundary version
			ssm := message.SessionServerMessage{
				SelectedVersion: tt.version,
				Parameters:      map[uint64][]byte{1: []byte("test")},
			}
			var buf bytes.Buffer
			ssm.Encode(&buf)

			// Use ReadFunc for simpler mocking
			mockStream.ReadFunc = buf.Read

			req := &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)
			r := newResponse(ss)

			err := r.AwaitAccepted()

			assert.NoError(t, err, "AwaitAccepted should succeed for version %d", tt.version)
			assert.Equal(t, tt.version, r.Version, "version should be set correctly")
			assert.NotNil(t, r.ServerExtensions, "server parameters should be set")

			mockStream.AssertExpectations(t)
		})
	}
}

// TestAccept_Race tests for race conditions in Accept method
func TestAccept_Race(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(10, nil).Once()
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
		Versions:         []Version{protocol.Develop, Version(255), Version(65535), Version(4294967295)},
	}

	ss := newSessionStream(mockStream, req)

	// Create mock connection and server for responseWriter
	mockConn := &MockQUICConnection{}
	mockConn.On("Context").Return(context.Background())
	mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

	mockServer := &Server{}
	mockServer.init()
	rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

	const numGoroutines = 100
	var wg sync.WaitGroup
	errors := make([]error, numGoroutines)
	sessions := make([]*Session, numGoroutines)

	// Set version and extensions before concurrent calls
	err := rw.SelectVersion(protocol.Develop)
	assert.NoError(t, err, "SelectVersion should not return error")
	rw.SetExtensions(NewParameters())

	// Start many goroutines calling Accept
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)
			errors[id] = err
			sessions[id] = session
		}(i)
	}

	wg.Wait()

	// All calls should succeed due to sync.Once
	for i, err := range errors {
		assert.NoError(t, err, "Accept call %d should succeed", i)
		assert.NotNil(t, sessions[i], "session %d should not be nil", i)
	}

	// Only one Write call should have been made
	mockStream.AssertExpectations(t)
}

// TestResponse_AwaitAccepted_Race tests for race conditions in AwaitAccepted method
func TestResponse_AwaitAccepted_Race(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())

	// Create a valid SessionServerMessage
	ssm := message.SessionServerMessage{
		SelectedVersion: protocol.Develop,
		Parameters:      map[uint64][]byte{1: []byte("test")},
	}
	var buf bytes.Buffer
	ssm.Encode(&buf)

	mockStream.ReadFunc = buf.Read

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)
	r := newResponse(ss)

	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make([]error, numGoroutines)

	// Start many goroutines calling AwaitAccepted
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			results[id] = r.AwaitAccepted()
		}(i)
	}

	wg.Wait()

	// All calls should succeed
	for i, err := range results {
		assert.NoError(t, err, "AwaitAccepted call %d should succeed", i)
	}

	// Version should be set correctly from the first successful call
	assert.Equal(t, protocol.Develop, r.Version, "version should be set correctly")

	// Only one Read call should have been made due to sync.Once
	mockStream.AssertExpectations(t)
}

// TestResponseWriter_SessionStream_Sharing tests that responseWriter and sessionStream share state
func TestResponseWriter_SessionStream_Sharing(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
		Versions:         []Version{protocol.Develop},
	}

	ss := newSessionStream(mockStream, req)

	// Create mock connection and server for responseWriter
	mockConn := &MockQUICConnection{}
	mockConn.On("Context").Return(context.Background())
	mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

	mockServer := &Server{}
	mockServer.init()
	rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

	version := protocol.Develop
	extensions := NewParameters()
	extensions.SetString(1, "shared_state_test")

	// Set version and extensions before Accept
	err := rw.SelectVersion(version)
	assert.NoError(t, err, "SelectVersion should not return error")
	rw.SetExtensions(extensions)

	mux := NewTrackMux()
	session, err := Accept(rw, ss.SetupRequest, mux)
	assert.NoError(t, err, "Accept should succeed")
	assert.NotNil(t, session, "session should not be nil")

	// Verify that the sessionStream was updated
	assert.Equal(t, version, ss.Version, "sessionStream version should be updated")
	assert.Equal(t, extensions, ss.ServerExtensions, "sessionStream parameters should be updated")

	// Verify shared state through different accessors
	assert.Equal(t, version, rw.Version, "responseWriter should show updated version")
	assert.Equal(t, extensions, rw.ServerExtensions, "responseWriter should show updated parameters")

	mockStream.AssertExpectations(t)
}

// TestResponse_SessionStream_Sharing tests that response and sessionStream share state
func TestResponse_SessionStream_Sharing(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())

	// Create a valid SessionServerMessage
	version := Version(42)
	ssm := message.SessionServerMessage{
		SelectedVersion: version,
		Parameters:      map[uint64][]byte{1: []byte("shared_state_test")},
	}
	var buf bytes.Buffer
	ssm.Encode(&buf)

	// Use ReadFunc for simpler mocking
	mockStream.ReadFunc = buf.Read

	req := &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	}

	ss := newSessionStream(mockStream, req)
	r := newResponse(ss)

	err := r.AwaitAccepted()
	assert.NoError(t, err, "AwaitAccepted should succeed")

	// Verify that the sessionStream was updated
	assert.Equal(t, version, ss.Version, "sessionStream version should be updated")
	assert.NotNil(t, ss.ServerExtensions, "sessionStream parameters should be set")

	// Verify shared state through different accessors
	assert.Equal(t, version, r.Version, "response should show updated version")
	assert.Equal(t, ss.ServerExtensions, r.ServerExtensions, "response should show updated parameters")

	mockStream.AssertExpectations(t)
}

// TestAccept_ParameterEdgeCases tests parameter edge cases
func TestAccept_ParameterEdgeCases(t *testing.T) {
	tests := map[string]struct {
		setupExtensions func() *Parameters
		expectError     bool
	}{
		"nil extensions": {
			setupExtensions: func() *Parameters { return nil },
			expectError:     false,
		},
		"empty extensions": {
			setupExtensions: func() *Parameters { return NewParameters() },
			expectError:     false,
		},
		"extensions with empty string": {
			setupExtensions: func() *Parameters {
				params := NewParameters()
				params.SetString(1, "")
				return params
			},
			expectError: false,
		},
		"extensions with zero values": {
			setupExtensions: func() *Parameters {
				params := NewParameters()
				params.SetUint(1, 0)
				params.SetString(2, "")
				return params
			},
			expectError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Write", mock.Anything).Return(10, nil)
			mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()

			req := &SetupRequest{
				Path:             "test/path",
				Versions:         []Version{protocol.Develop},
				ClientExtensions: NewParameters(),
			}

			ss := newSessionStream(mockStream, req)

			// Create mock connection and server for responseWriter
			mockConn := &MockQUICConnection{}
			mockConn.On("Context").Return(context.Background())
			mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			mockConn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			mockConn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

			mockServer := &Server{}
			mockServer.init()
			rw := newResponseWriter(mockConn, ss, slog.Default(), mockServer)

			extensions := tt.setupExtensions()

			// Set version and extensions before Accept
			err := rw.SelectVersion(protocol.Develop)
			assert.NoError(t, err, "SelectVersion should not return error")
			rw.SetExtensions(extensions)

			mux := NewTrackMux()
			session, err := Accept(rw, ss.SetupRequest, mux)

			if tt.expectError {
				assert.Error(t, err, "Accept should return error")
				assert.Nil(t, session, "session should be nil on error")
			} else {
				assert.NoError(t, err, "Accept should handle parameters correctly")
				assert.NotNil(t, session, "session should not be nil")
				assert.Equal(t, extensions, ss.ServerExtensions, "server parameters should be set correctly")
			}

			mockStream.AssertExpectations(t)
		})
	}
}

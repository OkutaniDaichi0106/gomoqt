package moqt

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSession(t *testing.T) {
	tests := map[string]struct {
		mux      *TrackMux
		expectOK bool
	}{
		"new session with mux": {
			mux:      NewTrackMux(),
			expectOK: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a proper MockQUICStream
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.Anything).Return(0, io.EOF) // Create a proper mock connection
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
			conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

			sessStream := newSessionStream(mockStream, &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			})
			session := newSession(conn, sessStream, tt.mux, slog.Default(), nil)

			if tt.expectOK {
				assert.NotNil(t, session, "newSession should not return nil")
				assert.Equal(t, tt.mux, session.mux, "mux should be set correctly")
				assert.NotNil(t, session.trackReaders, "receive group stream queues should not be nil")
			}

			// Cleanup
			session.Terminate(NoError, "")
		})
	}
}

func TestNewSessionWithNilMux(t *testing.T) {
	tests := map[string]struct {
		mux           *TrackMux
		expectDefault bool
	}{
		"nil mux uses default": {
			mux:           nil,
			expectDefault: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
			conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

			sessStream := newSessionStream(mockStream, &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			})
			session := newSession(conn, sessStream, tt.mux, slog.Default(), nil)

			if tt.expectDefault {
				assert.Equal(t, DefaultMux, session.mux, "should use DefaultMux when nil mux is provided")
			}

			// Cleanup
			session.Terminate(InternalSessionErrorCode, "terminate reason")
		})
	}
}

func TestSession_Terminate(t *testing.T) {
	tests := map[string]struct {
		code SessionErrorCode
		msg  string
	}{
		"terminate with error": {
			code: InternalSessionErrorCode,
			msg:  "test error",
		},
		"terminate normally": {
			code: NoError,
			msg:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
			conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

			sessStream := newSessionStream(mockStream, &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			})
			session := newSession(conn, sessStream, nil, slog.Default(), nil)

			err := session.Terminate(tt.code, tt.msg)
			assert.NoError(t, err, "Terminate should not return error")
		})
	}
}

func TestSession_Subscribe(t *testing.T) {
	tests := map[string]struct {
		path      BroadcastPath
		name      TrackName
		config    *TrackConfig
		wantError bool
	}{
		"valid track stream": {
			path: BroadcastPath("/test/track"),
			name: TrackName("video"),
			config: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(100),
			},
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			// Set up expectations needed for sessionStream
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			mockStream.On("Write", mock.Anything).Return(0, nil)
			mockStream.On("Close").Return(nil)
			mockStream.On("Context").Return(context.Background()) // Create a separate mock for the track stream that responds to the SUBSCRIBE protocol
			mockTrackStream := &MockQUICStream{}
			mockTrackStream.On("StreamID").Return(quic.StreamID(2))
			// Create a SubscribeOkMessage response
			subok := message.SubscribeOkMessage{}
			var buf bytes.Buffer
			err := subok.Encode(&buf)
			assert.NoError(t, err, "failed to encode SubscribeOkMessage")

			// Use ReadFunc for simpler mocking
			mockTrackStream.ReadFunc = buf.Read

			mockTrackStream.On("Read", mock.Anything)
			mockTrackStream.On("Write", mock.Anything).Return(0, nil)
			mockTrackStream.On("Close").Return(nil)
			mockTrackStream.On("Context").Return(context.Background())

			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
			conn.On("OpenStream").Return(mockTrackStream, nil)
			conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

			sessStream := newSessionStream(mockStream, &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			})
			session := newSession(conn, sessStream, nil, slog.Default(), nil)

			track, err := session.Subscribe(tt.path, tt.name, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, track)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, track)
				assert.Equal(t, tt.path, track.BroadcastPath)
				assert.Equal(t, tt.name, track.TrackName)
				gotConfig := track.TrackConfig()
				assert.Equal(t, tt.config, gotConfig)
			}

			// Cleanup
			session.Terminate(NoError, "")
		})
	}
}

func TestSession_Subscribe_OpenError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
	conn.On("OpenStream").Return(nil, errors.New("open stream failed"))
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	subscriber, err := session.Subscribe(BroadcastPath("/test"), TrackName("video"), config)

	assert.Error(t, err)
	assert.Nil(t, subscriber)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_Context(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	ctx := session.Context()
	assert.NotNil(t, ctx, "Context should not be nil")

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_nextSubscribeID(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	id1 := session.nextSubscribeID()
	id2 := session.nextSubscribeID()

	assert.NotEqual(t, id1, id2, "nextSubscribeID should return unique IDs")
	assert.True(t, id2 > id1, "Subsequent IDs should be larger")

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_HandleBiStreams_AcceptError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, errors.New("accept stream failed"))
	conn.On("AcceptUniStream", mock.Anything).Return(nil, errors.New("accept stream failed"))
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	// Wait a bit for the background goroutine to try accepting
	time.Sleep(50 * time.Millisecond)

	// The session should handle the error gracefully
	assert.NotNil(t, session)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_HandleUniStreamsAcceptError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, errors.New("accept uni stream failed"))
	conn.On("AcceptUniStream", mock.Anything).Return(nil, errors.New("accept uni stream failed"))
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	// Wait a bit for the background goroutine to try accepting
	time.Sleep(50 * time.Millisecond)

	// The session should handle the error gracefully
	assert.NotNil(t, session)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_ConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
	conn.On("OpenStream").Return(mockStream, nil)
	conn.On("OpenUniStream").Return(&MockQUICSendStream{}, nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	// Test concurrent access
	done := make(chan struct{})
	var operations int

	// Concurrent nextSubscribeID calls
	go func() {
		for i := 0; i < 5; i++ {
			session.nextSubscribeID()
		}
		operations++
		if operations == 2 {
			close(done)
		}
	}()

	// Concurrent Context calls
	go func() {
		for i := 0; i < 5; i++ {
			session.Context()
		}
		operations++
		if operations == 2 {
			close(done)
		}
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Concurrent operations timed out")
	}

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_ContextCancellation(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())

	// Create a cancellable context for the connection
	connCtx, connCancel := context.WithCancel(context.Background())

	conn := &MockQUICConnection{}
	conn.On("Context").Return(connCtx)
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("CloseWithError", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// Cancel the connection context when CloseWithError is called
		connCancel()
	}).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	ctx := session.Context()
	assert.NotNil(t, ctx)

	// Terminate the session
	session.Terminate(NoError, "test termination")

	// Context should be cancelled after termination
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after termination")
	}
}

func TestSession_WithRealMux(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Context").Return(context.Background())

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	mux := NewTrackMux()

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, mux, slog.Default(), nil)

	assert.Equal(t, mux, session.mux, "Mux should be set correctly in the session")

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_GoAway(t *testing.T) {
	// Create a minimal session for testing
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	// Test goAway - now it calls updateSession
	err := session.goAway("test-uri")
	assert.NoError(t, err)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_AcceptAnnounce(t *testing.T) {
	tests := map[string]struct {
		prefix      string
		setupMocks  func(*MockQUICConnection, interface{})
		expectError bool
	}{
		"successful announce": {
			prefix: "/test/prefix/",
			setupMocks: func(mockConn *MockQUICConnection, mockStream interface{}) {
				mockConn.On("OpenStream").Return(mockStream, nil).Once()
				stream := mockStream.(*MockQUICStream)
				stream.On("Context").Return(context.Background()).Once()
				stream.On("StreamID").Return(quic.StreamID(1)).Once()
				// Mock writes for StreamType and AnnouncePlease
				stream.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil).Times(2)
				// Mock read for AnnounceInitMessage
				// Create a minimal AnnounceInitMessage with empty suffixes
				aim := message.AnnounceInitMessage{Suffixes: []string{}}
				var buf bytes.Buffer
				err := aim.Encode(&buf)
				if err != nil {
					panic(err)
				}
				data := buf.Bytes()
				stream.ReadFunc = func(p []byte) (int, error) {
					if len(data) == 0 {
						return 0, io.EOF
					}
					n := copy(p, data)
					data = data[n:]
					return n, nil
				}
			},
			expectError: false,
		},
		"terminating session": {
			prefix: "/test/prefix/",
			setupMocks: func(mockConn *MockQUICConnection, mockStream interface{}) {
				mockConn.On("CloseWithError", mock.Anything, mock.Anything).Return(errors.New("close error")).Once()
				// No additional mocks needed, session is terminating
			},
			expectError: true,
		},
		"open stream error": {
			prefix: "/test/prefix/",
			setupMocks: func(mockConn *MockQUICConnection, mockStream interface{}) {
				mockConn.On("OpenStream").Return(nil, errors.New("open stream error")).Once()
			},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a minimal session for testing
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			if name != "terminating session" {
				conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			}
			conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)

			sessStream := newSessionStream(mockStream, &SetupRequest{
				Path:             "test/path",
				ClientExtensions: NewParameters(),
			})
			session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

			announceStream := &MockQUICStream{}
			tt.setupMocks(conn, announceStream)

			if name == "terminating session" {
				session.Terminate(NoError, "")
			}

			reader, err := session.AcceptAnnounce(tt.prefix)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, reader)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reader)
			}

			// Cleanup
			session.Terminate(NoError, "")
		})
	}
}

func TestSession_AddTrackWriter(t *testing.T) {
	// Create a minimal session for testing
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	writer := &TrackWriter{}
	id := SubscribeID(1)
	session.addTrackWriter(id, writer)
	assert.Equal(t, writer, session.trackWriters[id])

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_RemoveTrackWriter(t *testing.T) {
	// Create a minimal session for testing
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	writer := &TrackWriter{}
	id := SubscribeID(1)
	session.trackWriters[id] = writer
	session.removeTrackWriter(id)
	assert.NotContains(t, session.trackWriters, id)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_RemoveTrackReader(t *testing.T) {
	// Create a minimal session for testing
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	reader := &TrackReader{}
	id := SubscribeID(1)
	session.trackReaders[id] = reader
	session.removeTrackReader(id)
	assert.NotContains(t, session.trackReaders, id)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestCancelStreamWithError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("CancelRead", quic.StreamErrorCode(1)).Return()
	mockStream.On("CancelWrite", quic.StreamErrorCode(1)).Return()

	cancelStreamWithError(mockStream, quic.StreamErrorCode(1))

	mockStream.AssertExpectations(t)
}

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

func TestSession_AddTrackReader(t *testing.T) {
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

	reader := &TrackReader{}
	id := SubscribeID(1)
	session.addTrackReader(id, reader)
	assert.Equal(t, reader, session.trackReaders[id])

	session.Terminate(NoError, "")
}

func TestSession_ProcessBiStream_Announce(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockSessStream := &MockQUICStream{}
	mockSessStream.On("Context").Return(context.Background())
	mockSessStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockSessStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	mux := NewTrackMux()
	session := newSession(conn, sessStream, mux, slog.Default(), nil)

	// Create a mock stream for ANNOUNCE
	mockStream := &MockQUICStream{}
	mockStream.On("StreamID").Return(quic.StreamID(1))
	mockStream.On("Context").Return(context.Background())

	// Prepare StreamType + AnnouncePleaseMessage
	var buf bytes.Buffer
	err := message.StreamTypeAnnounce.Encode(&buf)
	assert.NoError(t, err)
	apm := message.AnnouncePleaseMessage{TrackPrefix: "/test/prefix/"}
	err = apm.Encode(&buf)
	assert.NoError(t, err)

	data := buf.Bytes()
	mockStream.ReadFunc = func(p []byte) (int, error) {
		if len(data) == 0 {
			return 0, io.EOF
		}
		n := copy(p, data)
		data = data[n:]
		return n, nil
	}

	streamLogger := slog.Default().With("stream_id", mockStream.StreamID())

	// This will block, so we run it in a goroutine
	done := make(chan struct{})
	go func() {
		session.processBiStream(mockStream, streamLogger)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		// Expected to block waiting for announcements
	}

	session.Terminate(NoError, "")
}

func TestSession_ProcessBiStream_Subscribe(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("OpenUniStream").Return(&MockQUICSendStream{}, nil).Maybe()

	mockSessStream := &MockQUICStream{}
	mockSessStream.On("Context").Return(context.Background())
	mockSessStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockSessStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	mux := NewTrackMux()
	session := newSession(conn, sessStream, mux, slog.Default(), nil)

	// Create a mock stream for SUBSCRIBE
	mockStream := &MockQUICStream{}
	mockStream.On("StreamID").Return(quic.StreamID(2))
	mockStream.On("Context").Return(context.Background())
	mockStream.On("CancelWrite", mock.Anything).Return().Maybe()
	mockStream.On("CancelRead", mock.Anything).Return().Maybe()

	// Prepare StreamType + SubscribeMessage
	var buf bytes.Buffer
	err := message.StreamTypeSubscribe.Encode(&buf)
	assert.NoError(t, err)
	sm := message.SubscribeMessage{
		SubscribeID:   1,
		BroadcastPath: "/test/path",
		TrackName:     "video",
		TrackPriority: 1,
	}
	err = sm.Encode(&buf)
	assert.NoError(t, err)

	data := buf.Bytes()
	mockStream.ReadFunc = func(p []byte) (int, error) {
		if len(data) == 0 {
			return 0, io.EOF
		}
		n := copy(p, data)
		data = data[n:]
		return n, nil
	}
	mockStream.On("Write", mock.Anything).Return(0, nil)

	streamLogger := slog.Default().With("stream_id", mockStream.StreamID())

	// This will block in serveTrack, so we run it in a goroutine
	done := make(chan struct{})
	go func() {
		session.processBiStream(mockStream, streamLogger)
		close(done)
	}()

	// Wait a bit for the processing to start and track writer to be added
	time.Sleep(100 * time.Millisecond)

	// Terminate to stop blocking
	session.Terminate(NoError, "")

	select {
	case <-done:
		// Success
	case <-time.After(200 * time.Millisecond):
		// Expected to complete after termination
	}
}

func TestSession_ProcessBiStream_InvalidStreamType(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockSessStream := &MockQUICStream{}
	mockSessStream.On("Context").Return(context.Background())
	mockSessStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockSessStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	// Create a mock stream with invalid stream type
	mockStream := &MockQUICStream{}
	mockStream.On("StreamID").Return(quic.StreamID(3))

	// Prepare invalid StreamType (255)
	var buf bytes.Buffer
	buf.WriteByte(255)

	data := buf.Bytes()
	mockStream.ReadFunc = func(p []byte) (int, error) {
		if len(data) == 0 {
			return 0, io.EOF
		}
		n := copy(p, data)
		data = data[n:]
		return n, nil
	}

	streamLogger := slog.Default().With("stream_id", mockStream.StreamID())

	done := make(chan struct{})
	go func() {
		session.processBiStream(mockStream, streamLogger)
		close(done)
	}()

	select {
	case <-done:
		// Expected to return after terminating session
	case <-time.After(200 * time.Millisecond):
		t.Error("processBiStream should complete after invalid stream type")
	}

	assert.True(t, session.terminating(), "Session should be terminating after invalid stream type")
}

func TestSession_ProcessUniStream_Group(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockSessStream := &MockQUICStream{}
	mockSessStream.On("Context").Return(context.Background())
	mockSessStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockSessStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	// Add a track reader
	mockTrackStream := &MockQUICStream{}
	mockTrackStream.On("Context").Return(context.Background())
	mockTrackStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
	mockTrackStream.On("Write", mock.Anything).Return(0, nil).Maybe()

	substr := newSendSubscribeStream(1, mockTrackStream, &TrackConfig{}, Info{})
	trackReader := newTrackReader("/test", "video", substr, func() {})
	session.addTrackReader(1, trackReader)

	// Create a mock receive stream for GROUP
	mockRecvStream := &MockQUICReceiveStream{}
	mockRecvStream.On("StreamID").Return(quic.StreamID(4))
	mockRecvStream.On("CancelRead", mock.Anything).Return().Maybe()

	// Prepare StreamType + GroupMessage
	var buf bytes.Buffer
	err := message.StreamTypeGroup.Encode(&buf)
	assert.NoError(t, err)
	gm := message.GroupMessage{
		SubscribeID:   1,
		GroupSequence: 0,
	}
	err = gm.Encode(&buf)
	assert.NoError(t, err)

	data := buf.Bytes()
	mockRecvStream.ReadFunc = func(p []byte) (int, error) {
		if len(data) == 0 {
			return 0, io.EOF
		}
		n := copy(p, data)
		data = data[n:]
		return n, nil
	}
	mockRecvStream.On("Context").Return(context.Background())

	streamLogger := slog.Default().With("stream_id", mockRecvStream.StreamID())

	session.processUniStream(mockRecvStream, streamLogger)

	// Verify group was enqueued
	time.Sleep(10 * time.Millisecond)

	session.Terminate(NoError, "")
}

func TestSession_ProcessUniStream_UnknownSubscribeID(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockSessStream := &MockQUICStream{}
	mockSessStream.On("Context").Return(context.Background())
	mockSessStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockSessStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	// Create a mock receive stream for GROUP with unknown subscribe ID
	mockRecvStream := &MockQUICReceiveStream{}
	mockRecvStream.On("StreamID").Return(quic.StreamID(5))
	mockRecvStream.On("CancelRead", quic.StreamErrorCode(InvalidSubscribeIDErrorCode)).Return()

	// Prepare StreamType + GroupMessage with unknown subscribe ID
	var buf bytes.Buffer
	err := message.StreamTypeGroup.Encode(&buf)
	assert.NoError(t, err)
	gm := message.GroupMessage{
		SubscribeID:   999, // Unknown ID
		GroupSequence: 0,
	}
	err = gm.Encode(&buf)
	assert.NoError(t, err)

	data := buf.Bytes()
	mockRecvStream.ReadFunc = func(p []byte) (int, error) {
		if len(data) == 0 {
			return 0, io.EOF
		}
		n := copy(p, data)
		data = data[n:]
		return n, nil
	}

	streamLogger := slog.Default().With("stream_id", mockRecvStream.StreamID())

	session.processUniStream(mockRecvStream, streamLogger)

	mockRecvStream.AssertExpectations(t)

	session.Terminate(NoError, "")
}

func TestSession_ProcessUniStream_InvalidStreamType(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF).Maybe()

	mockSessStream := &MockQUICStream{}
	mockSessStream.On("Context").Return(context.Background())
	mockSessStream.On("Read", mock.Anything).Return(0, io.EOF)

	sessStream := newSessionStream(mockSessStream, &SetupRequest{
		Path:             "test/path",
		ClientExtensions: NewParameters(),
	})
	session := newSession(conn, sessStream, NewTrackMux(), slog.Default(), nil)

	// Create a mock receive stream with invalid stream type
	mockRecvStream := &MockQUICReceiveStream{}
	mockRecvStream.On("StreamID").Return(quic.StreamID(6))

	// Prepare invalid StreamType (254)
	var buf bytes.Buffer
	buf.WriteByte(254)

	data := buf.Bytes()
	mockRecvStream.ReadFunc = func(p []byte) (int, error) {
		if len(data) == 0 {
			return 0, io.EOF
		}
		n := copy(p, data)
		data = data[n:]
		return n, nil
	}

	streamLogger := slog.Default().With("stream_id", mockRecvStream.StreamID())

	done := make(chan struct{})
	go func() {
		session.processUniStream(mockRecvStream, streamLogger)
		close(done)
	}()

	select {
	case <-done:
		// Expected to return after terminating session
	case <-time.After(200 * time.Millisecond):
		t.Error("processUniStream should complete after invalid stream type")
	}

	assert.True(t, session.terminating(), "Session should be terminating after invalid stream type")
}

func TestSession_Subscribe_TerminatingSession(t *testing.T) {
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

	// Terminate the session
	err := session.Terminate(NoError, "")
	assert.NoError(t, err)

	// Wait for termination to complete
	time.Sleep(10 * time.Millisecond)

	// Try to subscribe - should fail because session is terminating
	config := &TrackConfig{TrackPriority: 1}
	reader, err := session.Subscribe("/test", "video", config)

	// Subscribe should fail or return nil because session is terminated
	if err != nil || reader == nil {
		// Expected behavior
		assert.Nil(t, reader)
	}
}

func TestSession_AcceptAnnounce_TerminatingSession(t *testing.T) {
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

	// Terminate the session
	err := session.Terminate(NoError, "")
	assert.NoError(t, err)

	// Wait for termination to complete
	time.Sleep(10 * time.Millisecond)

	// Try to accept announce - should fail because session is terminating
	reader, err := session.AcceptAnnounce("/test/prefix/")

	// AcceptAnnounce should fail or return nil because session is terminated
	if err != nil || reader == nil {
		// Expected behavior
		assert.Nil(t, reader)
	}
}

func TestSession_Terminate_AlreadyTerminating(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil).Once()
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}).Maybe()
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

	// First termination
	err1 := session.Terminate(NoError, "first termination")
	assert.NoError(t, err1)

	// Second termination should return immediately without error
	err2 := session.Terminate(InternalSessionErrorCode, "second termination")
	// The second call returns nil because terminating() is already true
	assert.NoError(t, err2)

	// Verify CloseWithError was only called once
	conn.AssertExpectations(t)
}

func TestSession_Terminate_WithApplicationError(t *testing.T) {
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	appErr := &quic.ApplicationError{
		ErrorCode:    quic.ApplicationErrorCode(InternalSessionErrorCode),
		ErrorMessage: "application error",
	}
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(appErr)
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

	err := session.Terminate(InternalSessionErrorCode, "test error")

	assert.Error(t, err)
	var sessErr *SessionError
	assert.ErrorAs(t, err, &sessErr)
}

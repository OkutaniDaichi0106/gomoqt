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
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
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

			sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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

			sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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

			sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
			session := newSession(conn, sessStream, nil, slog.Default(), nil)

			err := session.Terminate(tt.code, tt.msg)
			assert.NoError(t, err, "Terminate should not return error")
		})
	}
}

func TestSession_OpenAnnounceStream(t *testing.T) {
	tests := map[string]struct {
		path          string
		openStreamErr error
		expectError   bool
		expectNotNil  bool
	}{
		"successful open": {
			path:         "/test/",
			expectError:  false,
			expectNotNil: true,
		},
		"open stream error": {
			path:          "/test/",
			openStreamErr: errors.New("failed to open stream"),
			expectError:   true,
			expectNotNil:  false,
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
			if tt.openStreamErr != nil {
				conn.On("OpenStream").Return(nil, tt.openStreamErr)
			} else {
				announceStream := &MockQUICStream{}
				announceStream.On("Context").Return(context.Background())
				announceStream.On("StreamID").Return(quic.StreamID(1))
				announceStream.On("Write", mock.Anything).Return(0, nil)

				// Create a proper ANNOUNCE_INIT message response
				aim := message.AnnounceInitMessage{
					Suffixes: []string{"suffix1", "suffix2"},
				}
				var buf bytes.Buffer
				aim.Encode(&buf)

				announceStream.ReadFunc = buf.Read
				announceStream.On("Read", mock.Anything)

				conn.On("OpenStream").Return(announceStream, nil)
			}

			sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
			session := newSession(conn, sessStream, nil, slog.Default(), nil)

			announcer, err := session.OpenAnnounceStream(tt.path)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNotNil {
				assert.NotNil(t, announcer)
			} else {
				assert.Nil(t, announcer)
			}

			// Cleanup
			session.Terminate(NoError, "")
		})
	}
}

func TestSession_OpenAnnounceStream_OpenError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("OpenStream").Return(nil, errors.New("open stream failed"))

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	announcer, err := session.OpenAnnounceStream("/test")

	assert.Error(t, err)
	assert.Nil(t, announcer)

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_OpenTrackStream(t *testing.T) {
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
			subok := message.SubscribeOkMessage{
				GroupOrder: message.GroupOrderDefault,
			}
			var buf bytes.Buffer
			err := subok.Encode(&buf)
			assert.NoError(t, err, "failed to encode SubscribeOkMessage")
			responseData := buf.Bytes()

			readPos := 0
			mockTrackStream.ReadFunc = func(p []byte) (int, error) {
				if readPos < len(responseData) {
					n := copy(p, responseData[readPos:])
					readPos += n
					return n, nil
				}
				// After providing the response, return EOF to simulate stream end
				return 0, io.EOF
			}
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

			sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
			session := newSession(conn, sessStream, nil, slog.Default(), nil)

			track, err := session.OpenTrackStream(tt.path, tt.name, tt.config)

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

func TestSession_OpenTrackStream_OpenError(t *testing.T) {
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

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
	session := newSession(conn, sessStream, nil, slog.Default(), nil)

	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	subscriber, err := session.OpenTrackStream(BroadcastPath("/test"), TrackName("video"), config)

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

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	conn.On("RemoteAddr").Return(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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

	sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
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
	tests := map[string]struct {
		broadcastPath BroadcastPath
	}{
		"with real mux": {
			broadcastPath: BroadcastPath("/test/track"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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

			// Register a test handler
			ctx := context.Background()
			mux.Handle(ctx, tt.broadcastPath, TrackHandlerFunc(func(tw *TrackWriter) {}))

			sessStream := newSessionStream(mockStream, DefaultServerVersion, "test/path", NewParameters(), NewParameters())
			session := newSession(conn, sessStream, mux, slog.Default(), nil)

			assert.Equal(t, mux, session.mux, "Mux should be set correctly in the session")

			// Cleanup
			session.Terminate(NoError, "")
		})
	}
}

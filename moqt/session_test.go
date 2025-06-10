package moqt

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
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
			mockStream.On("Read", mock.Anything).Return(0, io.EOF) // Create a proper mock connection
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine

			session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, tt.mux, slog.Default())

			if tt.expectOK {
				assert.NotNil(t, session, "newSession should not return nil")
				assert.Equal(t, tt.mux, session.mux, "mux should be set correctly")
				assert.NotNil(t, session.trackReceivers, "receive group stream queues should not be nil")
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
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine

			session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, tt.mux, slog.Default())

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
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine

			session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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
			path:         "/test",
			expectError:  false,
			expectNotNil: true,
		},
		"open stream error": {
			path:          "/test",
			openStreamErr: errors.New("failed to open stream"),
			expectError:   true,
			expectNotNil:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)

			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
			if tt.openStreamErr != nil {
				conn.On("OpenStream").Return(nil, tt.openStreamErr)
			} else {
				announceStream := &MockQUICStream{}
				announceStream.On("Write", mock.Anything).Return(0, nil)
				announceStream.On("Read", mock.Anything).Return(0, io.EOF)
				conn.On("OpenStream").Return(announceStream, nil)
			}

			session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
	conn.On("OpenStream").Return(nil, errors.New("open stream failed"))

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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
		config    *SubscribeConfig
		wantError bool
	}{
		"valid track stream": {
			path: BroadcastPath("/test/track"),
			name: TrackName("video"),
			config: &SubscribeConfig{
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
			mockStream.On("Read", mock.Anything).Return(0, io.EOF).Maybe()
			mockStream.On("Write", mock.Anything).Return(0, nil).Maybe()
			mockStream.On("Close").Return(nil).Maybe()

			// Create a separate mock for the track stream that responds to the SUBSCRIBE protocol
			mockTrackStream := &MockQUICStream{}

			// Create a SubscribeOkMessage response
			subok := message.SubscribeOkMessage{
				GroupOrder: message.GroupOrderDefault,
			}
			var buf bytes.Buffer
			_, err := subok.Encode(&buf)
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
			mockTrackStream.On("Read", mock.Anything).Maybe()
			mockTrackStream.On("Write", mock.Anything).Return(0, nil).Maybe()
			mockTrackStream.On("Close").Return(nil).Maybe()

			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)    // For handleBiStreams goroutine
			conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF) // For handleUniStreams goroutine
			conn.On("OpenStream").Return(mockTrackStream, nil)

			session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

			subscriber, err := session.OpenTrackStream(tt.path, tt.name, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, subscriber)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, subscriber)
				assert.Equal(t, tt.path, subscriber.BroadcastPath)
				assert.Equal(t, tt.name, subscriber.TrackName)
				gotConfig := subscriber.SubscribeStream.SubscribeConfig()
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

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
	conn.On("OpenStream").Return(nil, errors.New("open stream failed"))

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

	config := &SubscribeConfig{
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

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

	ctx := session.Context()
	assert.NotNil(t, ctx, "Context should not be nil")

	// Cleanup
	session.Terminate(NoError, "")
}

func TestSession_nextSubscribeID(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, errors.New("accept stream failed"))
	conn.On("AcceptUniStream", mock.Anything).Return(nil, errors.New("accept stream failed"))

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, errors.New("accept uni stream failed"))
	conn.On("AcceptUniStream", mock.Anything).Return(nil, errors.New("accept uni stream failed"))

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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
	mockStream.On("Write", mock.Anything).Return(0, nil).Maybe()
	mockStream.On("Close").Return(nil).Maybe()
	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
	conn.On("AcceptStream", mock.Anything).Return(nil, io.EOF)
	conn.On("AcceptUniStream", mock.Anything).Return(nil, io.EOF)
	conn.On("OpenStream").Return(mockStream, nil).Maybe()
	conn.On("OpenUniStream").Return(&MockQUICSendStream{}, nil).Maybe()

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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

	conn := &MockQUICConnection{}
	conn.On("Context").Return(context.Background())
	conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()
	conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)

	session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, nil, slog.Default())

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

			conn := &MockQUICConnection{}
			conn.On("Context").Return(context.Background())
			conn.On("CloseWithError", mock.Anything, mock.Anything).Return(nil)
			conn.On("AcceptStream", mock.Anything).Return(nil, context.Canceled).Maybe()
			conn.On("AcceptUniStream", mock.Anything).Return(nil, context.Canceled).Maybe()

			mux := NewTrackMux()

			// Register a test handler
			ctx := context.Background()
			mux.Handle(ctx, tt.broadcastPath, TrackHandlerFunc(func(p *Publisher) {}))

			session := newSession(conn, internal.DefaultServerVersion, "path", NewParameters(), NewParameters(), mockStream, mux, slog.Default())

			assert.Equal(t, mux, session.mux, "Mux should be set correctly in the session")

			// Cleanup
			session.Terminate(NoError, "")
		})
	}
}

package moqt

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
)

// createTestTrackContext creates a trackContext for testing purposes
func createTestTrackContext(sessCtx *sessionContext) *trackContext {
	return newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))
}

func TestNewTrackContext(t *testing.T) {
	tests := []struct {
		name       string
		sessCtx    *sessionContext
		id         SubscribeID
		path       BroadcastPath
		trackName  TrackName
		wantLogger bool
	}{
		{
			name: "with logger",
			sessCtx: newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				NewParameters(),
				slog.Default(),
				nil,
			),
			id:         SubscribeID(1),
			path:       BroadcastPath("/test/path"),
			trackName:  TrackName("test-track"),
			wantLogger: true,
		},
		{
			name: "without logger",
			sessCtx: newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				NewParameters(),
				nil,
				nil,
			),
			id:         SubscribeID(2),
			path:       BroadcastPath("/test/path2"),
			trackName:  TrackName("test-track-2"),
			wantLogger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTrackContext(tt.sessCtx, tt.id, tt.path, tt.trackName)

			if ctx == nil {
				t.Fatal("newTrackContext returned nil")
			}

			if ctx.id != tt.id {
				t.Errorf("id = %v, want %v", ctx.id, tt.id)
			}

			if ctx.path != tt.path {
				t.Errorf("path = %v, want %v", ctx.path, tt.path)
			}

			if ctx.name != tt.trackName {
				t.Errorf("name = %v, want %v", ctx.name, tt.trackName)
			}

			if tt.wantLogger && ctx.logger == nil {
				t.Error("expected logger to be set")
			}

			if !tt.wantLogger && ctx.logger != nil {
				t.Error("expected logger to be nil")
			}

			if ctx.Context == nil {
				t.Error("Context should not be nil")
			}

			if ctx.cancel == nil {
				t.Error("cancel function should not be nil")
			}
		})
	}
}

func TestTrackContext_Logger(t *testing.T) {
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	ctx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

	logger := ctx.Logger()
	if logger == nil {
		t.Error("Logger() returned nil")
	}
}

func TestTrackContext_Cancel(t *testing.T) {
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	ctx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

	// Context should not be done initially
	select {
	case <-ctx.Done():
		t.Error("context should not be done initially")
	default:
	}

	// Cancel the context
	testErr := ErrClosedTrack
	ctx.cancel(testErr)

	// Context should be done after cancel
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("context should be done after cancel")
	}

	// Check the cause
	if cause := context.Cause(ctx); cause != testErr {
		t.Errorf("context cause = %v, want %v", cause, testErr)
	}
}

func TestTrackContext_LoggerAttributes(t *testing.T) {
	// Test that trackContext.Logger() returns logger with correct attributes
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		logger,
		nil,
	)
	trackCtx := newTrackContext(sessCtx, SubscribeID(42), BroadcastPath("/test/path"), TrackName("test-track"))

	// Get the logger from track context
	trackLogger := trackCtx.Logger()
	assert.NotNil(t, trackLogger, "Track logger should not be nil")

	// Log a test message to verify attributes
	trackLogger.Info("test message")

	// Parse the logged output
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"remote_address":"session"`, "Should contain inherited remote_address attribute")
	assert.Contains(t, logOutput, `"subscribe_id":"42"`, "Should contain subscribe_id attribute")
	assert.Contains(t, logOutput, `"broadcast_path":"/test/path"`, "Should contain broadcast_path attribute")
	assert.Contains(t, logOutput, `"track_name":"test-track"`, "Should contain track_name attribute")
	assert.Contains(t, logOutput, `"msg":"test message"`, "Should contain the log message")
}

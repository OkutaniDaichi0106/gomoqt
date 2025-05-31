package moqt

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
)

// createTestGroupContext creates a groupContext for testing purposes
func createTestGroupContext(trackCtx *trackContext) *groupContext {
	return newGroupContext(trackCtx, GroupSequence(123))
}

func TestNewGroupContext(t *testing.T) {
	// Create a session context first
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	// Create a track context
	trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

	// Test basic functionality with a single sequence - sequence value doesn't affect core functionality
	seq := GroupSequence(123)
	ctx := newGroupContext(trackCtx, seq)

	if ctx == nil {
		t.Fatal("newGroupContext returned nil")
	}

	if ctx.seq != seq {
		t.Errorf("seq = %v, want %v", ctx.seq, seq)
	}

	if ctx.Context == nil {
		t.Error("Context should not be nil")
	}

	if ctx.cancel == nil {
		t.Error("cancel function should not be nil")
	}

	// Check that logger is properly set
	logger := ctx.Logger()
	if logger == nil {
		t.Error("Logger() should not return nil")
	}
}

func TestGroupContext_Logger(t *testing.T) {
	tests := []struct {
		name   string
		logger *slog.Logger
	}{
		{
			name:   "with logger",
			logger: slog.Default(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create contexts
			sessCtx := newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				NewParameters(),
				tt.logger,
				nil,
			)
			trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))
			groupCtx := newGroupContext(trackCtx, GroupSequence(123))

			gotLogger := groupCtx.Logger()
			if tt.logger == nil {
				assert.Nil(t, gotLogger, "Logger() should return nil when no logger is set")
			}
			if tt.logger != nil {
				assert.NotNil(t, gotLogger, "Logger() should not return nil when logger is set")
			}

		})
	}
}

func TestGroupContext_Cancel(t *testing.T) {
	// Create contexts using helper function
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	groupCtx := createTestGroupContext(trackCtx)

	// Context should not be done initially
	select {
	case <-groupCtx.Done():
		t.Error("context should not be done initially")
	default:
	}

	// Cancel the context
	testErr := ErrClosedGroup
	groupCtx.cancel(testErr)

	// Context should be done after cancel
	select {
	case <-groupCtx.Done():
		// Expected
	default:
		t.Error("context should be done after cancel")
	}

	// Check the cause
	if cause := context.Cause(groupCtx); cause != testErr {
		t.Errorf("context cause = %v, want %v", cause, testErr)
	}
}

func TestGroupContext_LoggerAttributes(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Create contexts with specific values for testing
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		logger,
		nil,
	)
	trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))
	groupCtx := newGroupContext(trackCtx, GroupSequence(123))

	// Get the logger from group context
	groupLogger := groupCtx.Logger()
	assert.NotNil(t, groupLogger, "Group logger should not be nil")

	// Log a test message to verify attributes
	groupLogger.Info("test message")

	// Parse the logged output
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"remote_address":"session"`, "Should contain session remote_address attribute")
	assert.Contains(t, logOutput, `"subscribe_id":"1"`, "Should contain subscribe_id attribute")
	assert.Contains(t, logOutput, `"broadcast_path":"/test"`, "Should contain broadcast_path attribute")
	assert.Contains(t, logOutput, `"track_name":"test"`, "Should contain track_name attribute")
	assert.Contains(t, logOutput, `"group_sequence":"123"`, "Should contain group_sequence attribute")
	assert.Contains(t, logOutput, `"msg":"test message"`, "Should contain the log message")
}

func TestGroupContextLoggerWithNilParent(t *testing.T) {
	// Test that group context logger is nil when parent track context logger is nil
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		nil, // nil logger
		nil,
	)
	trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))
	groupCtx := newGroupContext(trackCtx, GroupSequence(123))

	groupLogger := groupCtx.Logger()
	assert.Nil(t, groupLogger, "Group logger should be nil when parent logger is nil")
}

func TestGroupContextLoggerWithDifferentValues(t *testing.T) {
	// Test logger attributes with various GroupSequence values
	tests := []struct {
		name        string
		seq         GroupSequence
		expectedSeq string
	}{{
		name:        "with sequence 0",
		seq:         GroupSequence(0),
		expectedSeq: "0",
	},
		{
			name:        "with sequence 1",
			seq:         GroupSequence(1),
			expectedSeq: "1",
		},
		{
			name:        "with large sequence",
			seq:         GroupSequence(4294967295), // Max uint32
			expectedSeq: "4294967295",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))
			groupCtx := newGroupContext(trackCtx, tt.seq)

			groupLogger := groupCtx.Logger()
			assert.NotNil(t, groupLogger, "Group logger should not be nil")

			// Log a test message
			groupLogger.Info("test message")

			// Verify the group_sequence attribute
			logOutput := buf.String()
			expectedAttr := `"group_sequence":"` + tt.expectedSeq + `"`
			assert.Contains(t, logOutput, expectedAttr, "Should contain the correct group_sequence attribute")
		})
	}
}

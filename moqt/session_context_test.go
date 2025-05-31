package moqt

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/stretchr/testify/assert"
)

// createTestSessionContext creates a sessionContext for testing purposes
func createTestSessionContext(ctx context.Context) *sessionContext {
	return newSessionContext(
		ctx,
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)
}

func TestNewSessionContext(t *testing.T) {
	tests := []struct {
		name         string
		connCtx      context.Context
		version      protocol.Version
		path         string
		clientParams *Parameters
		serverParams *Parameters
		logger       *slog.Logger
		tracer       *moqtrace.SessionTracer
	}{
		{
			name:         "basic creation",
			connCtx:      context.Background(),
			version:      protocol.Version(0x1),
			path:         "/test",
			clientParams: NewParameters(),
			serverParams: NewParameters(),
			logger:       slog.Default(),
			tracer:       &moqtrace.SessionTracer{},
		},
		{
			name:         "with nil logger",
			connCtx:      context.Background(),
			version:      protocol.Version(0x2),
			path:         "/test2",
			clientParams: nil,
			serverParams: nil,
			logger:       nil,
			tracer:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newSessionContext(
				tt.connCtx,
				tt.version,
				tt.path,
				tt.clientParams,
				tt.serverParams,
				tt.logger,
				tt.tracer,
			)

			if ctx == nil {
				t.Fatal("newSessionContext returned nil")
			}

			if ctx.path != tt.path {
				t.Errorf("path = %v, want %v", ctx.path, tt.path)
			}

			if ctx.version != tt.version {
				t.Errorf("version = %v, want %v", ctx.version, tt.version)
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

func TestSessionContext_Logger(t *testing.T) {
	logger := slog.Default()
	ctx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		logger,
		nil,
	)

	if ctx.Logger() == nil {
		t.Error("Logger() should not return nil")
	}
}

func TestSessionContext_LoggerAttributes(t *testing.T) {
	// Test that sessionContext.Logger() returns logger with correct attributes
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

	// Get the logger from session context
	sessLogger := sessCtx.Logger()
	assert.NotNil(t, sessLogger, "Session logger should not be nil")

	// Log a test message to verify attributes
	sessLogger.Info("test message")

	// Parse the logged output
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"remote_address":"session"`, "Should contain remote_address attribute")
	assert.Contains(t, logOutput, `"msg":"test message"`, "Should contain the log message")
}

func TestSessionContext_Path(t *testing.T) {
	expectedPath := "/test/path"
	ctx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		expectedPath,
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	if ctx.Path() != expectedPath {
		t.Errorf("Path() = %v, want %v", ctx.Path(), expectedPath)
	}
}

func TestSessionContext_Version(t *testing.T) {
	expectedVersion := protocol.Version(0x123)
	ctx := newSessionContext(
		context.Background(),
		expectedVersion,
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	if ctx.Version() != expectedVersion {
		t.Errorf("Version() = %v, want %v", ctx.Version(), expectedVersion)
	}
}

func TestSessionContext_ClientParameters(t *testing.T) {
	tests := []struct {
		name   string
		params *Parameters
		isNil  bool
	}{
		{
			name:   "with parameters",
			params: NewParameters(),
			isNil:  false,
		},
		{
			name:   "with nil parameters",
			params: nil,
			isNil:  false, // Should return default params
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				tt.params,
				NewParameters(),
				slog.Default(),
				nil,
			)

			result := ctx.ClientParameters()
			if tt.isNil && result != nil {
				t.Error("ClientParameters() should return nil")
			}
			if !tt.isNil && result == nil {
				t.Error("ClientParameters() should not return nil")
			}
		})
	}
}

func TestSessionContext_ServerParameters(t *testing.T) {
	tests := []struct {
		name   string
		params *Parameters
		isNil  bool
	}{
		{
			name:   "with parameters",
			params: NewParameters(),
			isNil:  false,
		},
		{
			name:   "with nil parameters",
			params: nil,
			isNil:  false, // Should return default params
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				tt.params,
				slog.Default(),
				nil,
			)

			result := ctx.ServerParameters()
			if tt.isNil && result != nil {
				t.Error("ServerParameters() should return nil")
			}
			if !tt.isNil && result == nil {
				t.Error("ServerParameters() should not return nil")
			}
		})
	}
}

func TestSessionContext_Tracer(t *testing.T) {
	ctx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	tracer := ctx.Tracer()
	if tracer != nil {
		t.Error("Tracer() should return nil when no tracer is set")
	}
}

func TestSessionContext_Cancel(t *testing.T) {
	ctx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)

	// Context should not be done initially
	select {
	case <-ctx.Done():
		t.Error("context should not be done initially")
	default:
	}

	// Cancel the context
	testErr := ErrInternalError
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

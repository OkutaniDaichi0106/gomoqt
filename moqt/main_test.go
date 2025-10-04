package moqt

import (
	"io"
	"log/slog"
	"testing"
)

// Silence slog output during tests by setting a logger that writes to io.Discard.
func TestMain(m *testing.M) {
	// Replace default logger with a no-op logger for tests
	handler := slog.NewTextHandler(io.Discard, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	m.Run()
}

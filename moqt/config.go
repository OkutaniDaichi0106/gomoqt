package moqt

import (
	"net/http"
	"time"

	"github.com/okdaichi/gomoqt/moqt/bitrate"
)

// Config contains configuration options for MOQ sessions.
type Config struct {
	// ServerSetupExtensions func(clientParams *Parameters) (serverParams *Parameters, err error)

	// MaxSubscribeID SubscribeID // TODO:

	// NewSessionURI string // TODO:

	// CheckHTTPOrigin validates the HTTP Origin header for WebTransport connections.
	// If nil, all origins are accepted.
	CheckHTTPOrigin func(*http.Request) bool // TODO: Check HTTP header for security

	// SetupTimeout is the maximum time to wait for session setup to complete.
	// If zero, a default timeout of 5 seconds is used.
	SetupTimeout time.Duration

	// NewShiftDetector is a function that creates a new ShiftDetector instance.
	// If nil, a default EWMA detector with sensible defaults is used.
	NewShiftDetector func() bitrate.ShiftDetector
}

// newShiftDetector returns a new ShiftDetector instance.
// It uses the configured NewShiftDetector function.
// If Config is nil or NewShiftDetector is not set, returns nil (disables BPS monitoring).
func (c *Config) newShiftDetector() bitrate.ShiftDetector {
	if c != nil && c.NewShiftDetector != nil {
		return c.NewShiftDetector()
	}
	return nil
}

// setupTimeout returns the configured setup timeout or a default value.
func (c *Config) setupTimeout() time.Duration {
	if c != nil && c.SetupTimeout > 0 {
		return c.SetupTimeout
	}
	return 5 * time.Second
}

// checkHTTPOrigin returns the configured CheckHTTPOrigin function or nil.
func (c *Config) checkHTTPOrigin() func(*http.Request) bool {
	if c != nil {
		return c.CheckHTTPOrigin
	}
	return nil
}

// Clone creates a copy of the Config.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}
	return &Config{
		// ServerSetupExtensions: c.ServerSetupExtensions,
		// MaxSubscribeID: c.MaxSubscribeID,
		// NewSessionURI:  c.NewSessionURI,
		// CheckRoot:      c.CheckRoot,
		CheckHTTPOrigin:  c.CheckHTTPOrigin,
		SetupTimeout:     c.SetupTimeout,
		NewShiftDetector: c.NewShiftDetector,
	}
}

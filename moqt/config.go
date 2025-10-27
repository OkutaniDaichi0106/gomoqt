package moqt

import (
	"net/http"
	"time"
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
}

// Clone creates a copy of the Config.
func (c *Config) Clone() *Config {
	return &Config{
		// ServerSetupExtensions: c.ServerSetupExtensions,
		// MaxSubscribeID: c.MaxSubscribeID,
		// NewSessionURI:  c.NewSessionURI,
		// CheckRoot:      c.CheckRoot,
		SetupTimeout: c.SetupTimeout,
	}
}

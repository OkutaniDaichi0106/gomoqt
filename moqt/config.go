package moqt

import (
	"net/http"
	"time"
)

type Config struct {
	ClientSetupExtensions func() *Parameters

	// ServerSetupExtensions func(clientParams *Parameters) (serverParams *Parameters, err error)

	// MaxSubscribeID SubscribeID // TODO:

	// NewSessionURI string // TODO:

	CheckHTTPOrigin func(*http.Request) bool // TODO: Check HTTP header for security

	SetupTimeout time.Duration
}

func (c *Config) Clone() *Config {
	return &Config{
		ClientSetupExtensions: c.ClientSetupExtensions,
		// ServerSetupExtensions: c.ServerSetupExtensions,
		// MaxSubscribeID: c.MaxSubscribeID,
		// NewSessionURI:  c.NewSessionURI,
		// CheckRoot:      c.CheckRoot,
		SetupTimeout: c.SetupTimeout,
	}
}

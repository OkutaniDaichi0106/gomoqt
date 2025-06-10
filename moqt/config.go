package moqt

import (
	"time"
)

type Config struct {
	ClientSetupExtensions func() *Parameters

	ServerSetupExtensions func(clientParams *Parameters) (serverParams *Parameters, err error)

	// MaxSubscribeID SubscribeID // TODO:

	// NewSessionURI string // TODO:

	// CheckRoot func(r SetupRequest) bool // TODO:

	SetupTimeout time.Duration
}

func (c *Config) Clone() *Config {
	return &Config{
		ClientSetupExtensions: c.ClientSetupExtensions,
		ServerSetupExtensions: c.ServerSetupExtensions,
		// MaxSubscribeID: c.MaxSubscribeID,
		// NewSessionURI:  c.NewSessionURI,
		// CheckRoot:      c.CheckRoot,
		SetupTimeout: c.SetupTimeout,
	}
}

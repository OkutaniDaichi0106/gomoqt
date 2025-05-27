package moqt

import "time"

type Config struct {
	// Configurations
	MaxSubscribeID SubscribeID // TODO:

	NewSessionURI string // TODO:

	// SetupExtensions Parameters

	CheckRoot func(r SetupRequest) bool // TODO:

	Timeout time.Duration
}

func (c *Config) Clone() *Config {
	return &Config{
		MaxSubscribeID: c.MaxSubscribeID,
		NewSessionURI:  c.NewSessionURI,
		CheckRoot:      c.CheckRoot,
	}
}

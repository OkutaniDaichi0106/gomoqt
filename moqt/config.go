package moqt

type Config struct {
	// Configurations
	MaxSubscribeID SubscribeID // TODO:

	NewSessionURI string // TODO:

	// SetupExtensions Parameters

	CheckRoot func(r SetupRequest) bool // TODO:
}

func (c *Config) Clone() *Config {
	return &Config{
		MaxSubscribeID: c.MaxSubscribeID,
		NewSessionURI:  c.NewSessionURI,
		CheckRoot:      c.CheckRoot,
	}
}

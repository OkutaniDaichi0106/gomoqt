package moqt

type Config struct {
	// Configurations
	MaxSubscribeID SubscribeID // TODO:

	NewSessionURI string // TODO:

	// SetupExtensions Parameters

	CheckRoot func(r SetupRequest) bool
}

package moqt

type Config struct {
	// Configurations
	MaxSubscribeID SubscribeID

	NewSessionURI string

	SetupExtensions Parameters
}

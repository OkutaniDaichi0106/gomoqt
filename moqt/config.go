package moqt

type Config struct {
	// Configurations
	MaxSubscribeID uint64

	NewSessionURI string

	SetupExtension Parameters
}

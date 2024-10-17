package moqtransport

type Client struct {
	conn              Connection
	SupportedVersions []Version
}

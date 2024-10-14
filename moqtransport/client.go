package moqtransport

import "github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"

type Client struct {
	conn              Connection
	SupportedVersions []moqtmessage.Version
}

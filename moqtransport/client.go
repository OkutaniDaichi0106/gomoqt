package moqtransport

import "github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"

type client struct {
	conn              Connection
	SupportedVersions []moqtmessage.Version
	setupStream       Stream
}

package moqtransport

import "github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"

type Version uint64

const (
	FoalkDraft01 Version = Version(moqtmessage.FoalkDraft01)
)

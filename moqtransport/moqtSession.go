package moqtransport

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type moqtSession struct {
	TransportSession

	sessionStream Stream

	selectedVersion moqtmessage.Version
}

func (sess moqtSession) OpenAnnounceStream() {

}

func (sess moqtSession) OpenSubscribeStream() SubscribeStream {
}

func (sess moqtSession) Terminate(err TerminateError) {
	sess.CloseWithError(SessionErrorCode(err.Code()), err.Error())
}

type AnnounceConfig struct {
	AuthorizationInfo []string

	MaxCacheDuration time.Duration
}

type SubscribeConfig struct {
	moqtmessage.SubscriberPriority
	moqtmessage.GroupOrder

	AuthorizationInfo string
	DeliveryTimeout   time.Duration
}

package moqtransport

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type pubsubSessionID uint64

// TODO
type PubSubSession struct {
	pubsubSessionID

	sessionCore

	subscriptions map[moqtmessage.SubscribeID]*Subscription

	trackAliasMap trackAliasMap

	maxSubscribeID   moqtmessage.SubscribeID
	maxCacheDuration time.Duration

	contentStatuses map[moqtmessage.TrackAlias]*contentStatus
}

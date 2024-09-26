package moqtransport

import (
	"go-moq/moqtransport/moqtmessage"
	"time"
)

// TODO
type PubSubSession struct {
	sessionCore
	localTrackNamespace moqtmessage.TrackNamespace
	//remoteTrackNamespaces map[string]moqtmessage.TrackNamespace

	maxSubscribeID   moqtmessage.SubscribeID
	maxCacheDuration time.Duration

	//trackManager
}

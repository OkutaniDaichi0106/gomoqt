package moqtransport

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type SubscribeStream struct {
}

func (stream SubscribeStream) Subscribe() (TrackStatus, error) {
	/*
	 * Send a SUBSCRIBE message
	 */
	/*
	 * Receive a TRACK_STATUS message
	 */
}

func (stream SubscribeStream) RequestTrackStatus() (TrackStatus, error) {
	/*
	 * Send a TRACK_STATUS_REQUEST message
	 */
	/*
	 * Receive a TRACK_STATUS message
	 */
}

func (stream SubscribeStream) UpdateSubscription() (TrackStatus, error) {
	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	/*
	 * Receive a TRACK_STATUS message
	 */
}

type TrackStatus struct {
	TrackAlias    moqtmessage.TrackAlias
	Code          moqtmessage.TrackStatusCode
	LatestGroupID moqtmessage.GroupID
	GroupOrder    moqtmessage.GroupOrder
	GroupExpires  time.Duration
}

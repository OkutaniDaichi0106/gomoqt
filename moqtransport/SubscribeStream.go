package moqtransport

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

type SendSubscribeStream struct {
	stream           Stream
	qvReader         quicvarint.Reader
	subscribeCounter *uint64
	trackAliasMap    *trackAliasMap
}

func (stream SendSubscribeStream) Subscribe(trackNamespace moqtmessage.TrackNamespace, trackName string, config SubscribeConfig) (*TrackStatus, error) {
	/*
	 * Send a SUBSCRIBE message
	 */
	// Get next Subscribe ID
	newSubscribeNumber := atomic.AddUint64(stream.subscribeCounter, 1)
	// Get Track Alias
	trackAlias := stream.trackAliasMap.getAlias(trackNamespace, trackName)
	// Initialize a SUBSCRIBE message
	sm := moqtmessage.SubscribeMessage{
		SubscribeID:        moqtmessage.SubscribeID(newSubscribeNumber),
		TrackAlias:         trackAlias,
		TrackNamespace:     trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		MinGroupSequence:   0,
		MaxGroupSequence:   0,
		Parameters:         make(moqtmessage.Parameters),
	}
	// Append the parameters
	if config.AuthorizationInfo != nil {
		sm.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, config.AuthorizationInfo)
	}
	if config.DeliveryTimeout != nil {
		sm.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, config.DeliveryTimeout)
	}
	if config.Parameters != nil {
		for k, v := range *config.Parameters {
			sm.Parameters.AddParameter(k, v)
		}
	}

	_, err := stream.stream.Write(sm.Serialize())
	if err != nil {
		return nil, err
	}

	/*
	 * Receive a TRACK_STATUS message
	 */
	id, preader, err := moqtmessage.ReadControlMessage(stream.qvReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.TRACK_STATUS {
		return nil, ErrProtocolViolation
	}
	var tsm moqtmessage.TrackStatusMessage
	err = tsm.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}

	// Verify the responce
	if trackAlias != tsm.TrackAlias {
		return nil, errors.New("unexpected track alias")
	}
	if config.GroupOrder != tsm.GroupOrder {
		return nil, errors.New("unacceptable group order")
	}

	return &TrackStatus{
		TrackAlias: tsm.TrackAlias,
		Code:       tsm.Code,
		GroupOrder: tsm.GroupOrder,
	}, nil
}

func (stream SendSubscribeStream) RequestTrackStatus() (TrackStatus, error) {
	/*
	 * Send a TRACK_STATUS_REQUEST message
	 */
	/*
	 * Receive a TRACK_STATUS message
	 */
}

func (stream SendSubscribeStream) UpdateSubscription() (TrackStatus, error) {
	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	/*
	 * Receive a TRACK_STATUS message
	 */
}

type ReceiveSubscribeStream struct {
	stream   Stream
	qvReader quicvarint.Reader
}

func (stream ReceiveSubscribeStream) WaitSubscribe() (SubscribeConfig, error) {

}

type TrackStatus struct {
	TrackAlias    moqtmessage.TrackAlias
	Code          moqtmessage.TrackStatusCode
	LatestGroupID moqtmessage.GroupID
	GroupOrder    moqtmessage.GroupOrder
	GroupExpires  time.Duration
}

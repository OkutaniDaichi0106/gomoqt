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
	//
	subscription *Subscription
}

func (stream SendSubscribeStream) Subscribe(trackNamespace moqtmessage.TrackNamespace, trackName string, config SubscribeConfig) (*Subscription, *TrackStatus, error) {
	if stream.subscription != nil {
		return nil, nil, errors.New("subscribed stream")
	}

	/*
	 * Send a SUBSCRIBE message
	 */
	// Get next Subscribe ID
	newSubscribeNumber := atomic.AddUint64(stream.subscribeCounter, 1)
	// Get the Track Alias
	trackAlias := stream.trackAliasMap.getAlias(trackNamespace, trackName)
	// Get the new Subscribe ID
	newSubscribeID := moqtmessage.SubscribeID(newSubscribeNumber)
	// Initialize a SUBSCRIBE message
	sm := moqtmessage.SubscribeMessage{
		SubscribeID:        newSubscribeID,
		TrackAlias:         trackAlias,
		TrackNamespace:     trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		MinGroupSequence:   config.MinGroupSequence,
		MaxGroupSequence:   config.MaxGroupSequence,
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
		for k, v := range config.Parameters {
			sm.Parameters.AddParameter(k, v)
		}
	}

	_, err := stream.stream.Write(sm.Serialize())
	if err != nil {
		return nil, nil, err
	}

	/*
	 * Receive a TRACK_STATUS message
	 */
	ts, err := receiveTrackStatus(stream.qvReader, trackAlias)
	if err != nil {
		return nil, nil, err
	}
	// Verify if the Track's Group Order is the specified Group Order
	if config.GroupOrder != ts.GroupOrder {
		return nil, nil, errors.New("not specified group order")
	}

	return &Subscription{
		subscribeID:    newSubscribeID,
		trackNamespace: trackNamespace,
		trackName:      trackName,
		trackAlias:     trackAlias,
		config:         config,
	}, ts, nil
}

func (stream SendSubscribeStream) UpdateSubscription(subscription Subscription, config SubscribeConfig) (*TrackStatus, error) {
	/*
	 * Verify the new configuration is valid
	 */
	if config.GroupOrder != 0 {
		return nil, errors.New("invalid group order")
	}
	if config.MinGroupSequence > config.MaxGroupSequence {
		return nil, errors.New("min larger than max")
	}
	if config.MinGroupSequence < subscription.config.MinGroupSequence {
		return nil, errors.New("wider update")
	}
	if config.MaxGroupSequence > subscription.config.MaxGroupSequence {
		return nil, errors.New("wider update")
	}

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	// Initialize a SUBSCRIBE_UPDATE message
	sum := moqtmessage.SubscribeUpdateMessage{
		SubscribeID:        subscription.subscribeID,
		SubscriberPriority: config.SubscriberPriority,
		MinGroupNumber:     config.MinGroupSequence,
		MaxGroupNumber:     config.MaxGroupSequence,
		Parameters:         make(moqtmessage.Parameters),
	}
	// Append the parameters
	if config.AuthorizationInfo != nil {
		sum.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, config.AuthorizationInfo)
	}
	if config.DeliveryTimeout != nil {
		sum.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, config.DeliveryTimeout)
	}
	if config.Parameters != nil {
		for k, v := range config.Parameters {
			sum.Parameters.AddParameter(k, v)
		}
	}
	// Send the message
	_, err := stream.stream.Write(sum.Serialize())
	if err != nil {
		return nil, err
	}

	/*
	 * Receive a TRACK_STATUS message
	 */
	ts, err := receiveTrackStatus(stream.qvReader, subscription.trackAlias)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func receiveTrackStatus(qvReader quicvarint.Reader, trackAlias moqtmessage.TrackAlias) (*TrackStatus, error) {
	/*
	 * Receive a TRACK_STATUS message
	 */
	id, preader, err := moqtmessage.ReadControlMessage(qvReader)
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
		return nil, ErrProtocolViolation
	}

	return &TrackStatus{
		Code:          tsm.Code,
		GroupOrder:    tsm.GroupOrder,
		LatestGroupID: tsm.LatestGroupID,
		GroupExpires:  tsm.GroupExpires,
	}, nil
}

// TODO
/*
 * Cancel the all subscriptions on the stream.
 */
func (stream SendSubscribeStream) CancelSubscribe(err SubscribeError) {
	stream.stream.CancelWrite(StreamErrorCode(err.SubscribeErrorCode()))
}

type ReceiveSubscribeStream struct {
	stream        Stream
	qvReader      quicvarint.Reader
	subscriptions map[moqtmessage.SubscribeID]Subscription
}

func (stream ReceiveSubscribeStream) ReceiveSubscribe() (*Subscription, error) {
	// Read the SUBSCRIBE message
	id, preader, err := moqtmessage.ReadControlMessage(stream.qvReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.SUBSCRIBE {
		return nil, ErrUnexpectedMessage
	}
	var sm moqtmessage.SubscribeMessage
	err = sm.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}

	// Initialize a configuration of the subscription
	config := SubscribeConfig{
		SubscriberPriority: sm.SubscriberPriority,
		GroupOrder:         sm.GroupOrder,
		MinGroupSequence:   sm.MinGroupSequence,
		MaxGroupSequence:   sm.MaxGroupSequence,
	}
	if authInfo, ok := sm.Parameters.AuthorizationInfo(); ok {
		config.AuthorizationInfo = &authInfo
		sm.Parameters.Remove(moqtmessage.AUTHORIZATION_INFO)
	}
	if timeout, ok := sm.Parameters.DeliveryTimeout(); ok {
		config.DeliveryTimeout = &timeout
		sm.Parameters.Remove(moqtmessage.AUTHORIZATION_INFO)
	}
	config.Parameters = sm.Parameters

	return &Subscription{
		subscribeID:    sm.SubscribeID,
		trackNamespace: sm.TrackNamespace,
		trackName:      sm.TrackName,
		trackAlias:     sm.TrackAlias,
		config:         config,
	}, nil
}

func (stream ReceiveSubscribeStream) AllowSubscribe(subscription Subscription) {
	/***/
	stream.subscriptions[subscription.subscribeID] = subscription
} // TODO:

func (stream ReceiveSubscribeStream) RejectSubscribe(err SubscribeError) {
	stream.stream.CancelRead(StreamErrorCode(err.SubscribeErrorCode())) // TODO:
}

func (stream ReceiveSubscribeStream) ReceiveSubscribeUpdate() (*Subscription, error) {
	// Read the SUBSCRIBE_UPDATE message
	id, preader, err := moqtmessage.ReadControlMessage(stream.qvReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.SUBSCRIBE_UPDATE {
		return nil, ErrUnexpectedMessage
	}
	var sum moqtmessage.SubscribeUpdateMessage
	err = sum.DeserializePayload(preader)
	if err != nil {
		return nil, err
	}

	config := SubscribeConfig{}

	return &Subscription{
		subscribeID: sum.SubscribeID,
		config:      config,
	}, nil
}

// func (stream ReceiveSubscribeStream) PeekMessageID() (moqtmessage.MessageID, error) {
// 	peeker := bufio.NewReader(stream.qvReader)
// 	b, err := peeker.Peek(1 << 3)
// 	if err != nil {
// 		return 0, err
// 	}
// 	qvReader := quicvarint.Reader(bytes.NewReader(b))
// 	id, _, err := moqtmessage.ReadControlMessage(qvReader)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return id, nil
// }

type Subscription struct {
	subscribeID    moqtmessage.SubscribeID
	trackNamespace moqtmessage.TrackNamespace
	trackName      string
	trackAlias     moqtmessage.TrackAlias
	config         SubscribeConfig
}

func (s Subscription) GetConfig() SubscribeConfig {
	return s.config
}

type TrackStatus struct {
	Code          moqtmessage.TrackStatusCode
	LatestGroupID moqtmessage.GroupID
	GroupOrder    moqtmessage.GroupOrder
	GroupExpires  time.Duration
}

var ErrUnexpectedMessage = errors.New("unexpected message")

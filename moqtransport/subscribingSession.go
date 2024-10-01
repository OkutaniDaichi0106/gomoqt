package moqtransport

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type Announcement struct {
	trackNamespace moqtmessage.TrackNamespace

	AuthorizationInfo string

	Parameters moqtmessage.Parameters
}

/*
 *
 *
 */

type SubscribingSession struct {
	mu sync.RWMutex

	sessionCore

	trackAliasMap trackAliasMap

	subscriptions map[moqtmessage.SubscribeID]Subscription

	//chunkReaderChs map[moqtmessage.SubscribeID]chan<- ReceiveByteStream

	relayConfig map[moqtmessage.SubscribeID]struct {
		deliveryTimeout time.Duration
	}

	expiries map[moqtmessage.SubscribeID]*struct {
		expireCtx    context.Context
		expireCancel context.CancelFunc
	}
}

func (sess *SubscribingSession) AcceptUniStream(ctx context.Context) (ReceiveByteStream, error) {
	return sess.trSess.AcceptUniStream(ctx)
}

func (sess *SubscribingSession) WaitAnnounce() (*Announcement, error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return nil, err
	}

	if id != moqtmessage.ANNOUNCE {
		return nil, ErrProtocolViolation
	}

	var am moqtmessage.AnnounceMessage
	err = am.DeserializeBody(sess.controlReader)
	if err != nil {
		return nil, err
	}

	announcement := Announcement{
		trackNamespace: am.TrackNamespace,
		Parameters:     am.Parameters,
	}

	authInfo, ok := am.Parameters.AuthorizationInfo()
	if ok {
		announcement.AuthorizationInfo = authInfo
	}
	// TODO: Handle the paramter

	return &announcement, nil
}

func (sess *SubscribingSession) AllowAnnounce(announcement Announcement) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	// Register the Track Namaspace
	node := trackManager.trackNamespaceTree.insert(announcement.trackNamespace)

	node.mu.Lock()
	defer node.mu.Unlock()

	// Register the parameters
	node.params = &announcement.Parameters

	// Send an ANNOUNCE_OK message
	ao := moqtmessage.AnnounceOkMessage{
		TrackNamespace: announcement.trackNamespace,
	}

	_, err := sess.controlStream.Write(ao.Serialize())

	return err
}

func (sess *SubscribingSession) RejectAnnounce(announcement Announcement, annErr AnnounceError) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	// Send an ANNOUNCE_ERROR message
	ae := moqtmessage.AnnounceErrorMessage{
		TrackNamespace: announcement.trackNamespace,
		Code:           annErr.Code(),
		Reason:         annErr.Error(),
	}

	_, err := sess.controlStream.Write(ae.Serialize())
	if err != nil {
		log.Println(err)
		return
	}
}

func (sess *SubscribingSession) Subscribe(announcement Announcement, trackName string, config SubscribeConfig) (*Subscription, error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	// Set the default group order value, if the value is 0
	if config.GroupOrder == 0 {
		config.GroupOrder = moqtmessage.ASCENDING
	}

	// Set the default filter code, if the value is 0
	if config.SubscriptionFilter.Code == 0 {
		config.SubscriptionFilter.Code = moqtmessage.LATEST_GROUP
	}

	// Get new Subscribe ID
	newSubscribeID := moqtmessage.SubscribeID(len(sess.subscriptions))

	// Get the Track Alias
	alias := sess.trackAliasMap.getAlias(announcement.trackNamespace, trackName)

	/*
	 * Send a SUBSCRIBE message
	 */
	sm := moqtmessage.SubscribeMessage{
		SubscribeID:        newSubscribeID,
		TrackAlias:         alias,
		TrackNamespace:     announcement.trackNamespace,
		TrackName:          trackName,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		SubscriptionFilter: config.SubscriptionFilter,
		Parameters:         make(moqtmessage.Parameters),
	}

	// Add the authorization information parameter
	sm.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, config.AuthorizationInfo)

	// Add the delivery timeout parameter
	if config.DeliveryTimeout != 0 {
		sm.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, config.DeliveryTimeout)
	}

	_, err := sess.controlStream.Write(sm.Serialize())
	if err != nil {
		return nil, err
	}

	subscription := Subscription{
		subscribeID:    newSubscribeID,
		trackAlias:     alias,
		trackNamespace: announcement.trackNamespace,
		trackName:      trackName,
		config:         config,
	}

	/*
	 * Receive a SUBSCRIBE_OK message of a SUBSCRIBE_ERROR message
	 */
	err = sess.receiveSubscribeResponce(subscription)
	if err != nil {
		retryErr, ok := err.(RetryTrackAliasError)
		if ok {
			sm.TrackAlias = retryErr.trackAlias
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //TODO
			defer cancel()

			return sess.retrySubscribe(subscription, ctx)
		}

		return nil, err
	}

	return &subscription, nil
}

func (sess *SubscribingSession) retrySubscribe(subscription Subscription, ctx context.Context) (*Subscription, error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		/*
		 * Send a SUBSCRIBE message
		 */
		sm := moqtmessage.SubscribeMessage{
			SubscribeID:        subscription.subscribeID,
			TrackAlias:         subscription.trackAlias,
			TrackNamespace:     subscription.trackNamespace,
			TrackName:          subscription.trackName,
			SubscriberPriority: subscription.config.SubscriberPriority,
			GroupOrder:         subscription.config.GroupOrder,
			SubscriptionFilter: subscription.config.SubscriptionFilter,
		}
		_, err := sess.controlStream.Write(sm.Serialize())
		if err != nil {
			return nil, err
		}

		err = sess.receiveSubscribeResponce(subscription)

		if err != nil {
			retryErr, ok := err.(RetryTrackAliasError)
			if ok {
				subscription.trackAlias = retryErr.trackAlias
				return sess.retrySubscribe(subscription, ctx)
			}

			return nil, err
		}
		return &subscription, nil
	}
}

func (sess *SubscribingSession) receiveSubscribeResponce(subscription Subscription) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	/*
	 * Receive a SUBSCRIBE_OK message or a SUBSCRIBE_ERROR message
	 */
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}

	switch id {
	case moqtmessage.SUBSCRIBE_OK:
		var so moqtmessage.SubscribeOkMessage
		err = so.DeserializePayload(sess.controlReader)
		if err != nil {
			return err
		}

		if so.SubscribeID != subscription.subscribeID {
			return ErrProtocolViolation
		}

		if so.GroupOrder != subscription.config.GroupOrder {
			return ErrProtocolViolation
		}

		//TODO: Handle the delivery timeout
		//timeout, ok := so.Parameters.DeliveryTimeout()

		// Update the content status in track
		tnsNode, ok := trackManager.findTrackNamespace(subscription.trackNamespace)
		if !ok {
			return errors.New("track namespace not found")
		}
		tnNode, ok := tnsNode.findTrackName(subscription.trackName)
		if !ok {
			tnNode = tnsNode.newTrackNameNode(subscription.trackName)
		}

		// Update the content status
		tnNode.contentStatus = &contentStatus{
			contentExists:   so.ContentExists,
			largestGroupID:  so.LargestGroupID,
			largestObjectID: so.LargestObjectID,
		}

		var ctx context.Context
		var cancel context.CancelFunc

		// If a expiry is specified, create the context with the expiry
		if so.Expires != 0 {
			ctx, cancel = context.WithTimeout(context.Background(), so.Expires)
		}
		sess.expiries[so.SubscribeID] = &struct {
			expireCtx    context.Context
			expireCancel context.CancelFunc
		}{
			expireCtx:    ctx,
			expireCancel: cancel,
		}

		return nil
	case moqtmessage.SUBSCRIBE_ERROR:
		var se moqtmessage.SubscribeErrorMessage // TODO: Handle Error Code
		err = se.DeserializePayload(sess.controlReader)
		if err != nil {
			return err
		}

		if se.SubscribeID != subscription.subscribeID {
			return ErrProtocolViolation
		}

		return GetSubscribeError(se)
	default:
		return ErrProtocolViolation
	}
}

func (sess *SubscribingSession) UpdateSubscription(subscribeID moqtmessage.SubscribeID, config SubscribeConfig) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	// Verify if a subscription with the Subscribe ID exists
	subscription, ok := sess.subscriptions[subscribeID]
	if !ok {
		return errors.New("subscription not found")
	}

	// Update the subscription
	subscription.config = config

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	sum := moqtmessage.SubscribeUpdateMessage{
		SubscribeID:        subscribeID,
		FilterRange:        config.SubscriptionFilter.Range,
		SubscriberPriority: config.SubscriberPriority,
		Parameters:         make(moqtmessage.Parameters), // TODO:
	}

	if config.AuthorizationInfo != "" {
		sum.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, config.AuthorizationInfo)
	}

	if config.DeliveryTimeout != 0 {
		sum.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, config.DeliveryTimeout)
	}

	_, err := sess.controlStream.Write(sum.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Receive a SUBSCRIBE_OK or a SUBSCRIBE_ERROR message
	 */

	err = sess.receiveSubscribeResponce(subscription)

	return err
}

func (sess *SubscribingSession) Unsubscribe(subscribeID moqtmessage.SubscribeID) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	us := moqtmessage.UnsubscribeMessage{
		SubscribeID: subscribeID,
	}

	_, err := sess.controlStream.Write(us.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Receive a SUBSCRIBE_DONE message
	 */
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}

	if id != moqtmessage.SUBSCRIBE_DONE {
		return ErrProtocolViolation
	}

	var sdm moqtmessage.SubscribeDoneMessage

	err = sdm.DeserializePayload(sess.controlReader)
	if err != nil {
		return err
	}

	if sdm.SubscribeID != subscribeID {
		return ErrProtocolViolation
	}

	if sdm.StatusCode != moqtmessage.SUBSCRIBE_DONE_UNSUBSCRIBED {
		return ErrProtocolViolation
	}

	return nil
}

func (sess *SubscribingSession) SubscribeNamespace(trackNamespacePrefix moqtmessage.TrackNamespacePrefix, authInfo string) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	/*
	 * Send a SUBSCRIBE_NAMESPACE message
	 */
	snsm := moqtmessage.SubscribeNamespaceMessage{
		TrackNamespacePrefix: trackNamespacePrefix,
		Parameters:           make(moqtmessage.Parameters),
	}

	if len(authInfo) > 0 {
		snsm.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, authInfo)
	}

	_, err := sess.controlStream.Write(snsm.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Receive a SUBSCRIBE_NAMESPACE_OK message or a SUBSCRIBE_NAMESPACE_ERROR message
	 */
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}
	switch id {
	case moqtmessage.SUBSCRIBE_NAMESPACE_OK:
		var sno moqtmessage.SubscribeNamespaceOkMessage

		err := sno.DeserializePayload(sess.controlReader)
		if err != nil {
			return err
		}

		if strings.Join(sno.TrackNamespacePrefix, "") != strings.Join(trackNamespacePrefix, "") {
			return errors.New("unexpected track namespace prefix")
		}

		return nil
	case moqtmessage.SUBSCRIBE_NAMESPACE_ERROR:
		var sne moqtmessage.SubscribeNamespaceErrorMessage

		err := sne.DeserializePayload(sess.controlReader)
		if err != nil {
			return err
		}

		if strings.Join(sne.TrackNamespacePrefix, "") != strings.Join(trackNamespacePrefix, "") {
			return errors.New("unexpected track namespace prefix")
		}

		return DefaultSubscribeNamespaceError{
			code:   sne.Code,
			reason: sne.Reason,
		}
	default:
		return ErrProtocolViolation
	}
}

func (sess *SubscribingSession) UnsubscribeNamespace(trackNamespacePrefix moqtmessage.TrackNamespacePrefix) error {
	/*
	 * Send a UNSUBSCRIBE_NAMESPACE massage
	 */
	usns := moqtmessage.UnsubscribeNamespace{
		TrackNamespacePrefix: trackNamespacePrefix,
	}

	_, err := sess.controlStream.Write(usns.Serialize())

	return err
}

func (sess *SubscribingSession) CancelAnnounce(trackNamespace moqtmessage.TrackNamespace, cancelErr AnnounceCancelError) error {
	// Send an ANNOUNCE_CANCEL message
	acm := moqtmessage.AnnounceCancelMessage{
		TrackNamespace: trackNamespace,
		ErrorCode:      cancelErr.Code(),
		Reason:         cancelErr.Reason(),
	}

	_, err := sess.controlStream.Write(acm.Serialize())

	return err
}

// TODO: this has not implemented yet
func (sess *SubscribingSession) RequestTrackStatus(tns moqtmessage.TrackNamespace, tn string) (*TrackStatus, error) {
	// Send a TRACK_STATUS_REQUEST message
	tsq := moqtmessage.TrackStatusRequest{
		TrackNamespace: tns,
		TrackName:      tn,
	}

	_, err := sess.controlStream.Write(tsq.Serialize())
	if err != nil {
		return nil, err
	}

	// Receive a TRACK_STATUS message
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return nil, err
	}

	if id != moqtmessage.TRACK_STATUS {
		return nil, ErrProtocolViolation
	}

	var tsm moqtmessage.TrackStatusMessage
	err = tsm.DeserializePayload(sess.controlReader)
	if err != nil {
		return nil, err
	}

	ts := TrackStatus{
		Code:         tsm.Code,
		LastGroupID:  tsm.LastGroupID,
		LastObjectID: tsm.LastObjectID,
	}

	return &ts, nil
}

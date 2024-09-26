package moqtransport

import (
	"context"
	"errors"
	"go-moq/moqtransport/moqtmessage"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

var defaultSubscribingSessionIDCounter uint64

type subscribingSessionID uint64

func nextSubscribingSessionID() subscribingSessionID {
	return subscribingSessionID(atomic.AddUint64(&defaultSubscribingSessionIDCounter, 1))
}

type Announcement struct {
	trackNamespace moqtmessage.TrackNamespace

	AuthorizationInfo string
}

/*
 *
 *
 */

type SubscribingSession struct {
	mu sync.RWMutex

	subscribingSessionID

	sessionCore

	/*
	 * ANNOUNCE messages received from the publisher
	 */
	announcements []Announcement

	trackAliasMap trackAliasMap

	subscriptions map[moqtmessage.SubscribeID]Subscription

	relayConfig map[moqtmessage.SubscribeID]struct {
		deliveryTimeout time.Duration
	}

	/*
	 * State of the subscribing content
	 * Initialized when sending new SUBSCRIBE message
	 * Updated when receiving contents or when receiving SUBSCRIBE_OK, SUBSCRIBE_DONE message
	 */
	contentStatuses map[moqtmessage.SubscribeID]*contentStatus

	expiries map[moqtmessage.SubscribeID]*struct {
		expireCtx    context.Context
		expireCancel context.CancelFunc
	}

	//TODO
	readerMap map[moqtmessage.SubscribeID]chan struct {
		moqtmessage.StreamHeader
		quicvarint.Reader
	}
}

func (sess *SubscribingSession) init() {
	sess.readerMap = make(map[moqtmessage.SubscribeID]chan struct {
		moqtmessage.StreamHeader
		quicvarint.Reader
	})

	go func() {
		for {
			stream, err := sess.trSess.AcceptUniStream(context.TODO())
			if err != nil {
				log.Println(err)
				return
			}

			qvReader := quicvarint.NewReader(stream)

			header, err := moqtmessage.DeserializeStreamHeader(qvReader)
			if err != nil {
				log.Print(err)
				return
			}

			if sess.readerMap[header.GetSubscribeID()] == nil {
				sess.readerMap[header.GetSubscribeID()] = make(chan struct {
					moqtmessage.StreamHeader
					quicvarint.Reader
				}, 1<<1)
			}

			sess.readerMap[header.GetSubscribeID()] <- struct {
				moqtmessage.StreamHeader
				quicvarint.Reader
			}{
				StreamHeader: header,
				Reader:       qvReader,
			}
		}
	}()
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
	}

	trackManager.addAnnouncement(am)

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

func (sess *SubscribingSession) Subscribe(announcement Announcement, trackName string, config SubscribeConfig) (ReceiveDataStream, error) {
	subscription, err := sess.subscribe(announcement, trackName, config)
	if err != nil {
		return nil, err
	}

	reader := <-sess.readerMap[subscription.subscribeID]

	log.Println("REACH!! receive a reader through channel")

	switch header := reader.StreamHeader.(type) {
	case *moqtmessage.StreamHeaderDatagram:

		readerCh := make(chan struct {
			moqtmessage.StreamHeaderDatagram
			quicvarint.Reader
		}, 1<<1)

		readerCh <- struct {
			moqtmessage.StreamHeaderDatagram
			quicvarint.Reader
		}{
			StreamHeaderDatagram: *header,
			Reader:               reader.Reader,
		}

		// Continue to receive stream readers in a goroutine
		go func() {
			for {
				reader := <-sess.readerMap[subscription.subscribeID]
				header, ok := reader.StreamHeader.(*moqtmessage.StreamHeaderDatagram)
				// Ignore the header if it is an unexpected header
				if !ok {
					continue
				}

				readerCh <- struct {
					moqtmessage.StreamHeaderDatagram
					quicvarint.Reader
				}{
					StreamHeaderDatagram: *header,
					Reader:               reader.Reader,
				}
			}
		}()

		return &receiveDataStreamDatagram{
			closed:   false,
			header:   *header,
			groupID:  0,
			objectID: 0,
			readerCh: readerCh,
		}, nil

	case *moqtmessage.StreamHeaderTrack:
		readerCh := make(chan struct {
			moqtmessage.StreamHeaderTrack
			quicvarint.Reader
		}, 1<<1)

		readerCh <- struct {
			moqtmessage.StreamHeaderTrack
			quicvarint.Reader
		}{
			StreamHeaderTrack: *header,
			Reader:            reader.Reader,
		}

		// Continue to receive stream readers in a goroutine
		go func() {
			for {
				reader := <-sess.readerMap[subscription.subscribeID]
				header, ok := reader.StreamHeader.(*moqtmessage.StreamHeaderTrack)
				// Ignore the header if it is an unexpected header
				if !ok {
					continue
				}

				readerCh <- struct {
					moqtmessage.StreamHeaderTrack
					quicvarint.Reader
				}{
					StreamHeaderTrack: *header,
					Reader:            reader.Reader,
				}
			}
		}()

		return &receiveDataStreamTrack{
			closed:   false,
			header:   *header,
			groupID:  0,
			objectID: 0,

			readerCh: readerCh,
		}, nil

	case *moqtmessage.StreamHeaderPeep:
		readerCh := make(chan struct {
			moqtmessage.StreamHeaderPeep
			quicvarint.Reader
		}, 1<<1)

		readerCh <- struct {
			moqtmessage.StreamHeaderPeep
			quicvarint.Reader
		}{
			StreamHeaderPeep: *header,
			Reader:           reader.Reader,
		}

		// Continue to receive stream readers in a goroutine
		go func() {
			for {
				reader := <-sess.readerMap[subscription.subscribeID]
				header, ok := reader.StreamHeader.(*moqtmessage.StreamHeaderPeep)
				// Ignore the header if it is an unexpected header
				if !ok {
					continue
				}

				readerCh <- struct {
					moqtmessage.StreamHeaderPeep
					quicvarint.Reader
				}{
					StreamHeaderPeep: *header,
					Reader:           reader.Reader,
				}
			}
		}()

		return &receiveDataStreamPeep{
			closed:   false,
			header:   *header,
			objectID: 0,

			readerCh: readerCh,
		}, nil
	default:
		return nil, ErrProtocolViolation
	}
}

func (sess *SubscribingSession) subscribe(announcement Announcement, trackName string, config SubscribeConfig) (*Subscription, error) {
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
	newSubscribeID := moqtmessage.SubscribeID(len(sess.contentStatuses))

	// Get the Track Alias
	alias := sess.trackAliasMap.getAlias(announcement.trackNamespace, trackName)

	// Initialize the subscription status
	sess.contentStatuses[newSubscribeID] = &contentStatus{
		contentExist:    false,
		largestGroupID:  0,
		largestObjectID: 0,
	}

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

	err = sess.receiveSubscribeResponce(subscription)
	if err != nil {
		retryErr, ok := err.(RetryTrackAliasError)
		if ok {
			sm.TrackAlias = retryErr.trackAlias
			ctx, _ := context.WithTimeout(context.Background(), 30*time.Second) //TODO

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
		err = so.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		if so.SubscribeID != subscription.subscribeID {
			return ErrProtocolViolation
		}

		if so.GroupOrder != subscription.config.GroupOrder {
			return ErrProtocolViolation
		}

		//
		timeout, ok := so.Parameters.DeliveryTimeout()
		if ok {
			if subscription.config.DeliveryTimeout != 0 && subscription.config.DeliveryTimeout != timeout {
				return errors.New("unexpected delivery timeout")
			}
			sess.relayConfig[subscription.subscribeID] = struct {
				deliveryTimeout time.Duration
			}{
				deliveryTimeout: timeout,
			}
		}

		var ctx context.Context
		var cancel context.CancelFunc

		// If a expiry is specified, create the context with the expiry
		if so.Expires != 0 {
			ctx, cancel = context.WithTimeout(context.Background(), so.Expires)
		}

		// Update the subscription status
		sess.contentStatuses[so.SubscribeID] = &contentStatus{
			contentExist:    so.ContentExists,
			largestGroupID:  so.LargestGroupID,
			largestObjectID: so.LargestObjectID,
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
		err = se.DeserializeBody(sess.controlReader)
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

	err = sdm.DeserializeBody(sess.controlReader)
	if err != nil {
		return err
	}

	if sdm.SubscribeID != subscribeID {
		return ErrProtocolViolation
	}

	if sdm.StatusCode != moqtmessage.SUBSCRIBE_DONE_UNSUBSCRIBED {

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
	switch id {
	case moqtmessage.SUBSCRIBE_NAMESPACE_OK:
		var sno moqtmessage.SubscribeNamespaceOkMessage

		err := sno.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		if strings.Join(sno.TrackNamespacePrefix, "") != strings.Join(trackNamespacePrefix, "") {
			return errors.New("unexpected track namespace prefix")
		}

		return nil
	case moqtmessage.SUBSCRIBE_NAMESPACE_ERROR:
		var sne moqtmessage.SubscribeNamespaceErrorMessage

		err := sne.DeserializeBody(sess.controlReader)
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

func (sess *SubscribingSession) requestTrackStatus(tns moqtmessage.TrackNamespace, tn string) (*TrackStatus, error) {
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
	err = tsm.DeserializeBody(sess.controlReader)
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

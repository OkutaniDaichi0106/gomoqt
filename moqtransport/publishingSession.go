package moqtransport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

/*
 *
 */
type PublishingSession struct {
	mu sync.Mutex

	sessionCore

	maxSubscribeID moqtmessage.SubscribeID

	subscriptions map[moqtmessage.SubscribeID]*Subscription

	/*
	 * Track Namespace - Track Name - Track Alias map
	 */
	trackAliasMap trackAliasMap

	//contentStatuses map[moqtmessage.TrackAlias]*contentStatus
}

func (sess *PublishingSession) OpenUniStream() (SendByteStream, error) {
	return sess.trSess.OpenUniStream()
}

func (sess *PublishingSession) OpenUniStreamSync(ctx context.Context) (SendByteStream, error) {
	return sess.trSess.OpenUniStreamSync(ctx)
}

func (sess *PublishingSession) Announce(trackNamespace moqtmessage.TrackNamespace, config AnnounceConfig) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	/*
	 * Send an ANNOUNCE message
	 */
	am := moqtmessage.AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     make(moqtmessage.Parameters, 2),
	}

	// Add the AUTHORIZATION_INFO parameter
	am.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, strings.Join(config.AuthorizationInfo, ""))

	// Add the MAX_CACHE_DURATION parameter
	am.Parameters.AddParameter(moqtmessage.MAX_CACHE_DURATION, config.MaxCacheDuration)

	_, err := sess.controlStream.Write(am.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Receive an ANNOUNCE_OK message or an ANNOUNCE_ERROR message
	 */

	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}

	switch id {
	case moqtmessage.ANNOUNCE_OK:
		var ao moqtmessage.AnnounceOkMessage
		err = ao.DeserializePayload(sess.controlReader)
		if err != nil {
			return err
		}

		// Verify if the Track Namespace in the responce is valid
		for i, v := range trackNamespace {
			if v != ao.TrackNamespace[i] {
				return errors.New("unexpected Track Namespace")
			}
		}

		// Register the Track Namespace
		trackManager.newTrackNamespace(trackNamespace)

		return nil
	case moqtmessage.ANNOUNCE_ERROR:
		var ae moqtmessage.AnnounceErrorMessage // TODO: Handle Error Code
		err = ae.DeserializePayload(sess.controlReader)
		if err != nil {
			return err
		}

		// Verify the Track Namespace in the responce
		for i, v := range trackNamespace {
			if v != ae.TrackNamespace[i] {
				return errors.New("unexpected Track Namespace")
			}
		}

		return errors.New(fmt.Sprint(ae.Code, ae.Reason))

	default:
		return ErrProtocolViolation
	}
}

func (sess *PublishingSession) Unannounce(trackNamespace moqtmessage.TrackNamespace) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	/*
	 * Send an UNANNOUNCE message
	 */
	um := moqtmessage.UnannounceMessage{
		TrackNamespace: trackNamespace,
	}

	_, err := sess.controlStream.Write(um.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Remove the Track Namespace
	 */
	err = trackManager.removeTrackNamespace(trackNamespace)
	if err != nil {
		return err
	}

	return nil
}

func (sess *PublishingSession) WaitSubscribe() (*Subscription, error) {
	/*
	 * Get new Subscribe ID
	 */
	newSubscribeID := moqtmessage.SubscribeID(len(sess.subscriptions))
	if newSubscribeID > sess.maxSubscribeID {
		return nil, ErrTooManySubscribes
	}

	/*
	 * Receive SUBSCRIBE message
	 */
	subscription, err := sess.receiveSubscribe(newSubscribeID)
	if err != nil {
		return nil, err
	}

	// Register the subscription
	sess.subscriptions[newSubscribeID] = subscription

	return subscription, nil
}

func (sess *PublishingSession) receiveSubscribe(newSubscribeID moqtmessage.SubscribeID) (*Subscription, error) {
	/*
	 * Receive a SUBSCRIBE message
	 */
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.SUBSCRIBE {
		return nil, ErrProtocolViolation
	}

	sm := moqtmessage.SubscribeMessage{}
	err = sm.DeserializePayload(sess.controlReader)
	if err != nil {
		return nil, err
	}

	// Verify the received Subscribe ID is a new one
	if newSubscribeID != sm.SubscribeID {
		return nil, ErrSubscribeFailed
	}

	_, ok := sess.subscriptions[newSubscribeID]
	if ok {
		return nil, ErrSubscribeFailed
	}

	// Verify the Track Namespace is available
	trackManager.trackNamespaceTree.trace(sm.TrackNamespace)

	// Verify the Track Name already exists
	tnsNode, ok := trackManager.findTrackNamespace(sm.TrackNamespace)
	if !ok {
		// TODO: create a new subscription
		panic("NO TRACKNAMESPACE!! MAKE NEW SUBSCRIPTION!!")
	}

	_, ok = tnsNode.findTrackName(sm.TrackName)

	// Create new Track if the Track Name did not exist
	if !ok {
		tnsNode.newTrackNameNode(sm.TrackName)

		// TODO: create a new subscription
		panic("NO TRACK!! MAKE NEW SUBSCRIPTION!!")
	}

	// Verify if the Track Alias is valid
	existingAlias := sess.trackAliasMap.getAlias(sm.TrackNamespace, sm.TrackName)

	if existingAlias != sm.TrackAlias {
		return nil, RetryTrackAliasError{}
	}

	// Get new Subscribe Config
	config := SubscribeConfig{
		SubscriberPriority: sm.SubscriberPriority,
		GroupOrder:         sm.GroupOrder,
		SubscriptionFilter: sm.SubscriptionFilter,
	}

	// Parse the parameters
	authInfo, ok := sm.Parameters.AuthorizationInfo()
	if ok {
		config.AuthorizationInfo = authInfo
	}

	timeout, ok := sm.Parameters.DeliveryTimeout()
	if ok {
		config.DeliveryTimeout = timeout
	}

	// Get new Subscription
	subscription := Subscription{
		subscribeID:    newSubscribeID,
		trackAlias:     sm.TrackAlias,
		trackNamespace: sm.TrackNamespace,
		trackName:      sm.TrackName,
		config:         config,
	}

	return &subscription, nil
}

func (sess *PublishingSession) AllowSubscribe(subscription *Subscription, expiry time.Duration) error {
	// Find the Track Namespace and verify if it is available
	tnsNode, ok := trackManager.findTrackNamespace(subscription.trackNamespace)
	if !ok {
		return errors.New("track namespace not found")
	}

	// Find the Track Name
	tnNode, ok := tnsNode.findTrackName(subscription.trackName)
	if !ok {
		// Register the Track Name if it did not exist
		tnNode = tnsNode.newTrackNameNode(subscription.trackName)
	}

	/*
	 * Send a SUBSCRIBE_OK message
	 */
	so := moqtmessage.SubscribeOkMessage{
		SubscribeID:     subscription.subscribeID,
		Expires:         expiry,
		GroupOrder:      subscription.config.GroupOrder,
		ContentExists:   tnNode.contentStatus.contentExists,
		LargestGroupID:  tnNode.contentStatus.largestGroupID,
		LargestObjectID: tnNode.contentStatus.largestObjectID,
		Parameters:      make(moqtmessage.Parameters), // TODO: Handler the parameters
	}

	so.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, subscription.config.DeliveryTimeout)

	if len(subscription.config.AuthorizationInfo) > 0 {
		so.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, subscription.config.AuthorizationInfo)
	}

	_, err := sess.controlStream.Write(so.Serialize())
	if err != nil {
		return ErrSubscribeFailed
	}

	return nil
}

func (sess *PublishingSession) RejectSubscribe(subscription *Subscription, subscribeError SubscribeError) {
	/*
	 * Send a SUBSCRIBE_ERROR
	 */

	sem := moqtmessage.SubscribeErrorMessage{
		SubscribeID: subscription.subscribeID,
		Code:        subscribeError.Code(),
		Reason:      subscribeError.Error(),
	}

	// Append Track Alias field if the Subscribe Error is the SubscribeRetryTrackAlias
	retryErr, ok := subscribeError.(RetryTrackAliasError)
	if ok {
		sem.TrackAlias = retryErr.trackAlias
	}

	_, err := sess.controlStream.Write(sem.Serialize())
	if err != nil {
		log.Println(err)
	}
}

func (sess *PublishingSession) EndSubscription(subscription Subscription, status SubscribeDoneStatus) error {
	tnNode, ok := trackManager.findTrackName(subscription.trackNamespace, subscription.trackName)
	if !ok {
		return errors.New("track not found")
	}

	// Send a SUBSCRIBE_DONE message
	sd := moqtmessage.SubscribeDoneMessage{
		SubscribeID:   subscription.subscribeID,
		StatusCode:    status.Code(),
		Reason:        status.Reason(),
		ContentExists: tnNode.contentStatus.contentExists,
		FinalGroupID:  subscription.finalGroupID,
		FinalObjectID: subscription.finalObjectID,
	}
	_, err := sess.controlStream.Write(sd.Serialize())
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (sess *PublishingSession) SendTrackStatus() {

}

func (p *PublishingSession) WaitSubscribeNamespace() {
	// Send a SUBSCRIBE_OK
}

func (sess *PublishingSession) AllowSubscribeNamespace() {
	// Send a SUBSCRIBE_OK
}

func (p *PublishingSession) RejectSubscribeNamespace() {
	// Send a SUBSCRIBE_ERROR
}

type Subscription struct {
	subscribeID    moqtmessage.SubscribeID
	trackAlias     moqtmessage.TrackAlias
	trackNamespace moqtmessage.TrackNamespace
	trackName      string
	config         SubscribeConfig

	finalGroupID  moqtmessage.GroupID
	finalObjectID moqtmessage.ObjectID

	forwardingPreference moqtmessage.ObjectForwardingPreference
}

func (s Subscription) TrackName() string {
	return s.trackName
}

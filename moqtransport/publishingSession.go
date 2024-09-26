package moqtransport

import (
	"errors"
	"fmt"
	"go-moq/moqtransport/moqtmessage"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var defaultPublishingSessionIDCounter uint64

type publishingSessionID uint64

func nextPublishingSessionID() publishingSessionID {
	return publishingSessionID(atomic.AddUint64(&defaultPublishingSessionIDCounter, 1))
}

type trackAliasMap struct {
	mu         sync.RWMutex
	nameIndex  map[string]map[string]moqtmessage.TrackAlias
	aliasIndex map[moqtmessage.TrackAlias]struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}
}

func (tamap *trackAliasMap) getAlias(tns moqtmessage.TrackNamespace, tn string) moqtmessage.TrackAlias {
	/*
	 * If an Track Alias exists, return the existing Track Alias
	 */
	tamap.mu.RLock()

	existingAlias, ok := tamap.nameIndex[strings.Join(tns, "")][tn]
	if ok {
		return existingAlias
	}

	tamap.mu.RUnlock()

	/*
	 * If no Track Alias was found, get new Track Alias and register the Track Namespace and Track Name with it
	 */
	tamap.mu.Lock()

	newAlias := moqtmessage.TrackAlias(len(tamap.aliasIndex))

	tamap.nameIndex[strings.Join(tns, "")][tn] = newAlias

	tamap.aliasIndex[newAlias] = struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}{
		trackNamespace: tns,
		trackName:      tn,
	}

	tamap.mu.Lock()

	return newAlias
}

func (tamap *trackAliasMap) getName(ta moqtmessage.TrackAlias) (moqtmessage.TrackNamespace, string, bool) {
	data, ok := tamap.aliasIndex[ta]
	if !ok {
		return nil, "", false
	}

	return data.trackNamespace, data.trackName, true
}

type contentStatus struct {
	/*
	 * Current state of the contents
	 */
	contentExist  bool
	finalGroupID  moqtmessage.GroupID
	finalObjectID moqtmessage.ObjectID

	/*
	 * The Largest Group ID and Largest Object ID of the contents received so far
	 */
	largestGroupID  moqtmessage.GroupID
	largestObjectID moqtmessage.ObjectID
}

/*
 *
 */
type PublishingSession struct {
	publishingSessionID

	sessionCore

	maxSubscribeID moqtmessage.SubscribeID

	subscriptions map[moqtmessage.SubscribeID]*Subscription

	forwardingPreference map[string]map[string]moqtmessage.ObjectForwardingPreference

	/*
	 * Track Namespace - Track Name - Track Alias map
	 */
	trackAliasMap trackAliasMap

	contentStatuses map[moqtmessage.TrackAlias]*contentStatus
}

func (sess *PublishingSession) Announce(trackNamespace moqtmessage.TrackNamespace, config AnnounceConfig) error {
	/*
	 * Send ANNOUNCE message
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
		err = ao.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		// Verify if the Track Namespace in the responce is valid
		for i, v := range trackNamespace {
			if v != ao.TrackNamespace[i] {
				return errors.New("unexpected Track Namespace")
			}
		}

		// Register the ANNOUNCE message and the Track Namespace
		trackManager.addAnnouncement(am)

		return nil
	case moqtmessage.ANNOUNCE_ERROR:
		var ae moqtmessage.AnnounceErrorMessage // TODO: Handle Error Code
		err = ae.DeserializeBody(sess.controlReader)
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
	um := moqtmessage.UnannounceMessage{
		TrackNamespace: trackNamespace,
	}

	_, err := sess.controlStream.Write(um.Serialize())
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

	var onceErr *sync.Once
	// If some error exists, reject the subscription
	if err != nil {
		switch someErr := err.(type) {
		case SubscribeError:
			onceErr.Do(func() {
				sess.RejectSubscribe(newSubscribeID, someErr)
			})
		case TerminateError:
			sess.Terminate(someErr)
			return nil, err
		default:
			onceErr.Do(func() {
				sess.RejectSubscribe(newSubscribeID, ErrSubscribeFailed)
			})
		}
	}

	// Register the subscription
	sess.subscriptions[newSubscribeID] = subscription

	_, ok := sess.contentStatuses[subscription.trackAlias]
	if !ok {
		sess.contentStatuses[subscription.trackAlias] = &contentStatus{}
	}

	return subscription, nil
}

func (sess *PublishingSession) receiveSubscribe(newSubscribeID moqtmessage.SubscribeID) (*Subscription, error) {
	// Receive a SUBSCRIBE message
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.SUBSCRIBE {
		return nil, ErrProtocolViolation
	}

	sm := moqtmessage.SubscribeMessage{}
	err = sm.DeserializeBody(sess.controlReader)
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

	tnsNode.mu.RLock()
	defer tnsNode.mu.RUnlock()
	tnNode, ok := tnsNode.tracks[sm.TrackName]

	// Create new Track if the Track Name did not exist
	if !ok {
		tnsNode.mu.Lock()
		tnsNode.tracks[sm.TrackName] = &trackNameNode{
			value:                 sm.TrackName,
			sessionWithSubscriber: make(map[publishingSessionID]*PublishingSession),
		}

		tnNode = tnsNode.tracks[sm.TrackName]

		tnsNode.mu.Unlock()

		// TODO: create a new subscription
		panic("NO TRACK!! MAKE NEW SUBSCRIPTION!!")
	}

	// Register the session
	tnNode.sessionWithSubscriber[sess.publishingSessionID] = sess

	// Verify if the Track Alias is valid
	existingAlias := sess.trackAliasMap.getAlias(sm.TrackNamespace, sm.TrackName)

	if existingAlias != sm.TrackAlias {
		return nil, RetryTrackAliasError{}
	}

	// Parse the parameters
	authInfo, _ := sm.Parameters.AuthorizationInfo()

	timeout, _ := sm.Parameters.DeliveryTimeout()

	// Get new Subscribe Config
	config := SubscribeConfig{
		SubscriberPriority: sm.SubscriberPriority,
		GroupOrder:         sm.GroupOrder,
		SubscriptionFilter: sm.SubscriptionFilter,
		AuthorizationInfo:  authInfo,
		DeliveryTimeout:    timeout,
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
	var contentExists bool

	// Send a SUBSCRIBE_OK
	so := moqtmessage.SubscribeOkMessage{
		SubscribeID:     subscription.subscribeID,
		Expires:         expiry,
		GroupOrder:      subscription.config.GroupOrder,
		ContentExists:   contentExists,
		LargestGroupID:  sess.contentStatuses[subscription.trackAlias].largestGroupID,
		LargestObjectID: sess.contentStatuses[subscription.trackAlias].largestObjectID,
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

func (sess *PublishingSession) RejectSubscribe(subscribeID moqtmessage.SubscribeID, subscribeError SubscribeError) {
	/*
	 * Send a SUBSCRIBE_ERROR
	 */

	sem := moqtmessage.SubscribeErrorMessage{
		SubscribeID: subscribeID,
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

func (sess *PublishingSession) EndSubscription(subscribeID moqtmessage.SubscribeID, status SubscribeDoneStatus) error {
	// Get subscription information from Subscribe ID
	subscription, ok := sess.subscriptions[subscribeID]
	if !ok {
		return errors.New("subscription not found")
	}

	// Get current publishment information from Track Alias
	contentStatus, ok := sess.contentStatuses[subscription.trackAlias]
	if !ok {
		return errors.New("invalid subscripe ID")
	}

	// Send a SUBSCRIBE_DONE message
	sd := moqtmessage.SubscribeDoneMessage{
		SubscribeID:   subscribeID,
		StatusCode:    status.Code(),
		Reason:        status.Reason(),
		ContentExists: contentStatus.contentExist,
		FinalGroupID:  contentStatus.finalGroupID,
		FinalObjectID: contentStatus.finalObjectID,
	}
	_, err := sess.controlStream.Write(sd.Serialize())
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (sess *PublishingSession) sendTrackStatus() {

}

func (sess *PublishingSession) AllowSubscribeNamespace() {
	// Send a SUBSCRIBE_OK
}

func (p *PublishingSession) allowSubscribeNamespace() {
	// Send a SUBSCRIBE_OK
}

func (p *PublishingSession) RejectSubscribeNamespace() {
	// Send a SUBSCRIBE_ERROR
}

func (p *PublishingSession) rejectSubscribeNamespace() {
	// Send a SUBSCRIBE_ERROR
}

type Subscription struct {
	subscribeID    moqtmessage.SubscribeID
	trackAlias     moqtmessage.TrackAlias
	trackNamespace moqtmessage.TrackNamespace
	trackName      string
	config         SubscribeConfig

	//onceErr *sync.Once
}

func (s Subscription) TrackName() string {
	return s.trackName
}

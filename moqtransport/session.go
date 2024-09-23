package moqtransport

import (
	"context"
	"errors"
	"fmt"
	"go-moq/moqtransport/moqtmessage"
	"go-moq/moqtransport/moqtversion"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type sessionID uint64

type sessionCore struct {
	//sessionID sessionID

	trSess TransportSession

	controlStream Stream

	controlReader quicvarint.Reader

	selectedVersion moqtversion.Version
}

func (sess sessionCore) Terminate(err TerminateError) {
	sess.trSess.CloseWithError(SessionErrorCode(err.Code()), err.Error())
}

/*
 *
 */
type PublishingSession struct {
	sessionCore

	maxSubscribeID moqtmessage.SubscribeID

	//announcements []moqtmessage.AnnounceMessage

	subscriptions map[moqtmessage.SubscribeID]*struct {
		onceErr sync.Once
		config  SubscribeConfig
	}

	/*
	 * Track Namespace - Track Alias map
	 */
	aliasFromName map[string]moqtmessage.TrackAlias

	nameFromAlias map[moqtmessage.TrackAlias]struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}

	publishmentState map[moqtmessage.SubscribeID]*struct {
		/*
		 * Current state of the contents
		 */
		contenExist   bool
		finalGroupID  moqtmessage.GroupID
		finalObjectID moqtmessage.ObjectID

		/*
		 * The Largest Group ID and Largest Object ID of the contents received so far
		 */
		largestGroupID  moqtmessage.GroupID
		largestObjectID moqtmessage.ObjectID
	}
}

func (sess *PublishingSession) Announce(trackNamespace moqtmessage.TrackNamespace, config AnnounceConfig) error {
	// Send ANNOUNCE message
	am := moqtmessage.AnnounceMessage{
		TrackNamespace: trackNamespace,
		Parameters:     make(moqtmessage.Parameters, 2),
	}

	// Add the AUTHORIZATION_INFO parameter
	am.AddParameter(moqtmessage.AUTHORIZATION_INFO, strings.Join(config.AuthorizationInfo, ""))

	// Add the MAX_CACHE_DURATION parameter
	am.AddParameter(moqtmessage.MAX_CACHE_DURATION, config.MaxCacheDuration)

	_, err := sess.controlStream.Write(am.Serialize())
	if err != nil {
		return err
	}

	// Receive an ANNOUNCE_OK message or an ANNOUNCE_ERROR message
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

	// Register the ANNOUNCE message and the Track Namespace
	trackManager.addAnnouncement(am)

	return nil
}

func (sess *PublishingSession) unannounce(trackNamespace moqtmessage.TrackNamespace) error {
	um := moqtmessage.UnannounceMessage{
		TrackNamespace: trackNamespace,
	}

	_, err := sess.controlStream.Write(um.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (sess *PublishingSession) WaitSubscribe() (SubscribeConfig, error) {
	newSubscribeID := moqtmessage.SubscribeID(len(sess.subscriptions))

	// Receive SUBSCRIBE message
	err := sess.receiveSubscribe(newSubscribeID)

	// If any error exists, reject the subscription
	if err != nil {
		newSubscription, ok := sess.subscriptions[newSubscribeID]
		if !ok {
			return SubscribeConfig{}, err
		}

		switch someErr := err.(type) {
		case SubscribeError:
			newSubscription.onceErr.Do(func() {
				sess.RejectSubscribe(someErr)
			})
		case TerminateError:
			sess.Terminate(someErr)
		default:
		}
		return SubscribeConfig{}, err
	}

	return sess.subscriptions[newSubscribeID].config, nil
}

func (sess *PublishingSession) receiveSubscribe(newSubscribeID moqtmessage.SubscribeID) error {
	// Receive a SUBSCRIBE message
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}
	if id != moqtmessage.SUBSCRIBE {
		return ErrProtocolViolation
	}

	sm := moqtmessage.SubscribeMessage{}
	err = sm.DeserializeBody(sess.controlReader)
	if err != nil {
		return err
	}

	// Verify the received Subscribe ID is a new one
	if newSubscribeID != sm.SubscribeID {
		return ErrDefaultSubscribeFailed.NewSubscribeID(sm.SubscribeID)
	}

	_, ok := sess.subscriptions[newSubscribeID]
	if ok {
		return ErrDefaultSubscribeFailed.NewSubscribeID(sm.SubscribeID)
	}

	// Verify the Track Namespace is available
	trackManager.trackNamespaceTree.trace(sm.TrackNamespace)

	// Verify the  Track Name already exists
	tnsNode, ok := trackManager.trackNamespaceTree.trace(sm.TrackNamespace)
	if !ok {
		log.Println(err)
		// TODO: create a new subscription
		panic("NO TRACKNAMESPACE!! MAKE NEW SUBSCRIPTION!!")
	}

	tnsNode.mu.RLock()
	tnNode, ok := tnsNode.tracks[sm.TrackName]
	tnsNode.mu.RUnlock()

	// Create new Track if the Track Name did not exist
	if !ok {
		tnsNode.mu.Lock()
		tnsNode.tracks[sm.TrackName] = &trackNameNode{
			value:                 sm.TrackName,
			sessionWithSubscriber: []*PublishingSession{sess},
		}
		tnsNode.mu.Unlock()

		// TODO: create a new subscription
		panic("NO TRACK!! MAKE NEW SUBSCRIPTION!!")
	}

	// Register the session
	tnNode.sessionWithSubscriber = append(tnNode.sessionWithSubscriber, sess)

	// Verify if the Track Alias is already in use
	existingAlias, ok := sess.aliasFromName[strings.Join(sm.TrackNamespace, "")+sm.TrackName]
	if ok {
		return ErrDefaultRetryTrackAlias.NewTrackAlias(existingAlias).NewSubscribeID(sm.SubscribeID)
	}

	// Register the Track Alias and the the Track Namespace and the Track Name
	sess.aliasFromName[strings.Join(sm.TrackNamespace, "")+sm.TrackName] = sm.TrackAlias
	sess.nameFromAlias[sm.TrackAlias] = struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}{
		trackNamespace: sm.TrackNamespace,
		trackName:      sm.TrackName,
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

	// Register the subscribe configuration
	sess.subscriptions[newSubscribeID] = &struct {
		onceErr sync.Once
		config  SubscribeConfig
	}{
		config: config,
	}

	return nil
}

func (sess *PublishingSession) AllowSubscribe(subscribeID moqtmessage.SubscribeID, expiry time.Duration) error {
	var contentExists bool

	subscription, ok := sess.subscriptions[subscribeID]
	if ok {
		return ErrDefaultSubscribeFailed.NewSubscribeID(0)
	}

	// Send a SUBSCRIBE_OK
	so := moqtmessage.SubscribeOkMessage{
		SubscribeID:     subscribeID,
		Expires:         expiry,
		GroupOrder:      subscription.config.GroupOrder,
		ContentExists:   contentExists,
		LargestGroupID:  sess.publishmentState[subscribeID].finalGroupID,
		LargestObjectID: sess.publishmentState[subscribeID].finalObjectID,
		Parameters:      make(moqtmessage.Parameters), // TODO: Handler the parameters
	}

	so.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, subscription.config.DeliveryTimeout)

	if len(subscription.config.AuthorizationInfo) > 0 {
		so.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, subscription.config.AuthorizationInfo)
	}

	_, err := sess.controlStream.Write(so.Serialize())
	if err != nil {
		return ErrDefaultSubscribeFailed.NewSubscribeID(subscribeID)
	}

	return nil
}

func (sess *PublishingSession) RejectSubscribe(subscribeErr SubscribeError) {
	sess.rejectSubscribe(subscribeErr)
}

func (sess *PublishingSession) rejectSubscribe(se SubscribeError) {
	// Send a SUBSCRIBE_ERROR
	sem := moqtmessage.SubscribeErrorMessage{
		SubscribeID: se.SubscribeID(),
		Code:        se.Code(),
		Reason:      se.Error(),
	}

	// Append Track Alias field if the Subscribe Error is the SubscribeRetryTrackAlias
	if srta, ok := se.(RetryTrackAlias); ok {
		sem.TrackAlias = srta.newTrackAlias
	}

	_, err := sess.controlStream.Write(sem.Serialize())
	if err != nil {
		log.Println(err)
	}
}

var publishing map[moqtmessage.SubscribeID]struct {
	contentExists bool
	finalGroupID  moqtmessage.GroupID
	finalObjectID moqtmessage.ObjectID
}

func (sess *PublishingSession) endSubscription(id moqtmessage.SubscribeID, status SubscribeDoneStatus) {
	data, _ := publishing[id]
	// Send a SUBSCRIBE_DONE message
	sd := moqtmessage.SubscribeDoneMessage{
		SubscribeID:   id,
		StatusCode:    status.Code(),
		Reason:        status.Reason(),
		ContentExists: data.contentExists,
		FinalGroupID:  data.finalGroupID,
		FinalObjectID: data.finalObjectID,
	}
	_, err := sess.controlStream.Write(sd.Serialize())
	if err != nil {
		log.Println(err)
	}
}

func (p *PublishingSession) sendTrackStatus() {

}

func (p *PublishingSession) acceptSubscribeNamespace() {
	// Send a SUBSCRIBE_OK
}

func (p *PublishingSession) rejectSubscribeNamespace() {
	// Send a SUBSCRIBE_ERROR
}

func (p *PublishingSession) sendSingleObject([]byte) {

}

func (p *PublishingSession) sendMultipleObject(io.Reader) {

}

func (p *PublishingSession) sendDatagram([]byte) {
}

/*
 *
 *
 */
type SubscribingSession struct {
	sessionCore

	/*
	 * ANNOUNCE messages received from the publisher
	 */
	announcements []moqtmessage.AnnounceMessage

	/*
	 * Track Alias - Track Namespace mapping
	 */
	nameFromAlias map[moqtmessage.TrackAlias]struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}

	/*
	 * Track Namespace - Track Alias mapping
	 */
	aliasFromName map[string]moqtmessage.TrackAlias

	/*
	 * State of the subscribing content
	 * Initialized when sending new SUBSCRIBE message
	 * Updated when receiving contents or when receiving SUBSCRIBE_OK, SUBSCRIBE_DONE message
	 */
	subscriptionStatus map[moqtmessage.SubscribeID]*struct {
		config          SubscribeConfig
		contentExist    bool
		largestGroupID  moqtmessage.GroupID
		largestObjectID moqtmessage.ObjectID

		expireCtx    context.Context
		expireCancel context.CancelFunc
	}
}

func (sess SubscribingSession) AllowAnnounce(trackNamespace moqtmessage.TrackNamespace) error {
	// Send an ANNOUNCE_OK message
	ao := moqtmessage.AnnounceOkMessage{
		TrackNamespace: trackNamespace,
	}

	_, err := sess.controlStream.Write(ao.Serialize())

	return err
}

func (sess SubscribingSession) RejectAnnounce(trackNamespace moqtmessage.TrackNamespace, annErr AnnounceError) {
	// Send an ANNOUNCE_ERROR message
	ae := moqtmessage.AnnounceErrorMessage{
		TrackNamespace: trackNamespace,
		Code:           annErr.Code(),
		Reason:         annErr.Error(),
	}

	_, err := sess.controlStream.Write(ae.Serialize())
	if err != nil {
		log.Println(err)
		return
	}
}

func (sess SubscribingSession) Subscribe(trackNamespace moqtmessage.TrackNamespace, trackName string, config SubscribeConfig) error {
	return sess.subscribe(trackNamespace, trackName, config)
}

func (sess SubscribingSession) subscribe(tns moqtmessage.TrackNamespace, tn string, config SubscribeConfig) error {
	// Set the default group order value, if the value is 0
	if config.GroupOrder == 0 {
		config.GroupOrder = moqtmessage.ASCENDING
	}

	// Set the default filter code, if the value is 0
	if config.FilterCode == 0 {
		config.FilterCode = moqtmessage.LATEST_GROUP
	}

	// Get new Subscribe ID
	newSubscribeID := moqtmessage.SubscribeID(len(sess.subscriptionStatus))

	// Get Track Alias using Full Track Name
	fullTrackName := strings.Join(tns, "") + tn
	trackAlias, ok := sess.aliasFromName[fullTrackName]
	if !ok {
		// Create new Track Alias if it did not exist
		trackAlias = moqtmessage.TrackAlias(len(sess.aliasFromName))

		_, ok := sess.nameFromAlias[trackAlias]
		if ok {
			return ErrDuplicatedTrackAlias
		}

		sess.nameFromAlias[trackAlias] = struct {
			trackNamespace moqtmessage.TrackNamespace
			trackName      string
		}{
			trackNamespace: tns,
			trackName:      tn,
		}
	}

	// Initialize the subscription status
	sess.subscriptionStatus[newSubscribeID] = &struct {
		config          SubscribeConfig
		contentExist    bool
		largestGroupID  moqtmessage.GroupID
		largestObjectID moqtmessage.ObjectID
		expireCtx       context.Context
		expireCancel    context.CancelFunc
	}{
		config:          config,
		contentExist:    false,
		largestGroupID:  0,
		largestObjectID: 0,
	}

	/*
	 * Send a SUBSCRIBE message
	 */
	sm := moqtmessage.SubscribeMessage{
		SubscribeID:        newSubscribeID,
		TrackAlias:         trackAlias,
		TrackNamespace:     tns,
		TrackName:          tn,
		SubscriberPriority: config.SubscriberPriority,
		GroupOrder:         config.GroupOrder,
		SubscriptionFilter: config.SubscriptionFilter,
		Parameters:         make(moqtmessage.Parameters),
	}

	// Add the authorization information parameter
	sm.Parameters.AddParameter(moqtmessage.AUTHORIZATION_INFO, config.AuthorizationInfo)

	// Add the delivery timeout parameter
	sm.Parameters.AddParameter(moqtmessage.DELIVERY_TIMEOUT, config.DeliveryTimeout)

	_, err := sess.controlStream.Write(sm.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Receive a SUBSCRIBE_OK message or a SUBSCRIBE_ERROR message
	 */
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}

	switch id {
	case moqtmessage.SUBSCRIBE_OK:
		var ao moqtmessage.SubscribeOkMessage
		err = ao.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		authInfo, err := ao.Parameters.AuthorizationInfo()
		if err != nil {
			return err
		}

		timeout, err := ao.Parameters.DeliveryTimeout()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), ao.Expires)

		// Update the subscription status
		sess.subscriptionStatus[ao.SubscribeID] = &struct {
			config          SubscribeConfig
			contentExist    bool
			largestGroupID  moqtmessage.GroupID
			largestObjectID moqtmessage.ObjectID
			expireCtx       context.Context
			expireCancel    context.CancelFunc
		}{
			config: SubscribeConfig{
				GroupOrder:        ao.GroupOrder, // TODO: Verify the existing Group Order
				AuthorizationInfo: authInfo,
				DeliveryTimeout:   timeout,
			},
			contentExist:    ao.ContentExists,
			largestGroupID:  ao.LargestGroupID,
			largestObjectID: ao.LargestObjectID,
			expireCtx:       ctx,
			expireCancel:    cancel,
		}

		return nil
	case moqtmessage.SUBSCRIBE_ERROR:
		var ae moqtmessage.SubscribeErrorMessage // TODO: Handle Error Code
		err = ae.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		if ae.SubscribeID != newSubscribeID {
			return ErrProtocolViolation
		}

		switch ae.Code {
		case moqtmessage.SUBSCRIBE_INTERNAL_ERROR:
			return ErrDefaultSubscribeFailed.NewSubscribeID(newSubscribeID)
		case moqtmessage.INVALID_RANGE:
			return ErrDefaultInvalidRange.NewSubscribeID(newSubscribeID)
		case moqtmessage.RETRY_TRACK_ALIAS:
			// Verify the given Track Alias is a new one
			_, ok := sess.nameFromAlias[ae.TrackAlias]
			if ok {
				return ErrDuplicatedTrackAlias
			}

			// Set the given Track Alias and Send SUBSCRIBE message
			sm.TrackAlias = ae.TrackAlias

			ctx, _ := context.WithTimeout(context.Background(), 30*time.Second) // TODO: Handle the duration

			return sess.retrySubscribe(sm, ctx)
		default:
			return fmt.Errorf(ae.Reason, ae.Code)
		}

	default:
		return ErrProtocolViolation
	}
}

func (sess SubscribingSession) retrySubscribe(sm moqtmessage.SubscribeMessage, ctx context.Context) error {
	/*
	 * Send a SUBSCRIBE message
	 */
	_, err := sess.controlStream.Write(sm.Serialize())
	if err != nil {
		return err
	}

	/*
	 * Receive a SUBSCRIBE_OK message or a SUBSCRIBE_ERROR message
	 */
	id, err := moqtmessage.DeserializeMessageID(sess.controlReader)
	if err != nil {
		return err
	}

	switch id {
	case moqtmessage.SUBSCRIBE_OK:
		var ao moqtmessage.SubscribeOkMessage
		err = ao.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		authInfo, err := ao.Parameters.AuthorizationInfo()
		if err != nil {
			return err
		}

		timeout, err := ao.Parameters.DeliveryTimeout()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), ao.Expires)

		// Update the subscription status
		sess.subscriptionStatus[ao.SubscribeID] = &struct {
			config          SubscribeConfig
			contentExist    bool
			largestGroupID  moqtmessage.GroupID
			largestObjectID moqtmessage.ObjectID
			expireCtx       context.Context
			expireCancel    context.CancelFunc
		}{
			config: SubscribeConfig{
				GroupOrder:        ao.GroupOrder, // TODO: Verify the existing Group Order
				AuthorizationInfo: authInfo,
				DeliveryTimeout:   timeout,
			},
			contentExist:    ao.ContentExists,
			largestGroupID:  ao.LargestGroupID,
			largestObjectID: ao.LargestObjectID,
			expireCtx:       ctx,
			expireCancel:    cancel,
		}

		return nil
	case moqtmessage.SUBSCRIBE_ERROR:
		var ae moqtmessage.SubscribeErrorMessage // TODO: Handle Error Code
		err = ae.DeserializeBody(sess.controlReader)
		if err != nil {
			return err
		}

		if ae.SubscribeID != sm.SubscribeID {
			return ErrProtocolViolation
		}

		switch ae.Code {
		case moqtmessage.SUBSCRIBE_INTERNAL_ERROR:
			return ErrDefaultSubscribeFailed.NewSubscribeID(sm.SubscribeID)
		case moqtmessage.INVALID_RANGE:
			return ErrDefaultInvalidRange.NewSubscribeID(sm.SubscribeID)
		case moqtmessage.RETRY_TRACK_ALIAS:
			// Verify the given Track Alias is a new one
			_, ok := sess.nameFromAlias[ae.TrackAlias]
			if ok {
				return ErrDuplicatedTrackAlias
			}

			// Set the given Track Alias and Send SUBSCRIBE message
			sm.TrackAlias = ae.TrackAlias

			return sess.retrySubscribe(sm, ctx)
		default:
			return fmt.Errorf(ae.Reason, ae.Code)
		}

	default:
		return ErrProtocolViolation
	}
}

func (SubscribingSession) subscribeUpdate() {

}
func (SubscribingSession) unsubscribe() {

}
func (SubscribingSession) subscribeNamespace() {

}
func (SubscribingSession) unsubscribeNamespace() {

}

func (SubscribingSession) cancelAnnounce() {
	// Send an ANNOUNCE_CANCEL message
}

func (sess SubscribingSession) requestTrackStatus(tns moqtmessage.TrackNamespace, tn string) (*TrackStatus, error) {
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

func (sess SubscribingSession) deliverObjects(stream ObjectStream) {
	switch header := stream.Header().(type) {
	case *moqtmessage.StreamHeaderTrack:
	case *moqtmessage.StreamHeaderPeep:
		newState := &struct {
			contenExist   bool
			finalGroupID  moqtmessage.GroupID
			finalObjectID moqtmessage.ObjectID
		}{
			contenExist:  true,
			finalGroupID: header.GroupID,
		}
		sess.subscriptionStatus[header.SubscribeID] = newState

		//TODO: Increment the Object ID

	}
}

type PubSubSession struct {
	sessionCore
	localTrackNamespace moqtmessage.TrackNamespace
	//remoteTrackNamespaces map[string]moqtmessage.TrackNamespace

	maxSubscribeID   moqtmessage.SubscribeID
	maxCacheDuration time.Duration

	//trackManager
}

func distribute() {

}

// func (s *session) distribute(src webtransport.ReceiveStream, errCh chan<- error) {
// 	var wg sync.WaitGroup
// 	dataCh := make(chan []byte, 1<<4)
// 	buf := make([]byte, 1<<8)

// 	go func(src webtransport.ReceiveStream) {
// 		for {
// 			n, err := src.Read(buf)
// 			if err != nil {
// 				if err == io.EOF {
// 					// Send final data chunk
// 					dataCh <- buf[:n]
// 					return
// 				}
// 				log.Println(err)
// 				errCh <- err
// 				return
// 			}
// 			// Send data chunk
// 			dataCh <- buf[:n]
// 		}
// 	}(src)

// 	fullTrackNamespace := s.trackNamespace.GetFullName()

// 	dests, ok := subscribers[fullTrackNamespace]
// 	if !ok {
// 		//TODO: handle this as internal error
// 		panic("destinations not found")
// 	}

// 	for _, sess := range dests.sessions {
// 		// Increment the wait group by 1 and count the number of current processes
// 		wg.Add(1)

// 		go func(sess *SubscriberSession) {
// 			defer wg.Done()

// 			// Open a unidirectional stream
// 			dst, err := sess.getWebtransportSession().OpenUniStream()
// 			if err != nil {
// 				log.Println(err)
// 				errCh <- err
// 				return //TODO: handle the error
// 			}

// 			// Close the stream after whole data was sent
// 			defer dst.Close()

// 			for data := range dataCh {
// 				if len(data) == 0 {
// 					continue
// 				}
// 				_, err := dst.Write(data)
// 				if err != nil {
// 					log.Println(err)
// 					errCh <- err
// 					return
// 				}
// 			}
// 		}(sess)
// 	}

// 	// Wait untill the data has been sent to all sessions
// 	wg.Wait()
// }

/*
 *
 */
// var tracks trackManager

// type trackManager struct {
// 	mu            sync.RWMutex
// 	nameFromAlias map[moqtmessage.TrackAlias]string
// 	aliasFromName map[string]moqtmessage.TrackAlias

// 	/*
// 	 * The first keys are Track Namespaces
// 	 * The socond keys are Track Name
// 	 */
// 	tracks map[string]struct {
// 		track map[string]struct {
// 			contentExists   bool
// 			largestGroupID  moqtmessage.GroupID
// 			largestObjectID moqtmessage.ObjectID
// 		}
// 	}
// }

// func (manager *trackManager) register(trackName string) error {
// 	manager.mu.Lock()
// 	defer manager.mu.Unlock()

// 	// Check if the Track Name is already in use
// 	_, ok := manager.aliasFromName[trackName]
// 	if ok {
// 		return ErrDuplicatedTrackAlias
// 	}

// 	// Get new Track Alias
// 	newTrackAlias := moqtmessage.TrackAlias(len(manager.aliasFromName))

// 	manager.aliasFromName[trackName] = newTrackAlias

// 	manager.nameFromAlias[newTrackAlias] = trackName

// 	return nil
// }

// func (manager *trackManager) getTrackName(alias moqtmessage.TrackAlias) (string, error) {
// 	manager.mu.RLock()
// 	defer manager.mu.RUnlock()

// 	name, ok := manager.nameFromAlias[alias]
// 	if !ok {
// 		return "", errors.New("track name not found")
// 	}

// 	return name, nil
// }

// func (manager *trackManager) getTrackAlias(name string) (moqtmessage.TrackAlias, error) {
// 	manager.mu.RLock()
// 	defer manager.mu.RUnlock()

// 	alias, ok := manager.aliasFromName[name]
// 	if !ok {
// 		return 0, errors.New("track alias not found")
// 	}

// 	return alias, nil
// }

// func (manager *trackManager) delete(alias moqtmessage.TrackAlias) error {
// 	manager.mu.Lock()
// 	defer manager.mu.Unlock()

// 	name := manager.nameFromAlias[alias]

// 	delete(manager.nameFromAlias, alias)

// 	delete(manager.aliasFromName, name)

// 	return nil
// }

// func (manager *trackManager) newTrackAlias() moqtmessage.TrackAlias {
// 	manager.mu.RLock()
// 	defer manager.mu.RUnlock()

// 	return moqtmessage.TrackAlias(len(manager.aliasFromName))
// }

type AnnounceConfig struct {
	AuthorizationInfo []string

	MaxCacheDuration time.Duration
}

type SubscribeConfig struct {
	moqtmessage.SubscriberPriority
	moqtmessage.GroupOrder

	moqtmessage.SubscriptionFilter

	AuthorizationInfo string
	DeliveryTimeout   time.Duration
}

/*
 * Transport Session: A raw QUIC connection or a WebTransport session
 */
type TransportSession interface {
	AcceptStream(ctx context.Context) (Stream, error)
	AcceptUniStream(ctx context.Context) (ReceiveStream, error)
	CloseWithError(code SessionErrorCode, msg string) error
	ConnectionState() quic.ConnectionState
	Context() context.Context
	LocalAddr() net.Addr
	OpenStream() (Stream, error)
	OpenStreamSync(ctx context.Context) (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamSync(ctx context.Context) (str SendStream, err error)
	ReceiveDatagram(ctx context.Context) ([]byte, error)
	RemoteAddr() net.Addr
	SendDatagram(b []byte) error
}

type rawQuicConnectionWrapper struct {
	innerSession quic.Connection
}

func (sess *rawQuicConnectionWrapper) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.AcceptStream(ctx)
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	stream, err := sess.innerSession.AcceptUniStream(ctx)
	return &rawQuicReceiveStreamWrapper{innerReceiveStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) CloseWithError(code SessionErrorCode, msg string) error {
	return sess.innerSession.CloseWithError(quic.ApplicationErrorCode(code), msg)
}
func (sess *rawQuicConnectionWrapper) ConnectionState() quic.ConnectionState {
	return sess.innerSession.ConnectionState()
}
func (sess *rawQuicConnectionWrapper) Context() context.Context {
	return sess.innerSession.Context()
}
func (sess *rawQuicConnectionWrapper) LocalAddr() net.Addr {
	return sess.innerSession.LocalAddr()
}
func (sess *rawQuicConnectionWrapper) OpenStream() (Stream, error) {
	stream, err := sess.innerSession.OpenStream()
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.OpenStreamSync(ctx)
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenUniStream() (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStream()
	return &rawQuicSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStreamSync(ctx)
	return &rawQuicSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return sess.innerSession.ReceiveDatagram(ctx)
}
func (sess *rawQuicConnectionWrapper) RemoteAddr() net.Addr {
	return sess.innerSession.RemoteAddr()
}
func (sess *rawQuicConnectionWrapper) SendDatagram(b []byte) error {
	return sess.innerSession.SendDatagram(b)
}

type webtransportSessionWrapper struct {
	innerSession *webtransport.Session
}

func (sess *webtransportSessionWrapper) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.AcceptStream(ctx)
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	stream, err := sess.innerSession.AcceptUniStream(ctx)
	return &webtransportReceiveStreamWrapper{innerReceiveStream: stream}, err
}
func (sess *webtransportSessionWrapper) CloseWithError(code SessionErrorCode, msg string) error {
	return sess.innerSession.CloseWithError(webtransport.SessionErrorCode(code), msg)
}
func (sess *webtransportSessionWrapper) ConnectionState() quic.ConnectionState {
	return sess.innerSession.ConnectionState()
}
func (sess *webtransportSessionWrapper) Context() context.Context {
	return sess.innerSession.Context()
}
func (sess *webtransportSessionWrapper) LocalAddr() net.Addr {
	return sess.innerSession.LocalAddr()
}
func (sess *webtransportSessionWrapper) OpenStream() (Stream, error) {
	stream, err := sess.innerSession.OpenStream()
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.OpenStreamSync(ctx)
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenUniStream() (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStream()
	return &webtransportSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStreamSync(ctx)
	return &webtransportSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *webtransportSessionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return sess.innerSession.ReceiveDatagram(ctx)
}
func (sess *webtransportSessionWrapper) RemoteAddr() net.Addr {
	return sess.innerSession.RemoteAddr()
}
func (sess *webtransportSessionWrapper) SendDatagram(b []byte) error {
	return sess.innerSession.SendDatagram(b)
}

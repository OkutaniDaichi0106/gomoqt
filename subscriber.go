package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
	"github.com/quic-go/quic-go/quicvarint"
)

type subscriber interface {
	Interest(Interest) error

	Subscribe(Subscription) (Info, error)
	Unsubscribe(Subscription)
	UpdateSubscription(Subscription, SubscribeUpdate) (Info, error)

	Fetch(FetchRequest) (Group, moq.ReceiveStream, error)

	RequestInfo(InfoRequest) (Info, error)

	AcceptDataStream(context.Context) (DataReceiver, error)
}

var _ subscriber = (*Subscriber)(nil)

type Subscriber struct {
	isInit bool

	sess              *session
	dataReceiverQueue dataReceiverQueue
	subscriberManager *subscriberManager
}

func (s Subscriber) init() {
	if s.isInit {
		return
	}

	s.isInit = true
}

func (s Subscriber) AcceptDataStream(ctx context.Context) (DataReceiver, error) {
	stream, err := acceptGroupStream(s.sess.conn, ctx)
	if err != nil {
		slog.Error("failed to accept an unidirectional stream")
	}

	group, err := readGroup(stream)

	return dataReceiver{
		Group:  group,
		stream: stream,
	}, nil
}

func (s Subscriber) Interest(interest Interest) error {
	slog.Debug("indicating interest", slog.Any("interest", interest))
	/*
	 * Open an Announce Stream
	 */
	stream, err := openAnnounceStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return err
	}

	aim := message.AnnounceInterestMessage{
		TrackPathPrefix: interest.TrackPrefix,
		Parameters:      message.Parameters(interest.Parameters),
	}

	err = aim.Encode(stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully indicated interest", slog.Any("interest", interest))

	for {

	}

}

func (s Subscriber) Subscribe(subscription Subscription) (info Info, err error) {
	slog.Debug("making a subscription", slog.Any("subscription", subscription))

	// Initialize
	if s.subscriberManager.subscribeSendStreams == nil {
		s.subscriberManager = newSubscriberManager()
	}

	// Open a Subscribe Stream
	stream, err := openSubscribeStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return
	}

	// Set the next Subscribe ID to the Subscription
	subscription.subscribeID = s.subscriberManager.nextSubscribeID()

	/*
	 * Send a SUBSCRIBE message
	 */
	// Set parameters
	if subscription.Parameters == nil {
		subscription.Parameters = make(Parameters)
	}
	if subscription.DeliveryTimeout > 0 {
		subscription.Parameters.Add(DELIVERY_TIMEOUT, subscription.DeliveryTimeout)
	}
	// Initialize a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:        message.SubscribeID(subscription.subscribeID),
		TrackPath:          subscription.TrackPath,
		SubscriberPriority: message.Priority(subscription.SubscriberPriority),
		GroupOrder:         message.GroupOrder(subscription.GroupOrder),
		MinGroupSequence:   message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence:   message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:         message.Parameters(subscription.Parameters),
	}
	err = sm.Encode(stream)
	if err != nil {
		slog.Error("failed to encode a SUBSCRIBE message", slog.String("error", err.Error()), slog.Any("message", sm))
		return
	}

	/*
	 * Receive an INFO message
	 */
	info, err = readInfo(stream)
	if err != nil {
		slog.Error("failed to get a Info", slog.String("error", err.Error()))
		return
	}

	slog.Info("Successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))

	// Register the Subscribe Writer
	err = s.subscriberManager.addSubscription(subscription, stream)
	if err != nil {
		slog.Error("failed to add subscribe sender", slog.String("error", err.Error()))
		return
	}

	return info, nil
}

func (s Subscriber) UpdateSubscription(subscription Subscription, update SubscribeUpdate) (info Info, err error) {
	//
	slog.Debug("updating a subscription",
		slog.Any("subscription", subscription),
		slog.Any("to", update),
	)

	// Verify if the new group range is valid
	if update.MinGroupSequence > update.MaxGroupSequence {
		slog.Debug("MinGroupSequence is larger than MaxGroupSequence")
		return info, ErrInvalidRange
	}
	// Verify if the minimum group sequence become larger
	if subscription.MinGroupSequence > update.MinGroupSequence {
		slog.Debug("the new MinGroupSequence is smaller than the old MinGroupSequence")
		return info, ErrInvalidRange
	}
	// Verify if the maximum group sequence become smaller
	if subscription.MaxGroupSequence < update.MaxGroupSequence {
		slog.Debug("the new MaxGroupSequence is larger than the old MaxGroupSequence")
		return info, ErrInvalidRange
	}

	// Find a sent subscription
	sentSubscription, ok := s.subscriberManager.findSentSubscription(subscription.subscribeID)
	if !ok {
		return info, ErrTrackDoesNotExist
	}

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	// Set parameters
	if update.Parameters == nil {
		update.Parameters = make(Parameters)
	}
	if update.DeliveryTimeout > 0 {
		update.Parameters.Add(DELIVERY_TIMEOUT, update.DeliveryTimeout)
	}
	// Initialize
	sum := message.SubscribeUpdateMessage{
		SubscribeID:        message.SubscribeID(subscription.subscribeID),
		SubscriberPriority: message.Priority(update.SubscriberPriority),
		GroupOrder:         message.GroupOrder(update.GroupOrder),
		GroupExpires:       update.GroupExpires,
		MinGroupSequence:   message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence:   message.GroupSequence(update.MaxGroupSequence),
		Parameters:         message.Parameters(update.Parameters),
	}

	err = sum.Encode(sentSubscription.stream)
	if err != nil {
		slog.Debug("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return
	}

	info, err = readInfo(sentSubscription.stream)
	if err != nil {
		slog.Debug("failed to get an Info")
		return
	}

	// Update the subscription
	if update.SubscriberPriority != 0 {
		subscription.SubscriberPriority = update.SubscriberPriority
	}
	if update.GroupExpires != 0 {
		subscription.GroupExpires = update.GroupExpires
	}
	if update.GroupOrder != 0 {
		subscription.GroupOrder = update.GroupOrder
	}
	subscription.MinGroupSequence = update.MinGroupSequence
	subscription.MaxGroupSequence = update.MaxGroupSequence
	subscription.Parameters = update.Parameters
	if update.DeliveryTimeout != 0 {
		subscription.DeliveryTimeout = update.DeliveryTimeout
	}

	return info, nil
}

func (s Subscriber) Unsubscribe(subscription Subscription) {
	// Find sent subscription
	sentSubscription, ok := s.subscriberManager.findSentSubscription(subscription.subscribeID)
	if !ok {
		return
	}

	// Close gracefully
	err := sentSubscription.stream.Close()
	if err != nil {
		slog.Error("failed to close a subscribe stream", slog.String("error", err.Error()))
	}

	// var code moq.StreamErrorCode

	// var strerr moq.StreamError
	// if errors.As(err, &strerr) {
	// 	code = strerr.StreamErrorCode()
	// } else {
	// 	suberr, ok := err.(SubscribeError)
	// 	if ok {
	// 		code = moq.StreamErrorCode(suberr.SubscribeErrorCode())
	// 	} else {
	// 		code = ErrInternalError.StreamErrorCode()
	// 	}
	// }

	// // Send the error
	// sentSubscription.stream.CancelRead(code)
	// sentSubscription.stream.CancelWrite(code)

	// Remove
	s.subscriberManager.removeSubscriberSender(sentSubscription.subscribeID)

	slog.Info("unsubscribed")
}

func (s Subscriber) Fetch(req FetchRequest) (group Group, rcvstream moq.ReceiveStream, err error) {
	/*
	 * Open a Fetch Stream
	 */
	stream, err := openFetchStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open a Fetch Stream", slog.String("error", err.Error()))
		return
	}

	/*
	 * Send a FETCH message
	 */
	fm := message.FetchMessage{
		TrackPath:          req.TrackPath,
		SubscriberPriority: message.Priority(req.SubscriberPriority),
		GroupSequence:      message.GroupSequence(req.GroupSequence),
		FrameSequence:      message.FrameSequence(req.FrameSequence),
	}

	err = fm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return
	}

	/*
	 * Receive a GROUP message
	 */
	group, err = readGroup(stream)
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return
	}

	rcvstream = stream

	return
}

func (s Subscriber) RequestInfo(req InfoRequest) (Info, error) {
	slog.Debug("requesting information of a track", slog.Any("info request", req))

	/*
	 * Open an Info Stream
	 */
	stream, err := openInfoStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open an Info Stream", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 * Send an INFO_REQUEST message
	 */
	irm := message.InfoRequestMessage{
		TrackPath: req.TrackPath,
	}
	err = irm.Encode(stream)
	if err != nil {
		slog.Error("failed to send an INFO_REQUEST message", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 * Receive a INFO message
	 */
	var im message.InfoMessage
	err = im.Decode(stream)
	if err != nil {
		slog.Error("failed to get a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 * Close the Info Stream
	 */
	err = stream.Close()
	if err != nil {
		slog.Error("failed to close an Info Stream", slog.String("error", err.Error()))
	}

	info := Info{
		PublisherPriority:   Priority(im.PublisherPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}

	slog.Info("Successfully get track information", slog.Any("info", info))

	return info, nil
}

func openAnnounceStream(conn moq.Connection) (moq.Stream, error) {
	slog.Debug("opening an Announce Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_announce,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openSubscribeStream(conn moq.Connection) (moq.Stream, error) {
	slog.Debug("opening an Subscribe Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_subscribe,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openInfoStream(conn moq.Connection) (moq.Stream, error) {
	slog.Debug("opening an Info Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_info,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openFetchStream(conn moq.Connection) (moq.Stream, error) {
	slog.Debug("opening an Fetch Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_fetch,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

// func acceptDataStream(conn moq.Connection, ctx context.Context) (Group, moq.ReceiveStream, error) {
// 	stream, err := acceptGroupStream(conn, ctx)
// 	if err != nil {
// 		slog.Error("failed to accept a Group Stream", slog.String("error", err.Error()))
// 		return Group{}, nil, err
// 	}

// 	group, err := readGroup(stream)
// 	if err != nil {
// 		slog.Error("failed to get a Group", slog.String("error", err.Error()))
// 		return Group{}, nil, err
// 	}

// 	return group, stream, nil
// }

func acceptGroupStream(conn moq.Connection, ctx context.Context) (moq.ReceiveStream, error) {
	// Accept an unidirectional stream
	stream, err := conn.AcceptUniStream(ctx)
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Receive a STREAM_TYPE message
	var stm message.StreamTypeMessage
	err = stm.Decode(stream)
	if err != nil {
		slog.Error("failed to receive a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func receiveDatagram(conn moq.Connection, ctx context.Context) (Group, []byte, error) {
	data, err := conn.ReceiveDatagram(ctx)
	if err != nil {
		slog.Error("failed to receive a datagram", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	reader := bytes.NewReader(data)

	group, err := readGroup(quicvarint.NewReader(reader))
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	// Read payload in the rest of the data
	buf := make([]byte, reader.Len())
	_, err = reader.Read(buf)

	if err != nil {
		slog.Error("failed to read payload", slog.String("error", err.Error()))
		return group, nil, err
	}

	return group, buf, nil
}

/*
 * Interest Manager
 */
// func newInterestManager() *interestManager {
// 	return &interestManager{
// 		//announceReceivers: make(map[string]*AnnounceReceiver),
// 	}
// }

// type interestManager struct {
// 	/***/
// 	interests map[string]Interest

// 	/*
// 	 * Received announcements
// 	 */
// 	receivedAnnouncements map[string]Announcement
// 	anMu                  sync.RWMutex
// }

// func (im *interestManager) addAnnounceReceiver(ar *AnnounceReceiver) error {
// 	im.anMu.Lock()
// 	defer im.anMu.Unlock()

// 	_, ok := im.receivedAnnouncements[ar.interest.TrackPrefix]
// 	if ok {
// 		return errors.New("duplicated interest")
// 	}

// 	im.receivedAnnouncements[ar.interest.TrackPrefix] = ar

// 	return nil
// }

// func (im *interestManager) removeAnnounceReceiver(trackPrefix string) {
// 	im.anMu.Lock()
// 	defer im.anMu.Unlock()

// 	delete(im.receivedAnnouncements, trackPrefix)
// }

/*
 * Subscribe Manager
 */

func newSubscriberManager() *subscriberManager {
	return &subscriberManager{
		subscribeSendStreams: make(map[SubscribeID]*subscribeSendStream),
	}
}

type subscriberManager struct {
	/*
	 *
	 */
	interestSentStreams map[string]*interestSentStream
	issMu               sync.RWMutex

	/*
	 *
	 */
	subscribeIDCounter uint64

	/*
	 *
	 */
	subscribeSendStreams map[SubscribeID]*subscribeSendStream
	sssMu                sync.RWMutex
}

// func (sm *subscriberManager) getAnnouncements() (announcements []Announcement) {
// 	sm.raMu.RLock()
// 	defer sm.raMu.RUnlock()

// 	announcements = make([]Announcement, len(sm.receivedAnnouncements))

// 	for _, announcement := range announcements {
// 		announcements = append(announcements, announcement)
// 	}

// 	return announcements
// }

func (sm *subscriberManager) nextSubscribeID() SubscribeID {
	// Get a new Subscribe ID
	new := SubscribeID(sm.subscribeIDCounter)
	// Increment
	sm.subscribeIDCounter++

	return new
}

func (sm *subscriberManager) findSentSubscription(id SubscribeID) (*subscribeSendStream, bool) {
	sm.sssMu.RLock()
	defer sm.sssMu.RUnlock()

	sentSubscription, ok := sm.subscribeSendStreams[id]
	if !ok {
		return nil, false
	}

	return sentSubscription, true
}

func (sm *subscriberManager) addSubscription(subscription Subscription, stream moq.Stream) error {
	sm.sssMu.Lock()
	defer sm.sssMu.Unlock()

	_, ok := sm.subscribeSendStreams[subscription.subscribeID]
	if ok {
		return ErrDuplicatedSubscribeID
	}

	sm.subscribeSendStreams[subscription.subscribeID] = &subscribeSendStream{
		Subscription: subscription,
		stream:       stream,
	}

	return nil
}

func (sm *subscriberManager) removeSubscriberSender(id SubscribeID) {
	sm.sssMu.Lock()
	defer sm.sssMu.Unlock()

	delete(sm.subscribeSendStreams, id)
}

func (sm *subscriberManager) addInterestSendStream(iss *interestSentStream) {
	sm.issMu.Lock()
	defer sm.issMu.Unlock()

	old, ok := sm.interestSentStreams[iss.interest.TrackPrefix]
	if ok {
		// TODO:terminate the stream
		sm.removeInterestSendStream(old)
	}

	sm.interestSentStreams[iss.interest.TrackPrefix] = iss
}

func (sm *subscriberManager) removeInterestSendStream(iss *interestSentStream) {
	sm.issMu.Lock()
	defer sm.issMu.Unlock()

	iss.close()

	delete(sm.interestSentStreams, iss.interest.TrackPrefix)
}

type interestSentStream struct {
	interest      Interest
	liveCh        chan struct{}
	announcements map[string]Announcement
	stream        moq.Stream
	mu            sync.RWMutex
	closeCh       chan struct{}
}

func (iss *interestSentStream) getAnnouncements() []Announcement {
	<-iss.liveCh

	iss.mu.RLock()
	defer iss.mu.RUnlock()

	announcements := make([]Announcement, 0, len(iss.announcements))

	for _, announcement := range iss.announcements {
		announcements = append(announcements, announcement)
	}

	return announcements
}

func (iss *interestSentStream) close() {
	iss.closeCh <- struct{}{}

	iss.mu.Lock()
	defer iss.mu.Unlock()

	err := iss.stream.Close()
	if err != nil {
		slog.Error("failed to close an interest send stream", slog.String("error", err.Error()))
		return
	}
}

func (iss *interestSentStream) closeWithError(err error) {
	if err == nil {
		iss.close()
		return
	}

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		annerr, ok := err.(AnnounceError)
		if ok {
			code = moq.StreamErrorCode(annerr.AnnounceErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	iss.stream.CancelRead(code)
	iss.stream.CancelWrite(code)

	if err != nil {
		slog.Error("failed to close an interest send stream", slog.String("error", err.Error()))
		return
	}

	iss.close()
}

func (iss *interestSentStream) listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			// TODO:
			return ctx.Err()
		case <-iss.closeCh:
			// TOOD:
			return nil
		default:
			announcemnt, err := readAnnouncement(iss.stream)
			if err != nil {
				slog.Error("failed to read an announcement", slog.String("error", err.Error()))
				//
				iss.closeWithError(err)
				return err
			}

			// handle the announcement
			err = func() error {
				iss.mu.Lock()
				defer iss.mu.Unlock()

				_, ok := iss.announcements[announcemnt.TrackPath]
				switch announcemnt.status {
				case ACTIVE:
					if !ok {
						// Add
						iss.announcements[announcemnt.TrackPath] = announcemnt
					} else {
						return errors.New("invalid active announcement")
					}
				case ENDED:
					if ok {
						// Remove
						delete(iss.announcements, announcemnt.TrackPath)
					} else {
						return errors.New("invalid ended announcement")
					}
				case LIVE:
					iss.liveCh <- struct{}{}
				default:
					return ErrProtocolViolation
				}

				return nil
			}()
			if err != nil {
				return err
			}
		}
	}
}

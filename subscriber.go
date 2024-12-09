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

type Subscriber interface {
	Interest(Interest) (AnnounceReceiver, error)
	Subscribe(Subscription) (*SubscribeSender, Info, error)
	Fetch(FetchRequest) (Group, moq.ReceiveStream, error)
	RequestInfo(InfoRequest) (Info, error)
}

var _ Subscriber = (*subscriber)(nil)

type subscriber struct {
	sess             *session
	interestManager  *interestManager
	subscribeManager *subscribeManager
}

func (s subscriber) Interest(interest Interest) (AnnounceReceiver, error) {
	slog.Debug("indicating interest", slog.Any("interest", interest))
	/*
	 * Open an Announce Stream
	 */
	stream, err := openAnnounceStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return AnnounceReceiver{}, err
	}

	aim := message.AnnounceInterestMessage{
		TrackPathPrefix: interest.TrackPrefix,
		Parameters:      message.Parameters(interest.Parameters),
	}

	err = aim.Encode(stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return AnnounceReceiver{}, err
	}

	slog.Info("Successfully indicated interest", slog.Any("interest", interest))

	return AnnounceReceiver{
		stream: stream,
	}, nil
}

func (s subscriber) Subscribe(subscription Subscription) (*SubscribeSender, Info, error) {
	slog.Debug("making a subscription", slog.Any("subscription", subscription))

	// Open a Subscribe Stream
	stream, err := openSubscribeStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return nil, Info{}, err
	}

	//
	if s.subscribeManager.subscribeSenders == nil {
		s.subscribeManager = newSubscriberManager()
	}

	// Set the next Subscribe ID to the Subscription
	subscription.subscribeID = s.subscribeManager.nextSubscribeID()

	/*
	 * Send a SUBSCRIBE message
	 */
	// Initialize a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:        message.SubscribeID(subscription.subscribeID),
		TrackPath:          subscription.TrackPath,
		SubscriberPriority: message.SubscriberPriority(subscription.SubscriberPriority),
		GroupOrder:         message.GroupOrder(subscription.GroupOrder),
		MinGroupSequence:   message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence:   message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:         message.Parameters(subscription.Parameters),
	}
	err = sm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()), slog.Any("message", sm))
		return nil, Info{}, err
	}

	/*
	 * Receive an INFO message and get an Info
	 */
	info, err := readInfo(stream)
	if err != nil {
		slog.Error("failed to get a Info", slog.String("error", err.Error()))
		return nil, Info{}, err
	}

	slog.Info("Successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))

	sr := SubscribeSender{
		stream:       stream,
		subscription: subscription,
	}

	// Register the Subscribe Writer
	err = s.subscribeManager.addSubscribeSender(&sr)
	if err != nil {
		slog.Error("failed to add subscribe sender", slog.String("error", err.Error()))
		return nil, Info{}, err
	}

	return &sr, info, nil
}

func (s subscriber) Fetch(req FetchRequest) (group Group, rcvstream moq.ReceiveStream, err error) {
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
		SubscriberPriority: message.SubscriberPriority(req.SubscriberPriority),
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

func (s subscriber) RequestInfo(req InfoRequest) (Info, error) {
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
		PublisherPriority:   PublisherPriority(im.PublisherPriority),
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

func acceptDataStream(conn moq.Connection, ctx context.Context) (Group, moq.ReceiveStream, error) {
	stream, err := acceptGroupStream(conn, ctx)
	if err != nil {
		slog.Error("failed to accept a Group Stream", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	group, err := readGroup(stream)
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	return group, stream, nil
}

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
func newInterestManager() *interestManager {
	return &interestManager{
		announceReceivers: make(map[string]*AnnounceReceiver),
	}
}

type interestManager struct {

	/*
	 * Received
	 */
	announceReceivers map[string]*AnnounceReceiver
	anMu              sync.RWMutex
}

func (im *interestManager) addAnnounceReceiver(ar *AnnounceReceiver) error {
	im.anMu.Lock()
	defer im.anMu.Unlock()

	_, ok := im.announceReceivers[ar.interest.TrackPrefix]
	if ok {
		return errors.New("duplicated interest")
	}

	im.announceReceivers[ar.interest.TrackPrefix] = ar

	return nil
}

func (im *interestManager) removeAnnounceReceiver(trackPrefix string) {
	im.anMu.Lock()
	defer im.anMu.Unlock()

	delete(im.announceReceivers, trackPrefix)

}

/*
 * Subscribe Manager
 */

func newSubscriberManager() *subscribeManager {
	return &subscribeManager{
		subscribeSenders: make(map[SubscribeID]*SubscribeSender),
	}
}

type subscribeManager struct {
	/*
	 *
	 */
	subscribeIDCounter uint64

	/*
	 *
	 */
	subscribeSenders map[SubscribeID]*SubscribeSender
	ssMu             sync.RWMutex
}

func (sm *subscribeManager) nextSubscribeID() SubscribeID {
	sm.subscribeIDCounter++
	return SubscribeID(sm.subscribeIDCounter)
}

func (sm *subscribeManager) addSubscribeSender(ss *SubscribeSender) error {
	sm.ssMu.Lock()
	defer sm.ssMu.Unlock()

	_, ok := sm.subscribeSenders[ss.subscription.subscribeID]
	if ok {
		return ErrDuplicatedSubscribeID
	}

	sm.subscribeSenders[ss.subscription.subscribeID] = ss

	return nil
}

func (sm *subscribeManager) removeSubscriberSender(id SubscribeID) {
	sm.ssMu.Lock()
	defer sm.ssMu.Unlock()

	delete(sm.subscribeSenders, id)
}

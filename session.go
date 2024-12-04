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

type session struct {
	conn   moq.Connection
	stream SessionStream
	//version Version

	/*
	 * Sent Subscriptions
	 */
	subscribeWriters map[SubscribeID]*SubscribeWriter
	ssMu             sync.RWMutex

	/*
	 *
	 */
	receivedSubscriptions map[string]Subscription
	rsMu                  sync.RWMutex
}

func (sess *session) SessionInit(conn moq.Connection) error {
	sess = &session{
		conn: conn,
	}

	/*
	 * Open a Session Stream
	 */
	stream, err := sess.openSessionStream()
	if err != nil {
		slog.Error("failed to open a Session Stream")
		return err
	}

	// Set the stream
	sess.stream = stream

	//
	sess.subscribeWriters = make(map[SubscribeID]*SubscribeWriter)

	//
	sess.receivedSubscriptions = make(map[string]Subscription)

	return nil
}

func (sess *session) Terminate(err error) {
	slog.Info("terminating a Session", slog.String("reason", err.Error()))

	var tererr TerminateError

	if err == nil {
		tererr = NoErrTerminate
	} else {
		var ok bool
		tererr, ok = err.(TerminateError)
		if !ok {
			tererr = ErrInternalError
		}
	}

	err = sess.conn.CloseWithError(moq.SessionErrorCode(tererr.TerminateErrorCode()), err.Error())
	if err != nil {
		slog.Error("failed to close the Connection", slog.String("error", err.Error()))
		return
	}

	slog.Info("terminated a Session")
}

func (sess *session) Interest(interest Interest) (AnnounceStream, error) {
	/*
	 * Open an Announce Stream
	 */
	stream, err := sess.openAnnounceStream()
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return AnnounceStream{}, err
	}

	aim := message.AnnounceInterestMessage{
		TrackPrefix: interest.TrackPrefix,
		Parameters:  message.Parameters(interest.Parameters),
	}

	err = aim.Encode(stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return AnnounceStream{}, err
	}

	slog.Info("Interested", slog.Any("track prefix", interest.TrackPrefix))

	return AnnounceStream{
		stream: stream,
	}, nil
}

func (sess *session) Subscribe(subscription Subscription) (*SubscribeWriter, Info, error) {
	slog.Debug("Subscribing", slog.Any("subscription", subscription))

	sess.ssMu.Lock()
	defer sess.ssMu.Unlock()

	// Open a Subscribe Stream
	stream, err := sess.openSubscribeStream()
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return nil, Info{}, err
	}

	if sess.subscribeWriters == nil {
		sess.subscribeWriters = make(map[SubscribeID]*SubscribeWriter)
	}

	// Set the next Subscribe ID to the Subscription
	subscription.subscribeID = SubscribeID(len(sess.subscribeWriters))

	sm := message.SubscribeMessage{
		SubscribeID:        message.SubscribeID(subscription.subscribeID),
		TrackNamespace:     subscription.TrackNamespace,
		TrackName:          subscription.TrackName,
		SubscriberPriority: message.SubscriberPriority(subscription.SubscriberPriority),
		GroupOrder:         message.GroupOrder(subscription.GroupOrder),
		MinGroupSequence:   message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence:   message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:         message.Parameters(subscription.Parameters),
	}

	/*
	 * Send a SUBSCRIBE message
	 */
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

	sw := SubscribeWriter{
		stream:       stream,
		subscription: subscription,
	}

	// Register the Subscribe Writer
	sess.subscribeWriters[subscription.subscribeID] = &sw

	return &sw, info, nil
}

func (sess *session) Fetch(req FetchRequest) (FetchStream, error) {
	/*
	 * Open a Fetch Stream
	 */
	stream, err := sess.openFetchStream()
	if err != nil {
		slog.Error("failed to open a Fetch Stream", slog.String("error", err.Error()))
		return FetchStream{}, err
	}

	/*
	 * Send a FETCH message
	 */
	fm := message.FetchMessage{
		TrackNamespace:     req.TrackName,
		TrackName:          req.TrackName,
		SubscriberPriority: message.SubscriberPriority(req.SubscriberPriority),
		GroupSequence:      message.GroupSequence(req.GroupSequence),
		FrameSequence:      message.FrameSequence(req.FrameSequence),
	}

	err = fm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return FetchStream{}, err
	}

	/*
	 * Receive a GROUP message
	 */
	group, err := readGroup(stream)
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return FetchStream{}, err
	}

	return FetchStream{
		stream: stream,
		group:  group,
	}, nil
}

func (sess *session) RequestInfo(req InfoRequest) (Info, error) {
	/*
	 * Open an Info Stream
	 */
	stream, err := sess.openInfoStream()
	if err != nil {
		slog.Error("failed to open an Info Stream", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 * Send an INFO_REQUEST message
	 */
	irm := message.InfoRequestMessage{
		TrackNamespace: req.TrackNamespace,
		TrackName:      req.TrackName,
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

	return Info{
		PublisherPriority:   PublisherPriority(im.PublisherPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}, nil
}

func (sess *session) openSessionStream() (moq.Stream, error) {
	stream, err := sess.conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_session,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (sess *session) openAnnounceStream() (moq.Stream, error) {
	stream, err := sess.conn.OpenStream()
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

func (sess *session) openSubscribeStream() (moq.Stream, error) {
	stream, err := sess.conn.OpenStream()
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

func (sess *session) openInfoStream() (moq.Stream, error) {
	stream, err := sess.conn.OpenStream()
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

func (sess *session) openFetchStream() (moq.Stream, error) {
	stream, err := sess.conn.OpenStream()
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

func (sess *session) openGroupStream() (moq.SendStream, error) {
	stream, err := sess.conn.OpenUniStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_group,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (sess *session) acceptGroupStream(ctx context.Context) (moq.ReceiveStream, error) {
	// Accept an unidirectional stream
	stream, err := sess.conn.AcceptUniStream(ctx)
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

func (sess *session) openDataStream(g Group) (moq.SendStream, error) {
	if g.groupSequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	stream, err := sess.openGroupStream()
	if err != nil {
		slog.Error("failed to open an unidirectional Stream", slog.String("error", err.Error()))
		return nil, err
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.subscribeID),
		GroupSequence:     message.GroupSequence(g.groupSequence),
		PublisherPriority: message.PublisherPriority(g.PublisherPriority),
	}

	// Send the GROUP message
	err = gm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (sess *session) acceptDataStream(ctx context.Context) (Group, moq.ReceiveStream, error) {
	stream, err := sess.acceptGroupStream(ctx)
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

func (sess *session) sendDatagram(g Group, payload []byte) error {
	if g.groupSequence == 0 {
		return errors.New("0 sequence number")
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.subscribeID),
		GroupSequence:     message.GroupSequence(g.groupSequence),
		PublisherPriority: message.PublisherPriority(g.PublisherPriority),
	}

	var buf bytes.Buffer

	// Encode the GROUP message
	err := gm.Encode(&buf)
	if err != nil {
		slog.Error("failed to encode a GROUP message", slog.String("error", err.Error()))
		return err
	}

	// Encode the payload
	_, err = buf.Write(payload)
	if err != nil {
		slog.Error("failed to encode a payload", slog.String("error", err.Error()))
		return err
	}

	// Send the data with the GROUP message
	err = sess.conn.SendDatagram(buf.Bytes())
	if err != nil {
		slog.Error("failed to send a datagram", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (sess *session) receiveDatagram(ctx context.Context) (Group, []byte, error) {
	data, err := sess.conn.ReceiveDatagram(ctx)
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

func (sess *session) acceptSubscription(subscription Subscription) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	// Get Full Track Name
	fullName := subscription.TrackNamespace + "/" + subscription.TrackName

	// Verify if the subscription is duplicated or not
	_, ok := sess.receivedSubscriptions[fullName]
	if ok {
		slog.Debug("duplicated subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	// Register the subscription
	sess.receivedSubscriptions[fullName] = subscription

	slog.Info("Accepted a new subscription", slog.Any("subscription", subscription))
}

func (sess *session) updateSubscription(subscription Subscription) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	// Get Full Track Name
	fullName := subscription.TrackNamespace + "/" + subscription.TrackName

	old, ok := sess.receivedSubscriptions[fullName]
	if !ok {
		slog.Debug("no subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	sess.receivedSubscriptions[fullName] = subscription

	slog.Info("updated a subscription", slog.Any("from", old), slog.Any("to", subscription))
}

func (sess *session) removeSubscription(subscription Subscription) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	// Get Full Track Name
	fullName := subscription.TrackNamespace + "/" + subscription.TrackName

	if subscription, ok := sess.receivedSubscriptions[fullName]; !ok {
		slog.Debug("no subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	delete(sess.receivedSubscriptions, fullName)
}

/*
 * Server Session Handler
 */
type ServerSessionHandler interface {
	HandleServerSession(*ServerSession)
}

type ServerSessionHandlerFunc func(*ServerSession)

func (f ServerSessionHandlerFunc) HandleServerSession(sess *ServerSession) {
	f(sess)
}

/*
 * Client Session Handler
 */
type ClientSessionHandler interface {
	HandleClientSession(*ClientSession)
}

type ClientSessionHandlerFunc func(*ClientSession)

func (f ClientSessionHandlerFunc) HandleClientSession(sess *ClientSession) {
	f(sess)
}

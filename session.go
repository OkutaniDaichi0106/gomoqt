package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
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

	receivedSubscriptions map[SubscribeID]Subscription
	rsMu                  sync.RWMutex

	doneCh chan struct{}
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
	stream, err := sess.openControlStream(stream_type_announce)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return AnnounceStream{}, err
	}

	tp := strings.Split(interest.TrackPrefix, "/")

	aim := message.AnnounceInterestMessage{
		TrackPrefix: tp,
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
	stream, err := sess.openControlStream(stream_type_subscribe)
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
		TrackNamespace:     strings.Split(subscription.TrackNamespace, "/"),
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
	stream, err := sess.openControlStream(stream_type_fetch)
	if err != nil {
		slog.Error("failed to open a Fetch Stream", slog.String("error", err.Error()))
		return FetchStream{}, err
	}

	err = message.FetchMessage(req).Encode(stream)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return FetchStream{}, err
	}

	return FetchStream{
		stream: stream,
	}, nil
}

func (sess *session) RequestInfo(req InfoRequest) (Info, error) {
	stream, err := sess.conn.OpenStream()
	if err != nil {
		slog.Error("failed to open an Info Request Stream", slog.String("error", err.Error()))
		return Info{}, err
	}

	// Send Announce Stream Type
	_, err = stream.Write([]byte{byte(stream_type_info)})
	if err != nil {
		slog.Error("failed to send Announce Stream Type")
		return Info{}, err
	}

	irm := message.InfoRequestMessage{
		TrackNamespace: strings.Split(req.TrackNamespace, "/"),
		TrackName:      req.TrackName,
	}

	err = irm.Encode(stream)

	if err != nil {
		slog.Error("failed to send a INFO_REQUEST message", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 *
	 */
	r, err := message.NewReader(stream)
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return Info{}, err
	}

	var info message.InfoMessage
	err = info.Decode(r)
	if err != nil {
		slog.Error("failed to get a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	return Info(info), nil
}

func (sess *session) openControlStream(st StreamType) (moq.Stream, error) {
	stream, err := sess.conn.OpenStream()
	if err != nil {
		slog.Error("failed to open an Info Request Stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Send Announce Stream Type
	_, err = stream.Write([]byte{byte(st)})
	if err != nil {
		slog.Error("failed to send Announce Stream Type")
		return nil, err
	}

	return stream, nil
}

func (sess *session) openDataStream(g Group) (moq.SendStream, error) {
	if g.groupSequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	stream, err := sess.conn.OpenUniStream()
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
	stream, err := sess.conn.AcceptUniStream(ctx)
	if err != nil {
		slog.Error("failed to accept an unidirectional stream", slog.String("error", err.Error()))
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

	_, ok := sess.receivedSubscriptions[subscription.subscribeID]
	if ok {
		slog.Debug("duplicated subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	sess.receivedSubscriptions[subscription.subscribeID] = subscription
}

func (sess *session) updateSubscription(subscription Subscription) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	old, ok := sess.receivedSubscriptions[subscription.subscribeID]
	if !ok {
		slog.Debug("no subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	sess.receivedSubscriptions[subscription.subscribeID] = subscription

	slog.Info("updated a subscription", slog.Any("from", old), slog.Any("to", subscription))
}

func (sess *session) stopSubscription(id SubscribeID) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	if subscription, ok := sess.receivedSubscriptions[id]; !ok {
		slog.Debug("no subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	delete(sess.receivedSubscriptions, id)
}

type ServerSessionHandler interface {
	HandleServerSession(*ServerSession)
}

type ClientSessionHandler interface {
	HandleClientSession(*ClientSession)
}

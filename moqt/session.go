package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type session struct {
	conn    Connection
	sessStr SessionStream
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

	err = sess.conn.CloseWithError(SessionErrorCode(tererr.TerminateErrorCode()), err.Error())
	if err != nil {
		slog.Error("failed to close the Connection", slog.String("error", err.Error()))
		return
	}

	slog.Info("terminated a Session")
}

func (sess *session) Interest(interest Interest) (AnnounceStream, error) {
	//
	stream, err := sess.conn.OpenStream()
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return AnnounceStream{}, err
	}

	// Send Announce Stream Type
	_, err = stream.Write([]byte{byte(ANNOUNCE)})
	if err != nil {
		slog.Error("failed to send Announce Stream Type")
		return AnnounceStream{}, err
	}

	tp := strings.Split(interest.TrackPrefix, "/")

	aim := message.AnnounceInterestMessage{
		TrackPrefix: tp,
		Parameters:  message.Parameters(interest.Parameters),
	}

	_, err = stream.Write(aim.SerializePayload())
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return AnnounceStream{}, err
	}

	slog.Info("Interested", slog.Any("track prefix", interest.TrackPrefix))

	return AnnounceStream{
		reader: quicvarint.NewReader(stream),
		stream: stream,
	}, nil
}

func (sess *session) Subscribe(subscription Subscription) (*SubscribeWriter, Info, error) {
	sess.ssMu.Lock()
	defer sess.ssMu.Unlock()

	// Open a Subscribe Stream
	stream, err := sess.conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return nil, Info{}, err
	}

	// Send Announce Stream Type
	_, err = stream.Write([]byte{byte(SUBSCRIBE)})
	if err != nil {
		slog.Error("failed to send Announce Stream Type")
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
	_, err = stream.Write(sm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()), slog.Any("message", sm))
		return nil, Info{}, err
	}

	slog.Info("successfully subscribed", slog.Any("subscription", subscription))

	/*
	 * Receive an INFO message and get an Info
	 */
	info, err := getInfo(quicvarint.NewReader(stream))
	if err != nil {
		slog.Error("failed to get a Info", slog.String("error", err.Error()))
		return nil, Info{}, err
	}

	sw := SubscribeWriter{
		reader:       quicvarint.NewReader(stream),
		stream:       stream,
		subscription: subscription,
	}

	sess.subscribeWriters[subscription.subscribeID] = &sw

	return &sw, info, nil
}

func (sess *session) Fetch(req FetchRequest) (FetchStream, error) {
	stream, err := sess.conn.OpenStream()
	if err != nil {
		slog.Error("failed to open an Fetch Stream", slog.String("error", err.Error()))
		return FetchStream{}, err
	}

	// Send Announce Stream Type
	_, err = stream.Write([]byte{byte(FETCH)})
	if err != nil {
		slog.Error("failed to send Announce Stream Type")
		return FetchStream{}, err
	}

	_, err = stream.Write(message.FetchMessage(req).SerializePayload())
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
	_, err = stream.Write([]byte{byte(INFO)})
	if err != nil {
		slog.Error("failed to send Announce Stream Type")
		return Info{}, err
	}

	irm := message.InfoRequestMessage{
		TrackNamespace: strings.Split(req.TrackNamespace, "/"),
		TrackName:      req.TrackName,
	}

	_, err = stream.Write(irm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a INFO_REQUEST message", slog.String("error", err.Error()))
		return Info{}, err
	}

	var info message.InfoMessage
	err = info.DeserializePayload(quicvarint.NewReader(stream))
	if err != nil {
		slog.Error("failed to get a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	return Info(info), nil
}

func (sess *session) openDataStream(g Group) (SendStream, error) {
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
	_, err = stream.Write(gm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (sess *session) acceptDataStream(ctx context.Context) (Group, ReceiveStream, error) {
	stream, err := sess.conn.AcceptUniStream(ctx)
	if err != nil {
		slog.Error("failed to accept an unidirectional stream", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	group, err := getGroup(quicvarint.NewReader(stream))
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

	p := gm.SerializePayload()

	p = append(p, payload...)

	// Send the data with the GROUP message
	err := sess.conn.SendDatagram(p)
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

	group, err := getGroup(quicvarint.NewReader(reader))
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

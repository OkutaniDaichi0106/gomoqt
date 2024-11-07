package moqt

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Session struct {
	Connection    Connection
	SessionStream SessionStream
	//version Version

	/*
	 * Sent Subscriptions
	 */
	subscribeWriters map[SubscribeID]*SubscribeWriter
	ssMu             sync.RWMutex

	receivedSubscriptions map[SubscribeID]Subscription
	rsMu                  sync.RWMutex

	terrCh chan TerminateError
}

func (sess *Session) GoAway(url string, timeout time.Duration) error {
	gam := message.GoAwayMessage{
		NewSessionURI: url,
	}

	_, err := sess.SessionStream.Write(gam.SerializePayload())
	if err != nil {
		slog.Error("failed to send a GOAWAY message", slog.String("error", err.Error()))
		return err
	}

	// Lock the Mutex and stop making new subscription
	sess.ssMu.Lock()
	sess.rsMu.Lock()

	time.Sleep(timeout)

	sess.Terminate(ErrGoAwayTimeout)

	return nil
}

func (sess *Session) Terminate(terr TerminateError) {
	err := sess.Connection.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
	if err != nil {
		slog.Error("failed to close the Session", slog.String("error", err.Error()))
		return
	}

	slog.Info("closed the Session", slog.String("reason", terr.Error()))
}

func (sess *Session) Interest(interest Interest) (AnnounceStream, error) {
	//
	stream, err := sess.Connection.OpenStream()
	if err != nil {
		slog.Error("failed to open an Announce Stream")
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

func (sess *Session) Subscribe(subscription Subscription) (*SubscribeWriter, Info, error) {
	sess.ssMu.Lock()
	defer sess.ssMu.Unlock()

	// Open a Subscribe Stream
	stream, err := sess.Connection.OpenStream()
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

func (sess *Session) Fetch(req FetchRequest) (FetchStream, error) {
	stream, err := sess.Connection.OpenStream()
	if err != nil {
		slog.Error("failed to open an Fetch Stream", slog.String("error", err.Error()))
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

func (sess *Session) RequestInfo(req InfoRequest) (Info, error) {
	stream, err := sess.Connection.OpenStream()
	if err != nil {
		slog.Error("failed to open an Info Request Stream", slog.String("error", err.Error()))
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

func (sess *Session) OpenDataStream(g Group) (SendStream, error) {
	stream, err := sess.Connection.OpenUniStream()
	if err != nil {
		slog.Error("failed to open an unidirectional Stream", slog.String("error", err.Error()))
		return nil, err
	}

	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(g.SubscribeID),
		GroupSequence:     message.GroupSequence(g.GroupSequence),
		PublisherPriority: message.PublisherPriority(g.PublisherPriority),
	}

	_, err = stream.Write(gm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func (sess *Session) acceptSubscription(subscription Subscription) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	_, ok := sess.receivedSubscriptions[subscription.subscribeID]
	if ok {
		slog.Debug("duplicated subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	sess.receivedSubscriptions[subscription.subscribeID] = subscription
}

func (sess *Session) updateSubscription(subscription Subscription) {
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

func (sess *Session) stopSubscription(id SubscribeID) {
	sess.rsMu.Lock()
	defer sess.rsMu.Unlock()

	if subscription, ok := sess.receivedSubscriptions[id]; !ok {
		slog.Debug("no subscription", slog.Any("Subscribe ID", subscription.subscribeID))
		return
	}

	delete(sess.receivedSubscriptions, id)
}

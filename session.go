package moqt

import (
	"context"
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type Session interface {
	// Terminate the session
	Terminate(error)

	// Open an Announce Stream
	OpenAnnounceStream(Interest) (*ReceiveAnnounceStream, error)

	// Open a Subscribe Stream
	OpenSubscribeStream(Subscription) (*SendSubscribeStream, error)

	// Open a Fetch Stream
	OpenFetchStream(Fetch) (ReceiveDataStream, error)

	// Open an Info Stream
	OpenInfoStream(InfoRequest) (Info, error)

	// Accept an Announce Stream
	AcceptAnnounceStream(context.Context) (*SendAnnounceStream, error)

	// Accept a Subscribe Stream
	AcceptSubscribeStream(context.Context) (*ReceivedSubscribeStream, error)

	// Accept a Fetch Stream
	AcceptFetchStream(context.Context) (*ReceivedFetch, error)

	// Accept an Info Stream
	AcceptInfoStream(context.Context) (*SendInfoStream, error)
}

var _ Session = &session{}

type session struct {
	conn   transport.Connection
	stream transport.Stream

	//
	receivedSubscriptionQueue *receivedSubscriptionQueue

	receivedInterestQueue *receivedInterestQueue

	receivedFetchQueue *receivedFetchQueue

	receivedInfoRequestQueue *receivedInfoRequestQueue

	dataReceiveStreamQueues map[SubscribeID]*receiveDataStreamQueue

	receivedDatagramQueues map[SubscribeID]*receivedDatagramQueue

	subscribeIDCounter uint64
}

// var _ Subscriber = &session{}
// var _ Publisher = &session{}

func (sess *session) Terminate(err error) {
	slog.Info("Terminating a session", slog.String("reason", err.Error()))

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

	err = sess.conn.CloseWithError(transport.SessionErrorCode(tererr.TerminateErrorCode()), err.Error())
	if err != nil {
		slog.Error("failed to close the Connection", slog.String("error", err.Error()))
		return
	}

	slog.Info("Terminated a session")
}

func (s *session) OpenAnnounceStream(interest Interest) (*ReceiveAnnounceStream, error) {
	slog.Debug("indicating interest", slog.Any("interest", interest))
	/*
	 * Open an Announce Stream
	 */
	stream, err := openAnnounceStream(s.conn)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return nil, err
	}

	aim := message.AnnounceInterestMessage{
		TrackPathPrefix: interest.TrackPrefix,
		Parameters:      message.Parameters(interest.Parameters),
	}

	err = aim.Encode(stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Info("Successfully indicated interest", slog.Any("interest", interest))

	return &ReceiveAnnounceStream{
		Interest: interest,
		active:   makeTracks(1),
		stream:   stream,
	}, nil
}

func (s *session) OpenSubscribeStream(subscription Subscription) (*SendSubscribeStream, error) {
	slog.Debug("making a subscription", slog.Any("subscription", subscription))

	// Open a Subscribe Stream
	stream, err := openSubscribeStream(s.conn)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return nil, err
	}

	nextID := s.nextSubscribeID()
	/*
	 * Send a SUBSCRIBE message
	 */
	// Set parameters
	if subscription.SubscribeParameters == nil {
		subscription.SubscribeParameters = make(Parameters)
	}
	if subscription.Track.DeliveryTimeout > 0 {
		subscription.SubscribeParameters.Add(DELIVERY_TIMEOUT, subscription.Track.DeliveryTimeout)
	}

	// Initialize a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(nextID),
		TrackPath:        subscription.Track.TrackPath,
		TrackPriority:    message.TrackPriority(subscription.Track.TrackPriority),
		GroupOrder:       message.GroupOrder(subscription.Track.GroupOrder),
		GroupExpires:     subscription.Track.GroupExpires,
		MinGroupSequence: message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:       message.Parameters(subscription.SubscribeParameters),
	}
	err = sm.Encode(stream)
	if err != nil {
		slog.Error("failed to encode a SUBSCRIBE message", slog.String("error", err.Error()), slog.Any("message", sm))
		return nil, err
	}

	/*
	 * Receive an INFO message
	 */
	info, err := readInfo(stream)
	if err != nil {
		slog.Error("failed to get a Info", slog.String("error", err.Error()))
		return nil, err
	}

	/*
	 * 	Update the subscription
	 */
	// Update the TrackPriority
	if info.TrackPriority != subscription.Track.TrackPriority {
		slog.Debug("TrackPriority is not updated")
		return nil, ErrPriorityMismatch
	}

	// Update the GroupOrder
	if subscription.Track.GroupOrder == 0 {
		subscription.Track.GroupOrder = info.GroupOrder
	} else {
		if info.GroupOrder != subscription.Track.GroupOrder {
			slog.Debug("GroupOrder is not updated")
			return nil, ErrGroupOrderMismatch
		}
	}

	// Update the GroupExpires
	if info.GroupExpires < subscription.Track.GroupExpires {
		subscription.Track.GroupExpires = info.GroupExpires
	}

	sentSubscription := &SendSubscribeStream{
		subscribeID:  nextID,
		Subscription: subscription,
		stream:       stream,
	}

	return sentSubscription, err
}

func (sess *session) OpenFetchStream(req Fetch) (ReceiveDataStream, error) {
	/*
	 * Open a Fetch Stream
	 */
	stream, err := openFetchStream(sess.conn)
	if err != nil {
		slog.Error("failed to open a Fetch Stream", slog.String("error", err.Error()))
		return nil, err
	}

	/*
	 * Send a FETCH message
	 */
	fm := message.FetchMessage{
		TrackPath:     req.TrackPath,
		GroupPriority: message.GroupPriority(req.GroupPriority),
		GroupSequence: message.GroupSequence(req.GroupSequence),
		FrameSequence: message.FrameSequence(req.FrameSequence),
	}

	err = fm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return nil, err
	}

	/*
	 * Receive a GROUP message
	 */
	group, err := readGroup(stream)
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return nil, err
	}

	return dataReceiveStream{
		ReceiveStream: stream,
		receivedGroup: group,
	}, nil
}

func (sess *session) OpenInfoStream(req InfoRequest) (Info, error) {
	slog.Debug("requesting information of a track", slog.Any("info request", req))

	/*
	 * Open an Info Stream
	 */
	stream, err := openInfoStream(sess.conn)
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
		TrackPriority:       TrackPriority(im.GroupPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}

	slog.Info("Successfully get track information", slog.Any("info", info))

	return info, nil
}

func (p *session) AcceptAnnounceStream(ctx context.Context) (*SendAnnounceStream, error) {
	for {
		if p.receivedSubscriptionQueue.Len() != 0 {
			return p.receivedInterestQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.receivedInterestQueue.Chan():
		}
	}
}

func (sess *session) AcceptSubscribeStream(ctx context.Context) (*ReceivedSubscribeStream, error) {
	for {
		if sess.receivedSubscriptionQueue.Len() != 0 {
			return sess.receivedSubscriptionQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receivedInterestQueue.Chan():
		}
	}
}

func (sess *session) AcceptFetchStream(ctx context.Context) (*ReceivedFetch, error) {
	for {
		if sess.receivedFetchQueue.Len() != 0 {
			return sess.receivedFetchQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receivedFetchQueue.Chan():
		}
	}
}

func (sess *session) AcceptInfoStream(ctx context.Context) (*SendInfoStream, error) {
	for {
		if sess.receivedInfoRequestQueue.Len() != 0 {
			return sess.receivedInfoRequestQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receivedInfoRequestQueue.Chan():
		}
	}
}

func (sess *session) AcceptDataStream(subscription *SendSubscribeStream, ctx context.Context) (ReceiveDataStream, error) {
	slog.Debug("accepting a data stream")

	queue, ok := sess.dataReceiveStreamQueues[subscription.SubscribeID()]
	if !ok {
		slog.Error("failed to get a data stream queue")
		return nil, errors.New("failed to get a data stream queue")
	}

	for {
		if queue.Len() > 0 {
			return queue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queue.Chan():
		default:
		}
	}
}

func (sess *session) AcceptDatagram(subscription *SendSubscribeStream, ctx context.Context) (ReceivedDatagram, error) {
	slog.Debug("accepting a datagram")

	queue, ok := sess.receivedDatagramQueues[subscription.SubscribeID()]
	if !ok {
		slog.Error("failed to get a datagram queue")
		return nil, errors.New("failed to get a datagram queue")
	}

	for {
		if queue.Len() > 0 {
			return queue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queue.Chan():
		default:
		}
	}
}

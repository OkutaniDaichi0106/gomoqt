package moqt

import (
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type subscriber interface {
	Interest(Interest) (*SentInterest, error)

	Subscribe(Subscription) (*SentSubscription, error)
	Unsubscribe(*SentSubscription)

	Fetch(Fetch) (DataReceiveStream, error)

	RequestInfo(InfoRequest) (Info, error)
}

var _ subscriber = (*Subscriber)(nil)

type Subscriber struct {
	sess *session

	*subscriberManager
}

func (s *Subscriber) Interest(interest Interest) (*SentInterest, error) {
	slog.Debug("indicating interest", slog.Any("interest", interest))
	/*
	 * Open an Announce Stream
	 */
	stream, err := openAnnounceStream(s.sess.conn)
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

	return &SentInterest{
		Interest: interest,
		active:   makeTracks(1),
		stream:   stream,
	}, nil
}

func (s *Subscriber) Subscribe(subscription Subscription) (*SentSubscription, error) {
	slog.Debug("making a subscription", slog.Any("subscription", subscription))

	// Open a Subscribe Stream
	stream, err := openSubscribeStream(s.sess.conn)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return nil, err
	}

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
		SubscribeID:      message.SubscribeID(s.getSubscribeID()),
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

	// Increment the subscribeIDCounter
	s.addSubscribeID()

	sentSubscription := &SentSubscription{
		Subscription: subscription,
		stream:       stream,
	}
	s.addSentSubscription(sentSubscription)

	return sentSubscription, err
}

func (s *Subscriber) Unsubscribe(subscription *SentSubscription) {
	// Close gracefully
	err := subscription.stream.Close()
	if err != nil {
		slog.Error("failed to close a subscribe stream", slog.String("error", err.Error()))
	}

	// Remove the subscription
	s.subscriberManager.removeSentSubscription(subscription.subscribeID)

	slog.Info("Unsubscribed")
}

func (s Subscriber) UnsubscribeWithError(subscription *SentSubscription, err error) {
	if err == nil {
		s.Unsubscribe(subscription)
		slog.Error("unsubscribe with no error")
		return
	}

	// Close with the error
	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	subscription.stream.CancelRead(code)
	subscription.stream.CancelWrite(code)

	// Remove the subscription
	s.subscriberManager.removeSentSubscription(subscription.subscribeID)

	slog.Info("Unsubscribed with an error")
}

func (s *Subscriber) Fetch(req Fetch) (DataReceiveStream, error) {
	/*
	 * Open a Fetch Stream
	 */
	stream, err := openFetchStream(s.sess.conn)
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

func (s *Subscriber) RequestInfo(req InfoRequest) (Info, error) {
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
		TrackPriority:       TrackPriority(im.GroupPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}

	slog.Info("Successfully get track information", slog.Any("info", info))

	return info, nil
}

func openAnnounceStream(conn transport.Connection) (transport.Stream, error) {
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

func openSubscribeStream(conn transport.Connection) (transport.Stream, error) {
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

func openInfoStream(conn transport.Connection) (transport.Stream, error) {
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

func openFetchStream(conn transport.Connection) (transport.Stream, error) {
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

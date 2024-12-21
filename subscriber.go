package moqt

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
	"github.com/quic-go/quic-go/quicvarint"
)

type subscriber interface {
	Interest(Interest) (*SentInterest, error)

	Subscribe(Subscription) (*SentSubscription, error)
	Unsubscribe(*SentSubscription)
	UpdateSubscription(*SentSubscription, SubscribeUpdate) error

	Fetch(Fetch) (DataReceiveStream, error)

	RequestInfo(InfoRequest) (Info, error)

	AcceptDataStream(context.Context) (DataReceiveStream, error)
}

var _ subscriber = (*Subscriber)(nil)

type Subscriber struct {
	sess *session

	*subscriberManager
}

func (s *Subscriber) AcceptDataStream(ctx context.Context) (DataReceiveStream, error) {
	slog.Debug("accepting a data stream")

	for {
		if s.dataReceiverQueue.Len() > 0 {
			return s.dataReceiverQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.dataReceiverQueue.Chan():
		default:
		}
	}

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

	for {
		// TODO:
	}

}

func (s *Subscriber) Subscribe(subscription Subscription) (*SentSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
		SubscribeID:      message.SubscribeID(s.subscribeIDCounter),
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
	s.subscribeIDCounter++

	return &SentSubscription{
		Subscription: subscription,
		stream:       stream,
	}, err
}

func (s *Subscriber) UpdateSubscription(subscription *SentSubscription, update SubscribeUpdate) error {
	subscription.mu.Lock()
	defer subscription.mu.Unlock()

	//
	slog.Debug("updating a subscription",
		slog.Any("subscription", subscription),
		slog.Any("to", update),
	)

	/*
	 * Verify the update
	 */
	// Verify if the new group range is valid
	if update.MinGroupSequence > update.MaxGroupSequence {
		slog.Debug("MinGroupSequence is larger than MaxGroupSequence")
		return ErrInvalidRange
	}
	// Verify if the minimum group sequence become larger
	if subscription.MinGroupSequence > update.MinGroupSequence {
		slog.Debug("the new MinGroupSequence is smaller than the old MinGroupSequence")
		return ErrInvalidRange
	}
	// Verify if the maximum group sequence become smaller
	if subscription.MaxGroupSequence < update.MaxGroupSequence {
		slog.Debug("the new MaxGroupSequence is larger than the old MaxGroupSequence")
		return ErrInvalidRange
	}

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	// Set parameters
	if update.SubscribeParameters == nil {
		update.SubscribeParameters = make(Parameters)
	}
	if update.DeliveryTimeout > 0 {
		update.SubscribeParameters.Add(DELIVERY_TIMEOUT, update.DeliveryTimeout)
	}
	// Initialize a SUBSCRIBE_UPDATE message
	sum := message.SubscribeUpdateMessage{
		SubscribeID:      message.SubscribeID(subscription.subscribeID),
		TrackPriority:    message.TrackPriority(update.TrackPriority),
		GroupOrder:       message.GroupOrder(update.GroupOrder),
		GroupExpires:     update.GroupExpires,
		MinGroupSequence: message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(update.MaxGroupSequence),
		Parameters:       message.Parameters(update.SubscribeParameters),
	}

	err := sum.Encode(subscription.stream)
	if err != nil {
		slog.Debug("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return err
	}

	// Receive an INFO message
	info, err := readInfo(subscription.stream)
	if err != nil {
		slog.Debug("failed to get an Info")
		return err
	}

	// Update the TrackPriority
	if info.TrackPriority == update.TrackPriority {
		subscription.Track.TrackPriority = info.TrackPriority
	} else {
		slog.Debug("TrackPriority is not updated")
		return ErrPriorityMismatch
	}

	// Update the GroupOrder
	if update.GroupOrder == 0 {
		subscription.Track.GroupOrder = info.GroupOrder
	} else {
		if info.GroupOrder != update.GroupOrder {
			slog.Debug("GroupOrder is not updated")
			return ErrGroupOrderMismatch
		}

		subscription.Track.GroupOrder = update.GroupOrder
	}

	// Update the GroupExpires
	if info.GroupExpires < update.GroupExpires {
		subscription.Track.GroupExpires = info.GroupExpires
	} else {
		subscription.Track.GroupExpires = update.GroupExpires
	}

	// Update the MinGroupSequence and MaxGroupSequence
	subscription.MinGroupSequence = update.MinGroupSequence
	subscription.MaxGroupSequence = update.MaxGroupSequence

	// Update the SubscribeParameters
	subscription.SubscribeParameters = update.SubscribeParameters

	// Update the DeliveryTimeout
	if update.DeliveryTimeout != 0 {
		subscription.Track.DeliveryTimeout = update.DeliveryTimeout
	}

	return nil
}

func (s *Subscriber) Unsubscribe(subscription *SentSubscription) {
	// Close gracefully
	err := subscription.stream.Close()
	if err != nil {
		slog.Error("failed to close a subscribe stream", slog.String("error", err.Error()))
	}

	slog.Info("Unsubscribed")
}

// func (s Subscriber) UnsubscribeWithError(subscription *SentSubscription, err error) {

// }

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
		ReceivedGroup: group,
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

func receiveDatagram(conn transport.Connection, ctx context.Context) (ReceivedGroup, []byte, error) {
	data, err := conn.ReceiveDatagram(ctx)
	if err != nil {
		slog.Error("failed to receive a datagram", slog.String("error", err.Error()))
		return ReceivedGroup{}, nil, err
	}

	reader := bytes.NewReader(data)

	group, err := readGroup(quicvarint.NewReader(reader))
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return ReceivedGroup{}, nil, err
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

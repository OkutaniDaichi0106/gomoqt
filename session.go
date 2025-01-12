package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type Session interface {
	// Terminate the session
	Terminate(error)

	// Open an Announce Stream
	OpenAnnounceStream(Interest) (ReceiveAnnounceStream, error)

	// Open a Subscribe Stream
	OpenSubscribeStream(Subscription) (SendSubscribeStream, error)

	// Open a Fetch Stream
	OpenFetchStream(FetchRequest) (SendFetchStream, error)

	// Open an Info Stream
	OpenInfoStream(InfoRequest) (Info, error)

	// Open a Data Stream
	OpenDataStream(ReceiveSubscribeStream, GroupSequence, GroupPriority) (SendDataStream, error)

	// Send a Datagram
	SendDatagram(ReceiveSubscribeStream, GroupSequence, GroupPriority, []byte) error

	// Accept an Announce Stream
	AcceptAnnounceStream(context.Context) (SendAnnounceStream, error)

	// Accept a Subscribe Stream
	AcceptSubscribeStream(context.Context) (ReceiveSubscribeStream, error)

	// Accept a Fetch Stream
	AcceptFetchStream(context.Context) (ReceiveFetchStream, error)

	// Accept an Info Stream
	AcceptInfoStream(context.Context) (SendInfoStream, error)

	// Accept a Data Stream
	AcceptDataStream(SendSubscribeStream, context.Context) (ReceiveDataStream, error)

	// Accept a Datagram
	AcceptDatagram(SendSubscribeStream, context.Context) (ReceivedDatagram, error)
}

var _ Session = (*session)(nil)

func newSession(conn transport.Connection, stream transport.Stream) *session {
	return &session{
		conn:                        conn,
		stream:                      stream,
		receiveSubscribeStreamQueue: newReceiveSubscribeStreamQueue(),
		sendAnnounceStreamQueue:     newReceivedInterestQueue(),
		receiveFetchStreamQueue:     newReceivedFetchQueue(),
		receivedInfoRequestQueue:    newReceivedInfoRequestQueue(),
		dataReceiveStreamQueues:     make(map[SubscribeID]*receiveDataStreamQueue),
		receivedDatagramQueues:      make(map[SubscribeID]*receivedDatagramQueue),
		subscribeIDCounter:          0,
	}
}

type session struct {
	conn   transport.Connection
	stream transport.Stream

	//
	receiveSubscribeStreamQueue *receiveSubscribeStreamQueue

	sendAnnounceStreamQueue *receivedInterestQueue

	receiveFetchStreamQueue *receivedFetchQueue

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

func (s *session) OpenAnnounceStream(interest Interest) (ReceiveAnnounceStream, error) {
	slog.Debug("indicating interest", slog.Any("interest", interest))
	/*
	 * Open an Announce Stream
	 */
	stream, err := openAnnounceStream(s.conn)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return nil, err
	}

	err = writeInterest(stream, interest)
	if err != nil {
		slog.Error("failed to write an Interest message", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Info("Successfully indicated interest", slog.Any("interest", interest))

	ras := &receiveAnnounceStream{
		interest: interest,
		stream:   stream,
		ch:       make(chan struct{}, 1),
	}

	// Receive Announcements
	go func() {
		var terr error
		// Read announcements
		for {
			slog.Debug("reading an announcement")

			// Read an ANNOUNCE message
			ann, err := readAnnouncement(stream, ras.interest.TrackPrefix)
			if err != nil {
				slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
				return
			}

			oldAnn, ok := ras.annMap[ann.TrackPath]

			if ok && oldAnn.AnnounceStatus == ann.AnnounceStatus {
				slog.Debug("duplicate announcement status")
				terr = ErrProtocolViolation
				break
			}

			if !ok && ann.AnnounceStatus == ENDED {
				slog.Debug("ended track is not announced")
				terr = ErrProtocolViolation
				break
			}

			switch ann.AnnounceStatus {
			case ACTIVE, ENDED:
				ras.annMap[ann.TrackPath] = ann
			case LIVE:
				ras.ch <- struct{}{}
			}
		}

		s.Terminate(terr)
	}()

	return ras, nil
}

func (s *session) OpenSubscribeStream(subscription Subscription) (SendSubscribeStream, error) {
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
	err = writeSubscription(stream, nextID, subscription)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()))
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
	if info.TrackPriority != subscription.TrackPriority {
		slog.Debug("TrackPriority is not updated")
		return nil, ErrPriorityMismatch
	}

	// Update the GroupOrder
	if subscription.GroupOrder == 0 {
		subscription.GroupOrder = info.GroupOrder
	} else {
		if info.GroupOrder != subscription.GroupOrder {
			slog.Debug("GroupOrder is not updated")
			return nil, ErrGroupOrderMismatch
		}
	}

	// Update the GroupExpires
	if info.GroupExpires < subscription.GroupExpires {
		subscription.GroupExpires = info.GroupExpires
	}

	// Create a data stream queue
	s.dataReceiveStreamQueues[nextID] = newReceiveDataStreamQueue()

	// Create a datagram queue
	s.receivedDatagramQueues[nextID] = newReceivedDatagramQueue()

	return &sendSubscribeStream{
		subscribeID:  nextID,
		subscription: subscription,
		stream:       stream,
	}, err
}

func (sess *session) OpenFetchStream(fetch FetchRequest) (SendFetchStream, error) {
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
	err = writeFetch(stream, fetch)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return nil, err
	}

	return &sendFetchStream{
		stream: stream,
		fetch:  fetch,
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
	err = writeInfoRequest(stream, req)
	if err != nil {
		slog.Error("failed to send an INFO_REQUEST message", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 * Receive a INFO message
	 */
	info, err := readInfo(stream)
	if err != nil {
		slog.Error("failed to get a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	// Close the stream
	err = stream.Close()
	if err != nil {
		slog.Error("failed to close the stream", slog.String("error", err.Error()))
		return Info{}, err
	}

	slog.Info("Successfully get track information", slog.Any("info", info))

	return info, nil
}

func (sess *session) OpenDataStream(substr ReceiveSubscribeStream, sequence GroupSequence, priority GroupPriority) (SendDataStream, error) {
	// Verify
	if sequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	// Open
	stream, err := openGroupStream(sess.conn)
	if err != nil {
		slog.Error("failed to open a group stream", slog.String("error", err.Error()))
		return nil, err
	}

	group := sentGroup{
		groupSequence: sequence,
		groupPriority: priority,
		sentAt:        time.Now(),
	}

	// Send the GROUP message
	err = writeGroup(stream, substr.SubscribeID(), group)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return sendDataStream{
			SendStream: stream,
			sentGroup:  group,
		},
		nil
}

func (sess *session) SendDatagram(substr ReceiveSubscribeStream, sequence GroupSequence, priority GroupPriority, payload []byte) error {
	// Verify
	if sequence == 0 {
		return errors.New("0 sequence number")
	}

	group := sentGroup{
		groupSequence: sequence,
		groupPriority: priority,
		sentAt:        time.Now(),
	}

	// Send
	err := sendDatagram(sess.conn, substr.SubscribeID(), group, payload)
	if err != nil {
		slog.Error("failed to send a datagram", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (p *session) AcceptAnnounceStream(ctx context.Context) (SendAnnounceStream, error) {
	for {
		if p.receiveSubscribeStreamQueue.Len() != 0 {
			return p.sendAnnounceStreamQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.sendAnnounceStreamQueue.Chan():
		}
	}
}

func (sess *session) AcceptSubscribeStream(ctx context.Context) (ReceiveSubscribeStream, error) {
	for {
		if sess.receiveSubscribeStreamQueue.Len() != 0 {
			return sess.receiveSubscribeStreamQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.sendAnnounceStreamQueue.Chan():
		}
	}
}

func (sess *session) AcceptFetchStream(ctx context.Context) (ReceiveFetchStream, error) {
	for {
		if sess.receiveFetchStreamQueue.Len() != 0 {
			return sess.receiveFetchStreamQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receiveFetchStreamQueue.Chan():
		}
	}
}

func (sess *session) AcceptInfoStream(ctx context.Context) (SendInfoStream, error) {
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

func (sess *session) AcceptDataStream(subscription SendSubscribeStream, ctx context.Context) (ReceiveDataStream, error) {
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

func (sess *session) AcceptDatagram(substr SendSubscribeStream, ctx context.Context) (ReceivedDatagram, error) {
	slog.Debug("accepting a datagram")

	queue, ok := sess.receivedDatagramQueues[substr.SubscribeID()]
	if !ok {
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

/*
 *
 *
 */

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

func openGroupStream(conn transport.Connection) (transport.SendStream, error) {
	slog.Debug("opening an Group Stream")

	stream, err := conn.OpenUniStream()
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

func sendDatagram(conn transport.Connection, id SubscribeID, g sentGroup, payload []byte) error {
	if g.groupSequence == 0 {
		return errors.New("0 sequence number")
	}

	var buf bytes.Buffer

	// Encode the group
	err := writeGroup(&buf, id, g)
	if err != nil {
		return err
	}

	// Encode the payload
	_, err = buf.Write(payload)
	if err != nil {
		slog.Error("failed to encode a payload", slog.String("error", err.Error()))
		return err
	}

	// Send the data with the GROUP message
	err = conn.SendDatagram(buf.Bytes())
	if err != nil {
		slog.Error("failed to send a datagram", slog.String("error", err.Error()))
		return err
	}

	return nil
}

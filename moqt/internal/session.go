package internal

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func OpenSessionStream(conn transport.Connection, scm message.SessionClientMessage) (*SessionStream, message.SessionServerMessage, error) {
	// Open a Session Stream
	stream, err := openSessionStream(conn)
	if err != nil {
		slog.Error("failed to open a Session Stream")
		return nil, message.SessionServerMessage{}, err
	}

	// Send a set-up request
	_, err = scm.Encode(stream)
	if err != nil {
		slog.Error("failed to request to set up", slog.String("error", err.Error()))
		return nil, message.SessionServerMessage{}, err
	}

	// Receive a set-up responce
	var ssm message.SessionServerMessage
	_, err = ssm.Decode(stream)
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", slog.String("error", err.Error()))
		return nil, message.SessionServerMessage{}, err
	}

	return &SessionStream{Stream: stream}, ssm, nil
}

func AcceptSessionStream(conn transport.Connection, ctx context.Context) (*SessionStream, *message.SessionClientMessage, error) {
	stream, err := acceptSessionStream(conn)
	if err != nil {
		slog.Error("failed to accept a session stream", slog.String("error", err.Error()))
		return nil, nil, err
	}

	var scm message.SessionClientMessage
	_, err = scm.Decode(stream)
	if err != nil {
		slog.Error("failed to decode a session client message", slog.String("error", err.Error()))
		return nil, nil, err
	}

	if !ContainVersion(DefaultServerVersion, scm.SupportedVersions) {
		slog.Error("no available version", slog.Any("versions", scm.SupportedVersions))
		return nil, nil, ErrProtocolViolation // TODO:
	}

	return &SessionStream{Stream: stream}, &scm, nil
}

func NewSession(conn transport.Connection, stream *SessionStream) *Session {
	sess := &Session{
		conn:                        conn,
		SessionStream:               *stream,
		receiveSubscribeStreamQueue: newReceiveSubscribeStreamQueue(),
		sendAnnounceStreamQueue:     newSendAnnounceStreamQueue(),
		sendInfoStreamQueue:         newReceiveInfoStreamQueue(),
		receiveGroupStreamQueues:    make(map[message.SubscribeID]*receiveGroupStreamQueue),
	}

	go listenSession(sess, context.Background())

	return sess
}

type Session struct {
	conn transport.Connection
	SessionStream

	//
	receiveSubscribeStreamQueue *receiveSubscribeStreamQueue

	sendAnnounceStreamQueue *sendAnnounceStreamQueue

	sendInfoStreamQueue *sendInfoStreamQueue

	receiveGroupStreamQueues map[message.SubscribeID]*receiveGroupStreamQueue
}

func (sess *Session) Terminate(err error) {
	slog.Info("Terminating a session", slog.String("reason", err.Error()))

	var tererr TerminateError

	if err == nil {
		tererr = NoErrTerminate
	} else {
		if errors.As(err, &tererr) {
		} else {
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

func (s *Session) OpenAnnounceStream(apm message.AnnouncePleaseMessage) (*ReceiveAnnounceStream, error) {
	slog.Debug(("opening an Announce Stream"), slog.Any("config", apm))

	// Open an Announce Stream
	stream, err := openAnnounceStream(s.conn)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return nil, err
	}

	_, err = apm.Encode(stream)
	if err != nil {
		slog.Error("failed to write an Interest message", slog.String("error", err.Error()))
		return nil, err
	}

	ras := newReceiveAnnounceStream(&apm, stream)

	slog.Debug("Opened an announce stream", slog.Any("config", apm))

	return ras, nil
}

func (s *Session) OpenSubscribeStream(sm message.SubscribeMessage) (message.InfoMessage, *SendSubscribeStream, error) {
	slog.Debug("opening a subscribe stream", slog.Any("config", sm))

	// Open a Subscribe Stream
	stream, err := openSubscribeStream(s.conn)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return message.InfoMessage{}, nil, err
	}

	// Send a SUBSCRIBE message
	_, err = sm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()))
		return message.InfoMessage{}, nil, err
	}

	// Receive an INFO message
	var im message.InfoMessage
	_, err = im.Decode(stream)
	if err != nil {
		slog.Error("failed to get a Info", slog.String("error", err.Error()))
		return message.InfoMessage{}, nil, err
	}

	// Create a receive group stream queue
	s.receiveGroupStreamQueues[sm.SubscribeID] = newGroupReceiverQueue(sm)

	slog.Debug("Successfully opened a subscribe stream", slog.Any("config", sm), slog.Any("info", im))

	return im, newSendSubscribeStream(&sm, stream), err

}

func (sess *Session) OpenInfoStream(irm message.InfoRequestMessage) (message.InfoMessage, error) {
	slog.Debug("requesting information of a track", slog.Any("info request", irm))

	// Open an Info Stream
	stream, err := openInfoStream(sess.conn)
	if err != nil {
		slog.Error("failed to open an Info Stream", slog.String("error", err.Error()))
		return message.InfoMessage{}, err
	}

	// Close the stream
	defer stream.Close()

	// Send an INFO_REQUEST message
	_, err = irm.Encode(stream)
	if err != nil {
		slog.Error("failed to send an INFO_REQUEST message", slog.String("error", err.Error()))

		return message.InfoMessage{}, err
	}

	// Receive a INFO message
	var im message.InfoMessage
	_, err = im.Decode(stream)
	if err != nil {
		slog.Error("failed to get a INFO message", slog.String("error", err.Error()))
		return message.InfoMessage{}, err
	}

	slog.Info("Successfully get track information", slog.Any("info", im))

	return im, nil
}

func (sess *Session) AcceptGroupStream(ctx context.Context, id message.SubscribeID) (*ReceiveGroupStream, error) {
	_, ok := sess.receiveGroupStreamQueues[id]
	if !ok {
		return nil, ErrProtocolViolation // TODO:
	}

	for {
		if sess.receiveGroupStreamQueues[id].Len() != 0 {
			return sess.receiveGroupStreamQueues[id].Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receiveGroupStreamQueues[id].Chan():
		}
	}
}

func (p *Session) AcceptAnnounceStream(ctx context.Context) (*SendAnnounceStream, error) {
	for {
		if p.sendAnnounceStreamQueue.Len() != 0 {
			annstr := p.sendAnnounceStreamQueue.Dequeue()
			return annstr, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.sendAnnounceStreamQueue.Chan():
		}
	}
}

func (sess *Session) AcceptSubscribeStream(ctx context.Context) (*ReceiveSubscribeStream, error) {
	for {
		if sess.receiveSubscribeStreamQueue.Len() != 0 {

			substr := sess.receiveSubscribeStreamQueue.Dequeue()
			return substr, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receiveSubscribeStreamQueue.Chan():
		}
	}
}

func (sess *Session) AcceptInfoStream(ctx context.Context) (*SendInfoStream, error) {
	for {
		if sess.sendInfoStreamQueue.Len() != 0 {
			infostr := sess.sendInfoStreamQueue.Dequeue()
			return infostr, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.sendInfoStreamQueue.Chan():
		}
	}
}

func (sess *Session) OpenGroupStream(gm message.GroupMessage) (*SendGroupStream, error) {
	slog.Debug("opening a Group Stream")

	stream, err := openGroupStream(sess.conn)
	if err != nil {
		slog.Error("failed to open a Group Stream", slog.String("error", err.Error()))
		return nil, err
	}

	_, err = gm.Encode(stream)
	if err != nil {
		slog.Error("failed to write a Group message", slog.String("error", err.Error()))
		return nil, err
	}

	return newSendGroupStream(&gm, stream), nil
}

func listenSession(sess *Session, ctx context.Context) {
	wg := new(sync.WaitGroup)
	// Listen bidirectional streams
	wg.Add(1)
	go func() {
		defer wg.Done()
		listenBiStreams(sess, ctx)
	}()

	// Listen unidirectional streams
	wg.Add(1)
	go func() {
		defer wg.Done()
		listenUniStreams(sess, ctx)
	}()

	wg.Wait()
}

func listenBiStreams(sess *Session, ctx context.Context) {
	for {
		/*
		 * Accept a bidirectional stream
		 */
		stream, err := sess.conn.AcceptStream(ctx)
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some control stream was opened")

		// Handle the stream
		go func(stream transport.Stream) {
			/*
			 * Get a Stream Type ID
			 */
			var stm message.StreamTypeMessage
			_, err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			// Handle the stream by the Stream Type ID
			switch stm.StreamType {
			case stream_type_announce:
				// Handle the announce stream
				slog.Debug("announce stream was opened")
				// Get an Interest
				var apm message.AnnouncePleaseMessage
				_, err := apm.Decode(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					stream.CancelRead(ErrInternalError.StreamErrorCode())
					stream.CancelWrite(ErrInternalError.StreamErrorCode())
					return
				}

				sas := newSendAnnounceStream(stream, apm)

				// Enqueue the interest
				sess.sendAnnounceStreamQueue.Enqueue(sas)
			case stream_type_subscribe:
				slog.Debug("subscribe stream was opened")

				var sm message.SubscribeMessage
				_, err := sm.Decode(stream)
				if err != nil {
					slog.Debug("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
					stream.CancelRead(ErrInternalError.StreamErrorCode())
					stream.CancelWrite(ErrInternalError.StreamErrorCode())
					return
				}

				// Create a receiveSubscribeStream
				rss := newReceiveSubscribeStream(&sm, stream)

				// Enqueue the subscription
				sess.receiveSubscribeStreamQueue.Enqueue(rss)
			case stream_type_info:
				slog.Debug("info stream was opened")

				// Get a received info-request
				var imr message.InfoRequestMessage
				_, err := imr.Decode(stream)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				sis := newSendInfoStream(stream, &imr)

				// Enqueue the info-request
				sess.sendInfoStreamQueue.Enqueue(sis)
			default:
				slog.Debug("An unknown type of stream was opend")

				// Terminate the session
				sess.Terminate(ErrProtocolViolation)

				return
			}
		}(stream)
	}
}

func listenUniStreams(sess *Session, ctx context.Context) {
	for {
		/*
		 * Accept a unidirectional stream
		 */
		stream, err := sess.conn.AcceptUniStream(ctx)
		if err != nil {
			slog.Error("failed to accept a unidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some data stream was opened")

		// Handle the stream
		go func(stream transport.ReceiveStream) {
			/*
			 * Get a Stream Type ID
			 */
			var stm message.StreamTypeMessage
			_, err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			// Handle the stream by the Stream Type ID
			switch stm.StreamType {
			case stream_type_group:
				slog.Debug("group stream was opened")

				var gm message.GroupMessage
				_, err := gm.Decode(stream)
				if err != nil {
					slog.Error("failed to get a group", slog.String("error", err.Error()))
					return
				}

				rgs := newReceiveGroupStream(&gm, stream)

				queue, ok := sess.receiveGroupStreamQueues[gm.SubscribeID]
				if !ok {
					slog.Error("failed to get a data receive stream queue", slog.String("error", "queue not found"))
					stream.CancelRead(ErrInternalError.StreamErrorCode())
					return
				}

				// Enqueue the receiver
				queue.Enqueue(rgs)
			default:
				slog.Debug("An unknown type of stream was opened")

				// Terminate the session
				sess.Terminate(ErrProtocolViolation)

				return
			}
		}(stream)
	}
}

/*
 *
 *
 */
func openAnnounceStream(conn transport.Connection) (transport.Stream, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_announce,
	}

	_, err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openSubscribeStream(conn transport.Connection) (transport.Stream, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_subscribe,
	}

	_, err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openInfoStream(conn transport.Connection) (transport.Stream, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_info,
	}

	_, err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openGroupStream(conn transport.Connection) (transport.SendStream, error) {
	stream, err := conn.OpenUniStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_group,
	}

	_, err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

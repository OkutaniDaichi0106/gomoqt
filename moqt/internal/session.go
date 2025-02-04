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
		conn:                         conn,
		SessionStream:                *stream,
		receiveSubscribeStreamQueue:  newReceiveSubscribeStreamQueue(),
		sendAnnounceStreamQueue:      newSendAnnounceStreamQueue(),
		receiveFetchStreamQueue:      newReceivedFetchQueue(),
		sendInfoStreamQueue:          newReceiveInfoStreamQueue(),
		dataReceiveGroupStreamQueues: make(map[message.SubscribeID]*receiveGroupStreamQueue),
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

	receiveFetchStreamQueue *receiveFetchStreamQueue

	sendInfoStreamQueue *sendInfoStreamQueue

	dataReceiveGroupStreamQueues map[message.SubscribeID]*receiveGroupStreamQueue
}

func (sess *Session) Terminate(err error) {
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

func (s *Session) OpenSubscribeStream(sm message.SubscribeMessage) (*SendSubscribeStream, message.InfoMessage, error) {
	slog.Debug("opening a subscribe stream", slog.Any("config", sm))

	// Open a Subscribe Stream
	stream, err := openSubscribeStream(s.conn)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
		return nil, message.InfoMessage{}, err
	}

	// Send a SUBSCRIBE message
	_, err = sm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()))
		return nil, message.InfoMessage{}, err
	}

	// Receive an INFO message
	var im message.InfoMessage
	_, err = im.Decode(stream)
	if err != nil {
		slog.Error("failed to get a Info", slog.String("error", err.Error()))
		return nil, message.InfoMessage{}, err
	}

	// Create a data stream queue
	s.dataReceiveGroupStreamQueues[sm.SubscribeID] = newGroupReceiverQueue()

	slog.Debug("Successfully opened a subscribe stream", slog.Any("config", sm), slog.Any("info", im))

	return newSendSubscribeStream(&sm, stream), im, err
}

func (sess *Session) OpenFetchStream(fm message.FetchMessage) (*SendFetchStream, error) {
	// Open a Fetch Stream
	stream, err := openFetchStream(sess.conn)
	if err != nil {
		slog.Error("failed to open a Fetch Stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Send a FETCH message
	_, err = fm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return nil, err
	}

	return newSendFetchStream(&fm, stream), nil
}

func (sess *Session) OpenInfoStream(irm message.InfoRequestMessage) (message.InfoMessage, error) {
	slog.Debug("requesting information of a track", slog.Any("info request", irm))

	// Open an Info Stream
	stream, err := openInfoStream(sess.conn)
	if err != nil {
		slog.Error("failed to open an Info Stream", slog.String("error", err.Error()))
		return message.InfoMessage{}, err
	}

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

	// Close the stream
	err = stream.Close()
	if err != nil {
		slog.Error("failed to close the stream", slog.String("error", err.Error()))
		return message.InfoMessage{}, err
	}

	slog.Info("Successfully get track information", slog.Any("info", im))

	return im, nil
}

func (sess *Session) AcceptGroupStream(ctx context.Context, id message.SubscribeID) (*ReceiveGroupStream, error) {
	_, ok := sess.dataReceiveGroupStreamQueues[id]
	if !ok {
		slog.Error("failed to get a data receive stream queue", slog.String("error", "queue not found"))
		sess.dataReceiveGroupStreamQueues[id] = newGroupReceiverQueue()
	}

	for {
		if sess.dataReceiveGroupStreamQueues[id].Len() != 0 {
			return sess.dataReceiveGroupStreamQueues[id].Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.dataReceiveGroupStreamQueues[id].Chan():
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

func (sess *Session) AcceptFetchStream(ctx context.Context) (*ReceiveFetchStream, error) {
	for {
		if sess.receiveFetchStreamQueue.Len() != 0 {
			fetstr := sess.receiveFetchStreamQueue.Dequeue()
			return fetstr, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sess.receiveFetchStreamQueue.Chan():
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

	// go func() {
	// 	select {
	// 	case code := <-errCodeCh:
	// 		substr.SendSubscribeGap(message.SubscribeGapMessage{
	// 			MinGapSequence: seq,
	// 			MaxGapSequence: seq,
	// 			GroupErrorCode: code,
	// 		})
	// 	}
	// }()
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
					closeStreamWithInternalError(stream, err)
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
					closeStreamWithInternalError(stream, err)
					return
				}

				// Create a receiveSubscribeStream
				rss := newReceiveSubscribeStream(&sm, stream)

				// Enqueue the subscription
				sess.receiveSubscribeStreamQueue.Enqueue(rss)
			case stream_type_fetch:
				slog.Debug("fetch stream was opened")
				// Get a fetch-request
				var fm message.FetchMessage
				_, err := fm.Decode(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					closeStreamWithInternalError(stream, err)
					return
				}

				rfs := newReceiveFetchStream(&fm, stream)

				// Enqueue the fetch
				sess.receiveFetchStreamQueue.Enqueue(rfs)
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

				data := newReceiveGroupStream(&gm, stream)

				queue, ok := sess.dataReceiveGroupStreamQueues[gm.SubscribeID]
				if !ok {
					slog.Error("failed to get a data receive stream queue", slog.String("error", "queue not found"))
					closeReceiveStreamWithInternalError(stream, ErrInternalError) // TODO:
					return
				}

				// Enqueue the receiver
				queue.Enqueue(data)
			default:
				slog.Debug("An unknown type of stream was opend")

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

func openFetchStream(conn transport.Connection) (transport.Stream, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_fetch,
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

func closeStreamWithInternalError(stream transport.Stream, err error) {

	slog.Debug("closing the stream with an internal error", slog.String("error", err.Error()))

	if err == nil {
		stream.Close()
	}

	// TODO:

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

	stream.CancelRead(code)
	stream.CancelWrite(code)

	slog.Debug("closed the stream with an internal error")
}

func closeReceiveStreamWithInternalError(stream transport.ReceiveStream, err error) {

	slog.Debug("closing the receive stream with an internal error", slog.String("error", err.Error()))

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

	stream.CancelRead(code)

	slog.Debug("closed the receive stream with an internal error")
}

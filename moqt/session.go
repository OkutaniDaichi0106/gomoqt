package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type Session interface {
	/*
	 * Methods for the Client
	 */
	// Update the session
	// UpdateSession(bitrate uint64) error // TODO:

	// Terminate the session
	Terminate(error)

	/*
	 * Methods for the Subscriber
	 */
	// Open an Announce Stream
	OpenAnnounceStream(*AnnounceConfig) (AnnouncementReader, error)

	// Open a Track Stream
	OpenTrackStream(TrackPath, *SubscribeConfig) (Info, ReceiveTrackStream, error)

	// Request Track Info
	GetInfo(TrackPath) (Info, error)
}

var _ Session = (*session)(nil)

type session struct {
	conn quic.Connection

	subscribeIDCounter uint64

	bitrate uint64 // TODO: use this when updating a session

	// sessionStreamCh is the channel for signaling session streams
	sessionStreamCh chan *sessionStream

	// sessionStream is the session stream for the session
	sessionStream *sessionStream

	// once               sync.Once // TODO: use this if needed

	receiveSubscribeStreamQueue *receiveSubscribeStreamQueue

	sendAnnounceStreamQueue *sendAnnounceStreamQueue

	sendInfoStreamQueue *sendInfoStreamQueue

	receiveGroupStreamQueues map[SubscribeID]*receiveGroupStreamQueue
}

func (s *session) Terminate(err error) {
	var tererr TerminateError
	if err == nil {
		tererr = NoErrTerminate
	} else {
		if !errors.As(err, &tererr) {
			tererr = ErrInternalError.WithReason(err.Error())
		}
	}

	code := quic.ConnectionErrorCode(tererr.TerminateErrorCode())
	reason := tererr.Error()

	err = s.conn.CloseWithError(code, reason)
	if err != nil {
		slog.Error("failed to close the Connection", "error", err)
		return
	}

	slog.Info("Terminated a session")
}

func (s *session) OpenAnnounceStream(config *AnnounceConfig) (AnnouncementReader, error) {
	if config == nil {
		config = &AnnounceConfig{TrackPattern: "/**"}
	}

	return s.openAnnounceStream(config)
}

func (s *session) OpenTrackStream(path TrackPath, config *SubscribeConfig) (Info, ReceiveTrackStream, error) {
	if config == nil {
		config = &SubscribeConfig{}
	}

	id := s.nextSubscribeID()

	slog.Debug("opening track stream", "subscribe_config", config.String(), "subscribe_id", id)

	im, ss, err := s.openSubscribeStream(id, path, *config)
	if err != nil {
		return NotFoundInfo, nil, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}

	return info, newReceiveTrackStream(s, info, ss), nil
}

func (s *session) GetInfo(path TrackPath) (Info, error) {
	slog.Debug("requesting track info", "track_path", path)

	im, err := s.openInfoStream(message.InfoRequestMessage{
		TrackPath: string(path),
	})
	if err != nil {
		slog.Error("failed to request track info",
			"track_path", path,
			"error", err,
		)
		return NotFoundInfo, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}

	slog.Debug("received track info", "info", info.String())

	return info, nil
}

// func (s *session) URI() *url.URL {
// 	return s.uri
// }

func (s *session) nextSubscribeID() SubscribeID {
	// Increment and return the previous value atomically
	id := atomic.AddUint64(&s.subscribeIDCounter, 1) - 1
	return SubscribeID(id)
}

func newSession(conn quic.Connection) *session {
	sess := &session{
		conn:                        conn,
		receiveSubscribeStreamQueue: newReceiveSubscribeStreamQueue(),
		sendAnnounceStreamQueue:     newSendAnnounceStreamQueue(),
		sendInfoStreamQueue:         newSendInfoStreamQueue(),
		receiveGroupStreamQueues:    make(map[SubscribeID]*receiveGroupStreamQueue),
		bitrate:                     0,
		sessionStreamCh:             make(chan *sessionStream),
	}

	wg := new(sync.WaitGroup) // TODO: sync.WaitGroup
	ctx := context.TODO()     // TODO: context.TODO()?

	// Listen bidirectional streams
	wg.Add(1)
	go func() {
		wg.Done()
		sess.listenBiStreams(ctx)
	}()

	// Listen unidirectional streams
	wg.Add(1)
	go func() {
		wg.Done()
		sess.listenUniStreams(ctx)
	}()

	wg.Wait()

	return sess
}

// TODO: Implement this method and use it
func (sess *session) updateSession(bitrate uint64) error {
	slog.Debug("updating a session", slog.Uint64("bitrate", bitrate))

	// Send a SESSION_UPDATE message
	err := sess.sessionStream.UpdateSession(bitrate)
	if err != nil {
		slog.Error("failed to update a session",
			"error", err,
		)
		return err
	}

	// Update the bitrate
	sess.bitrate = bitrate

	return nil
}

func (sess *session) openSessionStream(versions []protocol.Version, params *Parameters) error {
	slog.Debug("opening a session stream")

	// Close the session stream channel
	close(sess.sessionStreamCh)

	stream, err := openStream(sess.conn, stream_type_session)
	if err != nil {
		slog.Error("failed to open a session stream", "error", err)
		return err
	}

	// Send a SESSION_CLIENT message
	scm := message.SessionClientMessage{
		SupportedVersions: versions,
	}
	if scm.Parameters != nil {
		scm.Parameters = params.paramMap
	}
	_, err = scm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a SESSION_CLIENT message", "error", err)
		return err
	}

	// Receive a set-up response
	var ssm message.SessionServerMessage
	_, err = ssm.Decode(stream)
	if err != nil {
		slog.Error("failed to receive a SESSION_SERVER message", "error", err)
		return err
	}

	// Set the selected version and parameters
	sess.sessionStream = newSessionStream(
		stream,
		ssm.SelectedVersion,
		params,
		&Parameters{ssm.Parameters},
	)

	slog.Debug("opened a session stream")

	return nil
}

func (s *session) openAnnounceStream(config *AnnounceConfig) (*receiveAnnounceStream, error) {
	apm := message.AnnouncePleaseMessage{
		TrackPattern: config.TrackPattern,
	}

	slog.Debug("opening an announce stream", slog.Any("config", apm))

	// Open an Announce Stream
	stream, err := openStream(s.conn, stream_type_announce)
	if err != nil {
		slog.Error("failed to open an Announce Stream")
		return nil, err
	}

	_, err = apm.Encode(stream)
	if err != nil {
		slog.Error("failed to write an Interest message", "error", err)
		return nil, err
	}

	slog.Debug("opened an announce stream", "announce_config", config.String())

	return newReceiveAnnounceStream(stream, config), nil
}

func (s *session) openSubscribeStream(id SubscribeID, path TrackPath, config SubscribeConfig) (Info, *sendSubscribeStream, error) {
	// Open a Subscribe Stream
	stream, err := openStream(s.conn, stream_type_subscribe)
	if err != nil {
		slog.Error("failed to open a Subscribe Stream", "error", err)
		return NotFoundInfo, nil, err
	}

	// Send a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(id),
		TrackPath:        string(path),
		GroupOrder:       message.GroupOrder(config.GroupOrder),
		TrackPriority:    message.TrackPriority(config.TrackPriority),
		MinGroupSequence: message.GroupSequence(config.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(config.MaxGroupSequence),
	}
	_, err = sm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", "error", err)
		return NotFoundInfo, nil, err
	}

	// Receive an INFO message
	var im message.InfoMessage
	_, err = im.Decode(stream)
	if err != nil {
		slog.Error("failed to get a Info", "error", err)
		return NotFoundInfo, nil, err
	}

	// Create a receive group stream queue
	s.receiveGroupStreamQueues[id] = newGroupReceiverQueue(id, path, config)

	slog.Debug("Successfully opened a subscribe stream", slog.Any("config", sm), slog.Any("info", im))

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}

	return info, newSendSubscribeStream(id, path, config, stream), nil
}

func (sess *session) openInfoStream(irm message.InfoRequestMessage) (message.InfoMessage, error) {
	slog.Debug("requesting information of a track", slog.Any("info request", irm))

	// Open an Info Stream
	stream, err := openStream(sess.conn, stream_type_info)
	if err != nil {
		slog.Error("failed to open an Info Stream", "error", err)
		return message.InfoMessage{}, err
	}

	// Close the stream
	defer stream.Close()

	// Send an INFO_REQUEST message
	_, err = irm.Encode(stream)
	if err != nil {
		slog.Error("failed to send an INFO_REQUEST message", "error", err)

		return message.InfoMessage{}, err
	}

	// Receive a INFO message
	var im message.InfoMessage
	_, err = im.Decode(stream)
	if err != nil {
		slog.Error("failed to get a INFO message", "error", err)
		return message.InfoMessage{}, err
	}

	slog.Info("Successfully get track information", slog.Any("info", im))

	return im, nil
}

func (sess *session) openGroupStream(id SubscribeID, sequence GroupSequence) (*sendGroupStream, error) {
	stream, err := openStream(sess.conn, stream_type_group)
	if err != nil {
		slog.Error("failed to open a Group Stream", "error", err)
		return nil, err
	}

	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(id),
		GroupSequence: message.GroupSequence(sequence),
	}
	_, err = gm.Encode(stream)
	if err != nil {
		return nil, err
	}

	return newSendGroupStream(stream, id, sequence), nil
}

func (sess *session) acceptSessionStream(ctx context.Context, params func(*Parameters) (*Parameters, error)) error {
	if sess.sessionStream != nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case stream := <-sess.sessionStreamCh:
		if stream == nil {
			return errors.New("session stream is nil")
		}
		close(sess.sessionStreamCh)

		serverParams, err := params(stream.clientParameters)
		if err != nil {
			return err
		}

		version := internal.DefaultServerVersion

		// Send a SESSION_SERVER message
		ssm := message.SessionServerMessage{
			SelectedVersion: version,
			Parameters:      serverParams.paramMap,
		}

		_, err = ssm.Encode(sess.sessionStream.stream)
		if err != nil {
			slog.Error("failed to send a SESSION_SERVER message", "error", err)
			return err
		}

		// Set the selected version and parameters
		stream.selectedVersion = version

		// Set the server parameters
		stream.serverParameters = serverParams // TODO: Is this necessary?

		sess.sessionStream = stream

		return nil
	}
}

func (sess *session) acceptGroupStream(ctx context.Context, id SubscribeID) (*receiveGroupStream, error) {
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

func (p *session) acceptAnnounceStream(ctx context.Context) (*sendAnnounceStream, error) {
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

func (sess *session) acceptSubscribeStream(ctx context.Context) (*receiveSubscribeStream, error) {
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

func (sess *session) acceptInfoStream(ctx context.Context) (*sendInfoStream, error) {
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

// listenBiStreams accepts bidirectional streams and handles them based on their type.
// It listens for incoming streams and processes them in separate goroutines.
// The function handles session streams, announce streams, subscribe streams, and info streams.
// It also handles errors and terminates the session if an unknown stream type is encountered.
func (sess *session) listenBiStreams(ctx context.Context) {
	for {
		// Accept a bidirectional stream
		stream, err := sess.conn.AcceptStream(ctx)
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", "error", err)
			return
		}

		slog.Debug("A some stream was opened")

		// Handle the stream
		go func(stream quic.Stream) {
			// Decode the STREAM_TYPE message and get the stream type ID
			var stm message.StreamTypeMessage
			_, err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID",
					"error", err,
					"stream_id", stream.StreamID(),
				)
				return
			}

			// Handle the stream by the Stream Type ID
			switch stm.StreamType {
			case stream_type_session:
				slog.Debug("session stream was opened")

				var scm message.SessionClientMessage
				_, err := scm.Decode(stream)
				if err != nil {
					slog.Error("failed to get a SESSION_CLIENT message",
						"error", err,
						"stream_id", stream.StreamID(),
					)

					stream.CancelRead(ErrInternalError.StreamErrorCode())
					stream.CancelWrite(ErrInternalError.StreamErrorCode())
					return
				}

				ss := newSessionStream(stream, 0, &Parameters{scm.Parameters}, nil)

				// Enqueue the session stream
				sess.sessionStream = ss
				// sess.sessionStreamCh <- struct{}{}

				// Close the channel
				close(sess.sessionStreamCh)
			case stream_type_announce:
				// Handle the announce stream
				slog.Debug("announce stream was opened")

				// Decode the ANNOUNCE_PLEASE message
				var apm message.AnnouncePleaseMessage
				_, err := apm.Decode(stream)
				if err != nil {
					slog.Error("failed to get an Interest", "error", err)
					stream.CancelRead(ErrInternalError.StreamErrorCode())
					stream.CancelWrite(ErrInternalError.StreamErrorCode())
					return
				}

				// Create a sendAnnounceStream
				config := &AnnounceConfig{
					TrackPattern: string(apm.TrackPattern),
				}
				sas := newSendAnnounceStream(stream, config)

				// Enqueue the stream
				sess.sendAnnounceStreamQueue.Enqueue(sas)
			case stream_type_subscribe:
				slog.Debug("subscribe stream was opened")

				var sm message.SubscribeMessage
				_, err := sm.Decode(stream)
				if err != nil {
					slog.Debug("failed to read a SUBSCRIBE message", "error", err)
					stream.CancelRead(ErrInternalError.StreamErrorCode())
					stream.CancelWrite(ErrInternalError.StreamErrorCode())
					return
				}

				// Create a receiveSubscribeStream
				id := SubscribeID(sm.SubscribeID)
				path := TrackPath(sm.TrackPath)
				config := SubscribeConfig{
					GroupOrder:       GroupOrder(sm.GroupOrder),
					TrackPriority:    TrackPriority(sm.TrackPriority),
					MinGroupSequence: GroupSequence(sm.MinGroupSequence),
					MaxGroupSequence: GroupSequence(sm.MaxGroupSequence),
				}
				rss := newReceiveSubscribeStream(id, path, config, stream)

				// Enqueue the stream
				sess.receiveSubscribeStreamQueue.Enqueue(rss)
			case stream_type_info:
				slog.Debug("info stream was opened")

				// Get a received info-request
				var imr message.InfoRequestMessage
				_, err := imr.Decode(stream)
				if err != nil {
					slog.Error("failed to get a info-request", "error", err)
					return
				}

				sis := newSendInfoStream(stream, TrackPath(imr.TrackPath))

				// Enqueue the stream
				sess.sendInfoStreamQueue.Enqueue(sis)
			default:
				slog.Error("An unknown type of stream was opened")

				// Terminate the session
				sess.Terminate(ErrProtocolViolation)

				return
			}
		}(stream)
	}
}

func (sess *session) listenUniStreams(ctx context.Context) {
	for {
		/*
		 * Accept a unidirectional stream
		 */
		stream, err := sess.conn.AcceptUniStream(ctx)
		if err != nil {
			slog.Error("failed to accept a unidirectional stream", "error", err)
			return
		}

		slog.Debug("some data stream was opened")

		// Handle the stream
		go func(stream quic.ReceiveStream) {
			/*
			 * Get a Stream Type ID
			 */
			var stm message.StreamTypeMessage
			_, err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to get a Stream Type ID", "error", err)
				return
			}

			// Handle the stream by the Stream Type ID
			switch stm.StreamType {
			case stream_type_group:
				slog.Debug("group stream was opened")

				var gm message.GroupMessage
				_, err := gm.Decode(stream)
				if err != nil {
					slog.Error("failed to get a group", "error", err)
					return
				}

				id := SubscribeID(gm.SubscribeID)
				sequence := GroupSequence(gm.GroupSequence)
				rgs := newReceiveGroupStream(id, sequence, stream)

				queue, ok := sess.receiveGroupStreamQueues[id]
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

func handleSubscribeStream(ctx context.Context, sess *session, TrackHandler TrackHandler) {
	for {
		ss, err := sess.acceptSubscribeStream(ctx)
		if err != nil {
			return
		}

		if ss == nil {
			return
		}

		sts := newSendTrackStream(sess, ss)

		var info Info

		path := sts.TrackPath()
		info, err = TrackHandler.GetInfo(path)
		if err != nil {
			slog.Error("failed to get track info",
				"track_path", path,
				"error", err,
			)
			sts.CloseWithError(err)
			return
		}

		im := message.InfoMessage{
			TrackPriority:       message.TrackPriority(info.TrackPriority),
			LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
			GroupOrder:          message.GroupOrder(info.GroupOrder),
		}

		_, err = im.Encode(sts.subscribeStream.stream)
		if err != nil {
			return
		}

		go TrackHandler.ServeTrack(sts, sts.SubuscribeConfig())
	}
}

func handleInfoStream(ctx context.Context, sess *session, TrackHandler TrackHandler) {
	var info Info

	for {
		irs, err := sess.acceptInfoStream(ctx)
		if err != nil {
			slog.Error("failed to accept info stream",
				"error", err,
			)

			return
		}

		info, err = TrackHandler.GetInfo(irs.path)
		if err != nil {
			slog.Error("failed to get track info",
				"track_path", irs.path,
				"error", err,
			)
			irs.CloseWithError(err)
		}

		im := message.InfoMessage{
			TrackPriority:       message.TrackPriority(info.TrackPriority),
			LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
			GroupOrder:          message.GroupOrder(info.GroupOrder),
		}

		_, err = im.Encode(irs.stream)
		if err != nil {
			slog.Error("failed to send track info",
				"info", info,
				"error", err,
			)
			irs.CloseWithError(err)
		}
	}
}

func handleAnnounceStream(ctx context.Context, sess *session, AnnouncementHandler AnnouncementHandler) {
	for {
		annstr, err := sess.acceptAnnounceStream(ctx)
		if err != nil {
			return
		}

		go AnnouncementHandler.ServeAnnouncements(annstr, annstr.config)
	}
}

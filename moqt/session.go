package moqt

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSession(sessCtx *sessionContext, stream *sessionStream, conn quic.Connection, mux *TrackMux) *Session {
	if mux == nil {
		mux = DefaultMux
	}

	sess := &Session{
		ctx:                      sessCtx,
		conn:                     conn,
		mux:                      mux,
		sessionStream:            stream,
		receiveGroupStreamQueues: make(map[SubscribeID]*incomingGroupStreamQueue),
		sendGroupStreamQueues:    make(map[SubscribeID]*outgoingGroupStreamQueue),
		bitrate:                  0,
	}

	sess.wg.Add(2)
	// Listen bidirectional streams
	go func() {
		defer sess.wg.Done()
		sess.handleBiStreams(sessCtx)
	}()

	// Listen unidirectional streams
	go func() {
		defer sess.wg.Done()
		sess.handleUniStreams(sessCtx)
	}()

	return sess
}

type Session struct {
	conn quic.Connection

	mux *TrackMux // TODO

	subscribeIDCounter atomic.Uint64

	bitrate uint64 // TODO: use this when updating a session

	// sessionStreamch is the channel for signaling session streams
	// sessionStreamch chan struct{}

	// sessionStream is the session stream for the session
	sessionStream *sessionStream

	receiveGroupStreamQueues map[SubscribeID]*incomingGroupStreamQueue
	receiveGroupMapLocker    sync.RWMutex

	sendGroupStreamQueues map[SubscribeID]*outgoingGroupStreamQueue
	sendGroupMapLocker    sync.RWMutex

	ctx *sessionContext

	wg *sync.WaitGroup
}

func (s *Session) Terminate(reason error) {
	var trmerr TerminateError
	if reason == nil {
		trmerr = NoErrTerminate
	} else {
		if !errors.As(reason, &trmerr) {
			trmerr = ErrInternalError.WithReason(reason.Error())
		}
	}

	s.ctx.cancel(trmerr)

	code := quic.ConnectionErrorCode(trmerr.TerminateErrorCode())

	err := s.conn.CloseWithError(code, trmerr.Error())
	if err != nil {
		if logger := s.ctx.Logger(); logger != nil {
			logger.Error("failed to close the Connection",
				"error", err,
			)
		}
	}

	// Wait for finishing handling streams
	s.wg.Wait()

	if logger := s.ctx.Logger(); logger != nil {
		logger.Debug("terminated a session",
			"reason", trmerr,
		)
	}

}

func (s *Session) OpenAnnounceStream(prefix string) (AnnouncementReader, error) {
	if !strings.HasPrefix(prefix, "/") {
		panic("prefix must start with '/'")
	}

	return s.openAnnounceStream(prefix)
}

func (s *Session) OpenTrackStream(path BroadcastPath, name TrackName, config *SubscribeConfig) (*Subscriber, error) {
	if config == nil {
		config = &SubscribeConfig{}
	}
	id := s.nextSubscribeID()

	if logger := s.ctx.Logger(); logger != nil {
		logger.Debug("opening track stream",
			"subscribe_config", config.String(),
			"broadcast_path", path,
			"track_name", name,
			"subscribe_id", id)
	}

	trackCtx := newTrackContext(s.ctx, id, path, name)

	substr, err := s.openSubscribeStream(trackCtx, config)
	if err != nil {
		return nil, err
	}

	// Create a receive group stream queue
	queue := newIncomingGroupStreamQueue(substr.SubuscribeConfig)

	//
	s.receiveGroupMapLocker.Lock()
	s.receiveGroupStreamQueues[id] = queue
	s.receiveGroupMapLocker.Unlock()

	return &Subscriber{
		BroadcastPath:   path,
		TrackName:       name,
		SubscribeStream: substr,
		TrackReader:     newTrackReceiver(trackCtx, queue),
	}, nil
}

func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) nextSubscribeID() SubscribeID {
	// Increment and return the previous value atomically
	id := s.subscribeIDCounter.Add(1)
	return SubscribeID(id)
}

// TODO: Implement this method and use it
func (sess *Session) updateSession(bitrate uint64) error {
	if logger := sess.ctx.Logger(); logger != nil {
		logger.Debug("updating a session", "bitrate", bitrate)
	}

	// Send a SESSION_UPDATE message
	err := sess.sessionStream.UpdateSession(bitrate)
	if err != nil {
		if logger := sess.ctx.Logger(); logger != nil {
			logger.Error("failed to update a session",
				"error", err,
			)
		}
		return err
	}

	// Update the bitrate
	sess.bitrate = bitrate

	return nil
}

func (s *Session) openAnnounceStream(prefix string) (*receiveAnnounceStream, error) {
	apm := message.AnnouncePleaseMessage{
		TrackPrefix: prefix,
	}

	if logger := s.ctx.Logger(); logger != nil {
		logger.Debug("opening an announce stream", "config", apm)
	}

	// Open an Announce Stream
	stream, err := openStream(s.conn, stream_type_announce)
	if err != nil {
		if logger := s.ctx.Logger(); logger != nil {
			logger.Error("failed to open an Announce Stream", "error", err)
		}
		return nil, err
	}

	_, err = apm.Encode(stream)
	if err != nil {
		if logger := s.ctx.Logger(); logger != nil {
			logger.Error("failed to write an Interest message", "error", err)
		}
		return nil, err
	}

	return newReceiveAnnounceStream(stream, prefix), nil
}

func (s *Session) openSubscribeStream(trackCtx *trackContext, config *SubscribeConfig) (*sendSubscribeStream, error) {
	// Open a Subscribe Stream
	stream, err := openStream(s.conn, stream_type_subscribe)
	if err != nil {
		if logger := s.ctx.Logger(); logger != nil {
			logger.Error("failed to open a Subscribe Stream", "error", err)
		}
		return nil, err
	}

	// Send a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(trackCtx.id),
		BroadcastPath:    string(trackCtx.path),
		TrackName:        string(trackCtx.name),
		TrackPriority:    message.TrackPriority(config.TrackPriority),
		MinGroupSequence: message.GroupSequence(config.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(config.MaxGroupSequence),
	}
	_, err = sm.Encode(stream)
	if err != nil {
		if logger := s.ctx.Logger(); logger != nil {
			logger.Error("failed to send a SUBSCRIBE message", "error", err)
		}
		return nil, err
	}

	// Receive an INFO message
	var subok message.SubscribeOkMessage
	_, err = subok.Decode(stream)
	if err != nil {
		if logger := s.ctx.Logger(); logger != nil {
			logger.Error("failed to get a Info", "error", err)
		}
		return nil, err
	}

	substr := newSendSubscribeStream(trackCtx, config, stream)

	return substr, nil
}

func (sess *Session) goAway(uri string) {
	// TODO
}

// listenBiStreams accepts bidirectional streams and handles them based on their type.
// It listens for incoming streams and processes them in separate goroutines.
// The function handles session streams, announce streams, subscribe streams, and info streams.
// It also handles errors and terminates the session if an unknown stream type is encountered.
func (sess *Session) handleBiStreams(ctx context.Context) {
	logger := sess.ctx.Logger()
	if logger != nil {
		logger.Debug("listening for bidirectional streams")
	}

	for { // Accept a bidirectional stream
		stream, err := sess.conn.AcceptStream(ctx)
		if err != nil {
			if logger != nil {
				logger.Error("failed to accept a bidirectional stream", "error", err)
			}
			return
		}

		if logger != nil {
			logger.Debug("A some stream was opened")
		}

		// Handle the stream
		go sess.processBiStream(stream)
	}
}

func (sess *Session) processBiStream(stream quic.Stream) {
	logger := sess.ctx.Logger()
	// Decode the STREAM_TYPE message and get the stream type ID
	var stm message.StreamTypeMessage
	_, err := stm.Decode(stream)
	if err != nil {
		if logger != nil {
			logger.Error("failed to get a Stream Type ID",
				"error", err,
				"stream_id", stream.StreamID(),
			)
		}

		return
	}
	// Handle the stream by the Stream Type ID
	switch stm.StreamType {
	case stream_type_announce:
		// Handle the announce stream
		if logger != nil {
			logger.Debug("announce stream was opened")
		}

		// Decode the ANNOUNCE_PLEASE message
		var apm message.AnnouncePleaseMessage
		_, err := apm.Decode(stream)
		if err != nil {
			if logger != nil {
				logger.Error("failed to get an Interest", "error", err)
			}
			stream.CancelRead(ErrInternalError.StreamErrorCode())
			stream.CancelWrite(ErrInternalError.StreamErrorCode())
			return
		}

		prefix := apm.TrackPrefix

		annstr := newSendAnnounceStream(stream, prefix)

		sess.mux.ServeAnnouncements(annstr, prefix)
	case stream_type_subscribe:
		if logger != nil {
			logger.Debug("subscribe stream was opened")
		}

		var sm message.SubscribeMessage
		_, err := sm.Decode(stream)
		if err != nil {
			if logger != nil {
				logger.Debug("failed to read a SUBSCRIBE message", "error", err)
			}
			stream.CancelRead(ErrInternalError.StreamErrorCode())
			stream.CancelWrite(ErrInternalError.StreamErrorCode())
			return
		}

		// Create a receiveSubscribeStream
		id := SubscribeID(sm.SubscribeID)
		path := BroadcastPath(sm.BroadcastPath)
		name := TrackName(sm.TrackName)
		config := &SubscribeConfig{
			TrackPriority:    TrackPriority(sm.TrackPriority),
			MinGroupSequence: GroupSequence(sm.MinGroupSequence),
			MaxGroupSequence: GroupSequence(sm.MaxGroupSequence),
		}

		ann, handler := sess.mux.findHandler(path)

		trackCtx := newTrackContext(sess.ctx, id, path, name)

		substr := newReceiveSubscribeStream(trackCtx, stream, config)

		if !ann.IsActive() {
			substr.closeWithError(ErrTrackDoesNotExist)
			return
		}

		_, err = message.SubscribeOkMessage{
			GroupOrder: message.GroupOrder(ann.info.GroupOrder),
		}.Encode(stream)
		if err != nil {
			return
		}
		openStreamFunc := func(groupCtx *groupContext) (*sendGroupStream, error) {
			grpstr, err := openStream(sess.conn, stream_type_group)
			if err != nil {
				if logger := sess.ctx.Logger(); logger != nil {
					logger.Error("failed to open a Group Stream", "error", err)
				}
				return nil, err
			}

			_, err = message.GroupMessage{
				SubscribeID:   sm.SubscribeID,
				GroupSequence: message.GroupSequence(groupCtx.seq),
			}.Encode(grpstr)
			if err != nil {
				return nil, err
			}

			return newSendGroupStream(stream, groupCtx), nil
		}

		queue := newOutgoingGroupStreamQueue()
		sess.sendGroupMapLocker.Lock()
		sess.sendGroupStreamQueues[id] = queue
		sess.sendGroupMapLocker.Unlock()

		pub := &Publisher{
			BroadcastPath:   path,
			TrackName:       name,
			SubscribeStream: substr,
			TrackWriter:     newTrackSender(trackCtx, queue, openStreamFunc),
		}

		go handler.ServeTrack(pub)
	default:
		if logger != nil {
			logger.Error("An unknown type of stream was opened")
		}

		// Terminate the session
		sess.Terminate(ErrProtocolViolation)

		return
	}
}

func (sess *Session) handleUniStreams(ctx context.Context) {
	for { /*
		 * Accept a unidirectional stream
		 */
		stream, err := sess.conn.AcceptUniStream(ctx)
		if err != nil {
			if logger := sess.ctx.Logger(); logger != nil {
				logger.Error("failed to accept a unidirectional stream", "error", err)
			}
			return
		}

		if logger := sess.ctx.Logger(); logger != nil {
			logger.Debug("some data stream was opened")
		}

		// Handle the stream
		go sess.processUniStream(stream)
	}
}

func (sess *Session) processUniStream(stream quic.ReceiveStream) { /*
	 * Get a Stream Type ID
	 */
	var stm message.StreamTypeMessage
	_, err := stm.Decode(stream)
	if err != nil {
		if logger := sess.ctx.Logger(); logger != nil {
			logger.Error("failed to get a Stream Type ID", "error", err)
		}
		return
	}

	// Handle the stream by the Stream Type ID
	switch stm.StreamType {
	case stream_type_group:
		if logger := sess.ctx.Logger(); logger != nil {
			logger.Debug("group stream was opened")
		}

		var gm message.GroupMessage
		_, err := gm.Decode(stream)
		if err != nil {
			if logger := sess.ctx.Logger(); logger != nil {
				logger.Error("failed to get a group", "error", err)
			}
			return
		}

		id := SubscribeID(gm.SubscribeID)
		sequence := GroupSequence(gm.GroupSequence)
		rgs := newReceiveGroupStream(id, sequence, stream)
		_, ok := sess.receiveGroupStreamQueues[id]
		if !ok {
			if logger := sess.ctx.Logger(); logger != nil {
				logger.Error("failed to get a data receive stream queue", "error", "queue not found")
			}
			stream.CancelRead(ErrInternalError.StreamErrorCode())
			return
		}

		// Enqueue the receiver
		sess.receiveGroupStreamQueues[id].enqueue(rgs)

	default:
		if logger := sess.ctx.Logger(); logger != nil {
			logger.Debug("An unknown type of stream was opened")
		}

		// Terminate the session
		sess.Terminate(ErrProtocolViolation)

		return
	}
}

// func (sess *Session) handleSubscribeStreams() {
// 	ctx := sess.Context()
// 	var path BroadcastPath

// 	var ann *Announcement
// 	var handler TrackHandler

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 		}

// 		sub, err := sess.acceptSubscription(ctx)
// 		if err != nil {
// 			return
// 		}

// 	}
// }

package moqt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSession(conn quic.Connection, version protocol.Version, path string, clientParams, serverParams *Parameters,
	stream quic.Stream, mux *TrackMux, logger *slog.Logger) *Session {
	if mux == nil {
		mux = DefaultMux
	}

	if logger == nil {
		logger = slog.Default()
	}

	sess := &Session{
		sessionStream:    newSessionStream(conn.Context(), stream),
		path:             path,
		version:          version,
		clientParameters: clientParams,
		serverParameters: serverParams,
		logger:           logger,
		conn:             conn,
		mux:              mux,
		trackReceivers:   make(map[SubscribeID]*trackReceiver),
		trackSenders:     make(map[SubscribeID]*trackSender),
	}

	sess.wg.Add(3) // Listen for session stream closure
	go func() {
		defer sess.wg.Done()
		<-sess.Context().Done()
		if !sess.terminating() {
			logger.Warn("session stream closed unexpectedly",
				"reason", context.Cause(sess.Context()),
			)
			sess.Terminate(ProtocolViolationErrorCode, "session stream closed unexpectedly")
		}
	}()

	// Listen bidirectional streams
	go func() {
		defer sess.wg.Done()
		logger.Debug("starting bidirectional stream handler")
		sess.handleBiStreams()
		logger.Debug("bidirectional stream handler terminated")
	}()

	// Listen unidirectional streams
	go func() {
		defer sess.wg.Done()
		logger.Debug("starting unidirectional stream handler")
		sess.handleUniStreams()
		logger.Debug("unidirectional stream handler terminated")
	}()

	return sess
}

type Session struct {
	*sessionStream

	wg sync.WaitGroup // WaitGroup for session cleanup

	path string

	// Version of the protocol used in this session
	version protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters

	logger *slog.Logger

	conn quic.Connection
	// connMu sync.Mutex

	mux *TrackMux // TODO

	subscribeIDCounter atomic.Uint64

	trackReceivers        map[SubscribeID]*trackReceiver
	receiveGroupMapLocker sync.RWMutex

	trackSenders       map[SubscribeID]*trackSender
	sendGroupMapLocker sync.RWMutex

	isTerminating atomic.Bool
	sessErr       error
}

func (s *Session) terminating() bool {
	return s.isTerminating.Load()
}

func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) Terminate(code SessionErrorCode, msg string) error {
	if s.terminating() {
		s.logger.Debug("termination already in progress")
		return s.sessErr
	}

	s.isTerminating.Store(true)

	s.logger.Info("terminating session",
		"code", code,
		"message", msg,
	)

	// s.connMu.Lock()
	// defer s.connMu.Unlock()
	err := s.conn.CloseWithError(quic.ConnectionErrorCode(code), msg)
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			reason := &SessionError{
				ApplicationError: appErr,
			}
			s.sessErr = reason
			s.logger.Error("session already terminated",
				"error", reason,
			)
			return reason
		}
		s.sessErr = err
		s.logger.Error("session termination failed",
			"error", err,
		)
		return err
	}

	s.sessErr = &SessionError{
		ApplicationError: &quic.ApplicationError{
			ErrorCode:    quic.ApplicationErrorCode(code),
			ErrorMessage: msg,
		},
	}
	s.cancel(s.sessErr)

	// Wait for finishing handling streams
	s.wg.Wait()

	s.logger.Info("session terminated successfully")

	return nil
}

func (s *Session) OpenTrackStream(path BroadcastPath, name TrackName, config *SubscribeConfig) (*Subscription, error) {
	if s.terminating() {
		return nil, s.sessErr
	}

	if config == nil {
		config = &SubscribeConfig{}
	}

	id := s.nextSubscribeID()

	stream, err := s.conn.OpenStream()
	if err != nil {
		s.logger.Error("failed to open bidirectional stream",
			"error", err,
		)
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			return nil, &SessionError{
				ApplicationError: appErr,
			}
		}
		return nil, err
	}

	streamLogger := s.logger.With("stream_id", stream.StreamID())

	stm := message.StreamTypeMessage{
		StreamType: stream_type_subscribe,
	}
	err = stm.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) && strErr.Remote {
			stream.CancelRead(strErr.ErrorCode)
			streamLogger.Error("failed to encode stream type message",
				"error", strErr,
			)
			return nil, &SubscribeError{
				StreamError: strErr,
			}
		}
		strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)
		streamLogger.Error("failed to encode stream type message",
			"error", err,
		)
		return nil, err
	}

	// Send a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(id),
		BroadcastPath:    string(path),
		TrackName:        string(name),
		TrackPriority:    message.TrackPriority(config.TrackPriority),
		MinGroupSequence: message.GroupSequence(config.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(config.MaxGroupSequence),
	}
	err = sm.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) && strErr.Remote {
			stream.CancelRead(strErr.ErrorCode)
			streamLogger.Error("failed to encode SUBSCRIBE message",
				"error", strErr,
			)
			return nil, &SubscribeError{
				StreamError: strErr,
			}
		}

		strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)

		streamLogger.Error("failed to encode SUBSCRIBE message",
			"error", err,
		)
		return nil, err
	}

	var subok message.SubscribeOkMessage
	err = subok.Decode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			strErrCode := quic.StreamErrorCode(strErr.ErrorCode)
			stream.CancelWrite(strErrCode)

			streamLogger.Error("failed to read SUBSCRIBE_OK response",
				"error", strErr,
			)
			return nil, &SubscribeError{
				StreamError: strErr,
			}
		}
		strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)

		streamLogger.Error("failed to read SUBSCRIBE_OK response",
			"error", err,
		)
		return nil, err
	}

	substr := newSendSubscribeStream(s.ctx, id, stream, config)

	streamLogger.Debug("subscribe stream opened",
		"subscribe_id", id,
		"broadcast_path", path,
		"track_name", name,
		"subscribe_config", config,
	)
	// Create a receive group stream queue
	trackReceiver := newTrackReceiver(substr.ctx)
	s.receiveGroupMapLocker.Lock()
	s.trackReceivers[id] = trackReceiver
	s.receiveGroupMapLocker.Unlock()

	return &Subscription{
		BroadcastPath: path,
		TrackName:     name,
		TrackReader:   trackReceiver,
		Controller:    substr,
	}, nil
}

func (s *Session) nextSubscribeID() SubscribeID {
	// Increment and return the previous value atomically
	id := SubscribeID(s.subscribeIDCounter.Add(1))

	return id
}

func (sess *Session) OpenAnnounceStream(prefix string) (AnnouncementReader, error) {
	if sess.terminating() {
		return nil, sess.sessErr
	}

	// Create a logger with consistent context for this announcement

	stream, err := sess.conn.OpenStream()
	if err != nil {
		sess.logger.Error("failed to open stream for announce",
			"error", err,
		)
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {

			return nil, &SessionError{
				ApplicationError: appErr,
			}
		}

		return nil, err
	}

	// Create a stream-specific logger
	streamLogger := sess.logger.With("stream_id", stream.StreamID())
	streamLogger.Debug("opened bidirectional stream")

	st := message.StreamTypeMessage{
		StreamType: stream_type_announce,
	}
	err = st.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			strErrCode := quic.StreamErrorCode(InternalAnnounceErrorCode)
			stream.CancelRead(strErrCode)

			streamLogger.Error("failed to encode stream type message",
				"error", strErr,
			)
			return nil, &AnnounceError{
				StreamError: strErr,
			}
		}

		streamLogger.Error("failed to encode stream type message",
			"error", err,
		)
		return nil, err
	}

	apm := message.AnnouncePleaseMessage{
		TrackPrefix: prefix,
	}
	err = apm.Encode(stream)
	if err != nil {
		streamLogger.Error("failed to send ANNOUNCE_PLEASE message",
			"error", err,
		)
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			strErrCode := quic.StreamErrorCode(InternalAnnounceErrorCode)
			stream.CancelRead(strErrCode)

			return nil, &AnnounceError{
				StreamError: strErr,
			}
		}

		strErrCode := quic.StreamErrorCode(InternalAnnounceErrorCode)
		stream.CancelWrite(strErrCode)
		stream.CancelRead(strErrCode)

		return nil, err
	}

	streamLogger.Debug("announce stream opened successfully",
		"track_prefix", prefix,
	)

	return newReceiveAnnounceStream(sess.ctx, stream, prefix), nil
}

func (sess *Session) goAway(uri string) {
	// TODO
}

// listenBiStreams accepts bidirectional streams and handles them based on their type.
// It listens for incoming streams and processes them in separate goroutines.
// The function handles session streams, announce streams, subscribe streams, and info streams.
// It also handles errors and terminates the session if an unknown stream type is encountered.
func (sess *Session) handleBiStreams() {
	for { // Accept a bidirectional stream
		stream, err := sess.conn.AcceptStream(sess.ctx)
		if err != nil {
			sess.logger.Debug("failed to accept bidirectional stream",
				"error", err,
			)
			return
		}

		streamLogger := sess.logger.With("stream_id", stream.StreamID())
		streamLogger.Debug("accepted bidirectional stream")

		// Handle the stream
		go sess.processBiStream(stream, streamLogger)
	}
}

func (sess *Session) processBiStream(stream quic.Stream, streamLogger *slog.Logger) {
	var stm message.StreamTypeMessage
	err := stm.Decode(stream)
	if err != nil {
		streamLogger.Error("failed to decode stream type message",
			"error", err,
		)
		sess.Terminate(ProtocolViolationErrorCode, err.Error())
		return
	}

	switch stm.StreamType {
	case stream_type_announce:
		var apm message.AnnouncePleaseMessage
		err := apm.Decode(stream)
		if err != nil {
			streamLogger.Error("failed to decode ANNOUNCE_PLEASE message",
				"error", err,
			)
			sess.Terminate(ProtocolViolationErrorCode, err.Error())
			return
		}

		prefix := apm.TrackPrefix

		annstr := newSendAnnounceStream(stream, prefix)

		streamLogger.Debug("accepted announce stream")

		sess.mux.ServeAnnouncements(annstr, prefix)
	case stream_type_subscribe:
		var sm message.SubscribeMessage
		err := sm.Decode(stream)
		if err != nil {
			streamLogger.Error("failed to decode SUBSCRIBE message",
				"error", err,
			)
			sess.Terminate(InternalSessionErrorCode, err.Error())
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
		// Create a subscription-specific logger
		subLogger := streamLogger.With(
			"subscribe_id", id,
			"broadcast_path", path,
			"track_name", name,
			"config", config.String(),
		)

		handler := sess.mux.Handler(path)
		if handler == nil {
			subLogger.Warn("track not found for subscription")

			strErrCode := quic.StreamErrorCode(TrackNotFoundErrorCode)
			stream.CancelWrite(strErrCode)
			stream.CancelRead(strErrCode)
			return
		}

		substr := newReceiveSubscribeStream(sess.ctx, id, stream, config)

		subLogger.Debug("accepted a subscribe stream")

		openGroupStreamFunc := func(trackCtx context.Context, seq GroupSequence) (*sendGroupStream, error) {

			stream, err := sess.conn.OpenUniStream()
			if err != nil {
				sess.logger.Error("failed to open an unidirectional stream",
					"error", err,
				)

				var appErr *quic.ApplicationError
				if errors.As(err, &appErr) {
					sessErr := &SessionError{
						ApplicationError: appErr,
					}
					return nil, sessErr
				}

				return nil, err
			}

			// Add stream_id to the logger context
			streamLogger := subLogger.With(
				"stream_id", stream.StreamID(),
				"group_sequence", seq,
			)
			streamLogger.Debug("opened a group stream")

			stm := message.StreamTypeMessage{
				StreamType: stream_type_group,
			}
			err = stm.Encode(stream)
			if err != nil {
				streamLogger.Error("failed to send stream type message",
					"error", err,
				)

				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					return nil, &GroupError{StreamError: strErr}
				}

				strErrCode := quic.StreamErrorCode(InternalGroupErrorCode)
				stream.CancelWrite(strErrCode)

				return nil, GroupError{
					StreamError: &quic.StreamError{
						StreamID:  stream.StreamID(),
						ErrorCode: strErrCode,
					},
				}
			}

			gm := message.GroupMessage{
				SubscribeID:   sm.SubscribeID,
				GroupSequence: message.GroupSequence(seq),
			}
			err = gm.Encode(stream)
			if err != nil {
				streamLogger.Error("failed to send group message",
					"error", err,
				)

				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					streamLogger.Error("group message encoding failed",
						"error", strErr,
					)
					return nil, &GroupError{StreamError: strErr}
				}

				strErrCode := quic.StreamErrorCode(InternalGroupErrorCode)
				stream.CancelWrite(strErrCode)

				return nil, GroupError{
					StreamError: &quic.StreamError{
						StreamID:  stream.StreamID(),
						ErrorCode: strErrCode,
					},
				}
			}

			return newSendGroupStream(trackCtx, stream, seq), nil
		}

		trackSender := newTrackSender(substr.ctx, openGroupStreamFunc, substr.accept)
		trackSender.acceptFunc = func(info Info) {
			substr.WriteInfo(info)
		}
		sess.sendGroupMapLocker.Lock()
		sess.trackSenders[id] = trackSender
		sess.sendGroupMapLocker.Unlock()

		subLogger.Info("serving track for subscription")

		handler.ServeTrack(&Publication{
			BroadcastPath: path,
			TrackName:     name,
			Controller:    substr,
			TrackWriter:   trackSender,
		})
	default:
		streamLogger.Error("unknown bidirectional stream type",
			"stream_type", stm.StreamType,
		)
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown bidirectional stream type: %v", stm.StreamType))
		return
	}
}

func (sess *Session) handleUniStreams() {
	for {
		/*
		 * Accept a unidirectional stream
		 */
		stream, err := sess.conn.AcceptUniStream(sess.ctx)
		if err != nil {
			sess.logger.Debug("failed to accept unidirectional stream, handler stopping",
				"error", err,
			)
			return
		}

		streamLogger := sess.logger.With("stream_id", stream.StreamID())
		streamLogger.Debug("accepted unidirectional stream")

		// Handle the stream
		go sess.processUniStream(stream, streamLogger)
	}
}

func (sess *Session) processUniStream(stream quic.ReceiveStream, streamLogger *slog.Logger) {
	/*
	 * Get a Stream Type ID
	 */
	var stm message.StreamTypeMessage
	err := stm.Decode(stream)
	if err != nil {
		streamLogger.Error("failed to decode stream type message",
			"error", err,
		)
		return
	}

	// Handle the stream by the Stream Type ID
	switch stm.StreamType {
	case stream_type_group:
		var gm message.GroupMessage
		err := gm.Decode(stream)
		if err != nil {
			streamLogger.Error("failed to decode group message",
				"error", err,
			)
			return
		}

		id := SubscribeID(gm.SubscribeID)

		// Create a group-specific logger
		groupLogger := streamLogger.With(
			"subscribe_id", id,
			"group_sequence", gm.GroupSequence,
		)

		receiver, ok := sess.trackReceivers[id]
		if !ok {
			groupLogger.Warn("received group for unknown subscription")
			stream.CancelRead(quic.StreamErrorCode(InvalidSubscribeIDErrorCode))
			return
		}

		groupLogger.Debug("accepted group stream")

		// Enqueue the receiver
		receiver.enqueueGroup(GroupSequence(gm.GroupSequence), stream)
	default:
		streamLogger.Error("unknown unidirectional stream type received",
			"stream_type", stm.StreamType,
		)

		// Terminate the session
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown unidirectional stream type: %v", stm.StreamType))
		return
	}
}

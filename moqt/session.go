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

	// bitrate atomic.Uint64 // Bitrate in bits per second

	logger *slog.Logger

	conn   quic.Connection
	connMu sync.Mutex

	mux *TrackMux // TODO

	subscribeIDCounter atomic.Uint64

	trackReceivers        map[SubscribeID]*trackReceiver
	receiveGroupMapLocker sync.RWMutex

	trackSenders       map[SubscribeID]*trackSender
	sendGroupMapLocker sync.RWMutex

	isTerminating atomic.Bool
	termErr       error
}

func (s *Session) terminating() bool {
	return s.isTerminating.Load()
}

func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) Terminate(code SessionErrorCode, msg string) error {
	if s.terminating() {
		s.logger.Debug("termination already in progress",
			"code", code,
			"message", msg,
		)
		return s.termErr
	}

	s.isTerminating.Store(true)

	s.logger.Info("terminating session",
		"code", code,
		"message", msg,
		"remote_address", s.conn.RemoteAddr(),
	)

	s.connMu.Lock()
	defer s.connMu.Unlock()
	err := s.conn.CloseWithError(quic.ConnectionErrorCode(code), msg)
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			reason := &SessionError{
				ApplicationError: appErr,
			}
			s.termErr = reason
			s.logger.Error("session termination with application error",
				"error", reason,
			)
			return reason
		}
		s.termErr = err
		s.logger.Error("session termination failed",
			"error", err,
		)
		return err
	}

	s.termErr = &SessionError{
		ApplicationError: &quic.ApplicationError{
			ErrorCode:    quic.ApplicationErrorCode(code),
			ErrorMessage: msg,
		},
	}
	s.cancel(s.termErr)

	// Wait for finishing handling streams
	s.logger.Debug("waiting for stream handlers to complete")
	s.wg.Wait()

	s.logger.Info("session terminated successfully")

	return nil
}

func (s *Session) OpenTrackStream(path BroadcastPath, name TrackName, config *SubscribeConfig) (*Subscriber, error) {
	if s.terminating() {
		return nil, s.termErr
	}

	if config == nil {
		config = &SubscribeConfig{}
	}

	id := s.nextSubscribeID()

	s.connMu.Lock()
	stream, err := s.conn.OpenStream()
	s.connMu.Unlock()
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
	_, err = stm.Encode(stream)
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
	_, err = sm.Encode(stream)
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
	_, err = subok.Decode(stream)
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
	trackReceiver := newTrackReceiver(substr)
	s.receiveGroupMapLocker.Lock()
	s.trackReceivers[id] = trackReceiver
	s.receiveGroupMapLocker.Unlock()

	return &Subscriber{
		BroadcastPath:   path,
		TrackName:       name,
		SubscribeStream: substr,
		TrackReader:     trackReceiver,
	}, nil
}

func (s *Session) nextSubscribeID() SubscribeID {
	// Increment and return the previous value atomically
	id := SubscribeID(s.subscribeIDCounter.Add(1))

	return id
}

func (sess *Session) OpenAnnounceStream(prefix string) (AnnouncementReader, error) {
	if sess.terminating() {
		return nil, sess.termErr
	}

	// Create a logger with consistent context for this announcement

	sess.connMu.Lock()
	stream, err := sess.conn.OpenStream()
	if err != nil {
		sess.logger.Error("failed to open stream for announce",
			"error", err,
		)
		sess.connMu.Unlock()
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {

			return nil, &SessionError{
				ApplicationError: appErr,
			}
		}

		return nil, err
	}
	sess.connMu.Unlock()

	// Create a stream-specific logger
	streamLogger := sess.logger.With("stream_id", stream.StreamID())
	streamLogger.Debug("opened bidirectional stream")

	st := message.StreamTypeMessage{
		StreamType: stream_type_announce,
	}
	_, err = st.Encode(stream)
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
	_, err = apm.Encode(stream)
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
	_, err := stm.Decode(stream)
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
		_, err := apm.Decode(stream)
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
		_, err := sm.Decode(stream)
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

		substr := newReceiveSubscribeStream(id, stream, config)

		subLogger.Debug("accepted a subscribe stream")

		openGroupStreamFunc := func(seq GroupSequence) (*sendGroupStream, error) {
			// Create a group-specific logger
			groupLogger := subLogger.With("group_sequence", seq)

			stream, err := sess.conn.OpenUniStream()
			if err != nil {
				var appErr *quic.ApplicationError
				if errors.As(err, &appErr) {
					sessErr := &SessionError{
						ApplicationError: appErr,
					}
					groupLogger.Error("failed to open a group stream",
						"error", sessErr,
					)
					return nil, sessErr
				}

				groupLogger.Error("failed to open a group stream",
					"error", err,
				)
				return nil, err
			}

			// Add stream_id to the logger context
			streamLogger := groupLogger.With("stream_id", stream.StreamID())
			streamLogger.Debug("opened a group stream")

			stm := message.StreamTypeMessage{
				StreamType: stream_type_group,
			}
			_, err = stm.Encode(stream)
			if err != nil {
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					streamLogger.Error("group stream type message encoding failed",
						"error", strErr,
					)
					return nil, &GroupError{StreamError: strErr}
				}

				strErrCode := quic.StreamErrorCode(InternalGroupErrorCode)
				stream.CancelWrite(strErrCode)

				streamLogger.Error("group stream type message encoding failed",
					"error", err,
				)
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
			_, err = gm.Encode(stream)
			if err != nil {
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					streamLogger.Error("group message encoding failed",
						"error", strErr,
					)
					return nil, &GroupError{StreamError: strErr}
				}

				strErr = &quic.StreamError{
					StreamID:  stream.StreamID(),
					ErrorCode: quic.StreamErrorCode(InternalGroupErrorCode),
				}

				streamLogger.Error("group message encoding failed",
					"error", err,
				)
				return nil, GroupError{StreamError: strErr}
			}

			streamLogger.Debug("group stream setup completed")

			return newSendGroupStream(stream, seq), nil
		}

		trackSender := newTrackSender(substr, openGroupStreamFunc)
		sess.sendGroupMapLocker.Lock()
		sess.trackSenders[id] = trackSender
		sess.sendGroupMapLocker.Unlock()

		subLogger.Info("serving track for subscription")

		handler.ServeTrack(&Publisher{
			BroadcastPath:   path,
			TrackName:       name,
			SubscribeStream: substr,
			TrackWriter:     trackSender,
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
	_, err := stm.Decode(stream)
	if err != nil {
		streamLogger.Error("failed to decode stream type message",
			"error", err,
		)
		return
	}

	streamLogger.Debug("decoded stream type message",
		"stream_type", stm.StreamType,
	)

	// Handle the stream by the Stream Type ID
	switch stm.StreamType {
	case stream_type_group:
		var gm message.GroupMessage
		_, err := gm.Decode(stream)
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

		group := newReceiveGroupStream(GroupSequence(gm.GroupSequence), stream)

		groupLogger.Debug("accepted group stream")

		// Enqueue the receiver
		receiver.enqueueGroup(group)
	default:
		streamLogger.Error("unknown unidirectional stream type received",
			"stream_type", stm.StreamType,
		)

		// Terminate the session
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown unidirectional stream type: %v", stm.StreamType))
		return
	}
}

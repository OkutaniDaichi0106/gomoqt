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
		logger = slog.New(slog.DiscardHandler)
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
		trackReaders:     make(map[SubscribeID]*TrackReader),
		trackWriters:     make(map[SubscribeID]*TrackWriter),
	}

	sess.wg.Add(3)
	// Listen for session stream closure
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
		sess.handleBiStreams()
	}()

	// Listen unidirectional streams
	go func() {
		defer sess.wg.Done()
		sess.handleUniStreams()
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

	mux *TrackMux // TODO

	subscribeIDCounter atomic.Uint64

	trackReaders         map[SubscribeID]*TrackReader
	trackReaderMapLocker sync.RWMutex

	trackWriters         map[SubscribeID]*TrackWriter
	trackWriterMapLocker sync.RWMutex

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

	//
	s.sessionStream.close()

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

func (s *Session) OpenTrackStream(path BroadcastPath, name TrackName, config *TrackConfig) (*TrackReader, error) {
	if s.terminating() {
		return nil, s.sessErr
	}

	if config == nil {
		config = &TrackConfig{}
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

	err = message.StreamTypeSubscribe.Encode(stream)
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
		SubscribeID:      id,
		BroadcastPath:    string(path),
		TrackName:        string(name),
		TrackPriority:    config.TrackPriority,
		MinGroupSequence: config.MinGroupSequence,
		MaxGroupSequence: config.MaxGroupSequence,
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

	substr := newSendSubscribeStream(id, stream, config)

	streamLogger.Debug("subscribe stream opened",
		"subscribe_id", id,
		"broadcast_path", path,
		"track_name", name,
		"subscribe_config", config,
	)

	// Create a receive group stream queue
	trackReceiver := newTrackReader(path, name, substr, func() {
		s.removeTrackReader(id)
	})
	s.addTrackReader(id, trackReceiver)

	return trackReceiver, nil
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

	err = message.StreamTypeAnnounce.Encode(stream)
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
	var streamType message.StreamType
	err := streamType.Decode(stream)
	if err != nil {
		streamLogger.Error("failed to decode stream type message",
			"error", err,
		)
		sess.Terminate(ProtocolViolationErrorCode, err.Error())
		return
	}

	switch streamType {
	case message.StreamTypeAnnounce:
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
	case message.StreamTypeSubscribe:
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
		config := &TrackConfig{
			TrackPriority:    sm.TrackPriority,
			MinGroupSequence: sm.MinGroupSequence,
			MaxGroupSequence: sm.MaxGroupSequence,
		}
		// Create a subscription-specific logger
		subLogger := streamLogger.With(
			"subscribe_id", sm.SubscribeID,
			"broadcast_path", sm.BroadcastPath,
			"track_name", sm.TrackName,
			"config", config.String(),
		)

		handler := sess.mux.Handler(BroadcastPath(sm.BroadcastPath))
		if handler == nil {
			subLogger.Warn("track not found for subscription")

			strErrCode := quic.StreamErrorCode(TrackNotFoundErrorCode)
			stream.CancelWrite(strErrCode)
			stream.CancelRead(strErrCode)
			return
		}

		substr := newReceiveSubscribeStream(sm.SubscribeID, stream, config)

		subLogger.Debug("accepted a subscribe stream")

		trackWriter := newTrackWriter(BroadcastPath(sm.BroadcastPath), TrackName(sm.TrackName),
			substr, sess.conn.OpenUniStream, func() { sess.removeTrackWriter(sm.SubscribeID) })
		sess.addTrackWriter(sm.SubscribeID, trackWriter)

		subLogger.Info("serving track for subscription")

		handler.ServeTrack(trackWriter)
	default:
		streamLogger.Error("unknown bidirectional stream type",
			"stream_type", streamType,
		)
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown bidirectional stream type: %v", streamType))
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
	var streamType message.StreamType
	err := streamType.Decode(stream)
	if err != nil {
		streamLogger.Error("failed to decode stream type message",
			"error", err,
		)
		return
	}

	// Handle the stream by the Stream Type ID
	switch streamType {
	case message.StreamTypeGroup:
		var gm message.GroupMessage
		err := gm.Decode(stream)
		if err != nil {
			streamLogger.Error("failed to decode group message",
				"error", err,
			)
			return
		}

		// Create a group-specific logger
		groupLogger := streamLogger.With(
			"subscribe_id", gm.SubscribeID,
			"group_sequence", gm.GroupSequence,
		)

		track, ok := sess.trackReaders[gm.SubscribeID]
		if !ok {
			groupLogger.Warn("received group for unknown subscription")
			stream.CancelRead(quic.StreamErrorCode(InvalidSubscribeIDErrorCode))
			return
		}

		groupLogger.Debug("accepted group stream")

		// Enqueue the receiver
		track.enqueueGroup(gm.GroupSequence, stream)
	default:
		streamLogger.Error("unknown unidirectional stream type received",
			"stream_type", streamType,
		)

		// Terminate the session
		sess.Terminate(ProtocolViolationErrorCode, fmt.Sprintf("unknown unidirectional stream type: %v", streamType))
		return
	}
}

func (s *Session) addTrackWriter(id SubscribeID, writer *TrackWriter) {
	s.trackWriterMapLocker.Lock()
	defer s.trackWriterMapLocker.Unlock()
	s.trackWriters[id] = writer
}

func (s *Session) removeTrackWriter(id SubscribeID) {
	s.trackWriterMapLocker.Lock()
	defer s.trackWriterMapLocker.Unlock()

	if writer, ok := s.trackWriters[id]; ok {
		writer.Close()
		delete(s.trackWriters, id)
	}
}

func (s *Session) addTrackReader(id SubscribeID, reader *TrackReader) {
	s.trackReaderMapLocker.Lock()
	defer s.trackReaderMapLocker.Unlock()
	s.trackReaders[id] = reader
	s.logger.Debug("added track reader",
		"subscribe_id", id,
		"track_name", reader.TrackName,
	)
}

func (s *Session) removeTrackReader(id SubscribeID) {
	s.trackReaderMapLocker.Lock()
	defer s.trackReaderMapLocker.Unlock()

	if reader, ok := s.trackReaders[id]; ok {
		reader.Close()
		delete(s.trackReaders, id)
	}
}

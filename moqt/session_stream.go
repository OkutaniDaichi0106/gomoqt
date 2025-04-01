package moqt

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type SessionStream interface {
	UpdateSession(bitrate uint64) error
	ClientParameters() *Parameters
	ServerParameters() *Parameters
	Version() protocol.Version
	Close() error
	CloseWithError(err error) error
}

func newSessionStream(stream quic.Stream, selectedVersion protocol.Version, clientParameters, serverParameters *Parameters) *sessionStream {
	sess := &sessionStream{
		stream:           stream,
		selectedVersion:  selectedVersion,
		clientParameters: clientParameters,
		serverParameters: serverParameters,
	}

	// Start listening for updates in a separate goroutine
	go sess.listenUpdates()

	return sess
}

var _ SessionStream = (*sessionStream)(nil)

type sessionStream struct {
	stream quic.Stream
	mu     sync.Mutex

	// Versions selected by the server
	selectedVersion protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters

	closed   bool
	closeErr error
}

func (ss *sessionStream) UpdateSession(bitrate uint64) error {
	sum := message.SessionUpdateMessage{Bitrate: bitrate}
	_, err := sum.Encode(ss.stream)
	if err != nil {
		slog.Error("failed to send a SESSION_UPDATE message", "error", err)
		return err
	}

	slog.Debug("sent a SESSION_UPDATE message")

	return nil
}

func (ss *sessionStream) ClientParameters() *Parameters {
	return ss.clientParameters
}

func (ss *sessionStream) ServerParameters() *Parameters {
	return ss.serverParameters
}

func (ss *sessionStream) Version() protocol.Version {
	return ss.selectedVersion
}

func (ss *sessionStream) Close() error {
	return ss.stream.Close()
}

func (ss *sessionStream) CloseWithError(err error) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.closed {
		if ss.closeErr == nil {
			return fmt.Errorf("stream has already closed due to: %v", ss.closeErr)
		}

		return errors.New("stream has already closed")
	}

	if err == nil {
		err = ErrInternalError
	}

	ss.closed = true
	ss.closeErr = err

	var annerr TerminateError
	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := quic.StreamErrorCode(annerr.TerminateErrorCode())

	ss.stream.CancelRead(code)
	ss.stream.CancelWrite(code)

	slog.Debug("closed a session stream with an error",
		slog.Any("stream_id", ss.stream.StreamID()),
		slog.String("reason", err.Error()),
	)

	return nil
}

func (ss *sessionStream) listenUpdates() {
	var sum message.SessionUpdateMessage
	for {
		_, err := sum.Decode(ss.stream)
		if err != nil {
			slog.Error("failed to decode session update message", "error", err)
			return
		}

		slog.Debug("received a session update message",
			"stream_id", ss.stream.StreamID(),
			"bitrate", sum.Bitrate,
		)

		// TODO: Handle the session update message
	}
}

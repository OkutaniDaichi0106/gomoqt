package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type SessionStream interface {
	UpdateSession(bitrate uint64) error
	ClientParameters() *Parameters
	ServerParameters() *Parameters
	SelectedVersion() protocol.Version
}

func newSessionStream(version protocol.Version, clientParams *Parameters, serverParams *Parameters, stream quic.Stream) *sessionStream {
	return &sessionStream{
		stream:           stream,
		selectedVersion:  version,
		clientParameters: clientParams,
		serverParameters: serverParams,
	}
}

type sessionStream struct {
	stream quic.Stream

	/*
	 * Versions selected by the server
	 */
	selectedVersion protocol.Version

	clientParameters *Parameters

	serverParameters *Parameters
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

func (ss *sessionStream) SelectedVersion() protocol.Version {
	return ss.selectedVersion
}

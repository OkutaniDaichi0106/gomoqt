package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
)

type InfoHandler interface {
	HandleInfo(Info)
}

type InfoRequestHandler interface {
	HandleInfoRequest(InfoRequest, InfoWriter)
}

type InfoRequest message.InfoRequestMessage

type InfoRequestWriter interface {
	RequestInfo(InfoRequest) error
}

type Info message.InfoMessage

type InfoWriter interface {
	Answer(Info)
}

var _ InfoWriter = (*defaultInfoWriter)(nil)

type defaultInfoWriter struct {
	errCh  chan error
	stream Stream
}

func (w defaultInfoWriter) Answer(i Info) {
	_, err := w.stream.Write(message.InfoMessage(i).SerializePayload())
	if err != nil {
		slog.Error("failed to send an INFO message", slog.String("error", err.Error()))
		w.errCh <- err
		return
	}

	w.errCh <- nil
	slog.Info("answered information of a track")
}

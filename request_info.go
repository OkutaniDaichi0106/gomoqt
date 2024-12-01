package moqt

import (
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

// type InfoHandler interface {
// 	HandleInfo(Info)
// }

type InfoRequestHandler interface {
	HandleInfoRequest(InfoRequest, *Info, InfoWriter)
}

type InfoRequest struct {
	TrackNamespace string
	TrackName      string
}

type Info message.InfoMessage

type InfoWriter struct {
	doneCh chan struct{}
	stream moq.Stream
}

func (w InfoWriter) Answer(i Info) {
	err := message.InfoMessage(i).Encode(w.stream)
	if err != nil {
		slog.Error("failed to send an INFO message", slog.String("error", err.Error()))
		w.Reject(err)
		return
	}

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Info("answered an info")
}

func (w InfoWriter) Reject(err error) {
	if err == nil {
		err := w.stream.Close()
		if err != nil {
			slog.Debug("failed to close an Info Stream", slog.String("error", err.Error()))
		}

		return
	}

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		inferr, ok := err.(InfoError)
		if ok {
			code = moq.StreamErrorCode(inferr.InfoErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	w.stream.CancelRead(code)
	w.stream.CancelWrite(code)

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Info("rejected an info request")
}

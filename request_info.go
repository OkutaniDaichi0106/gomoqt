package moqt

import (
	"errors"
	"io"
	"log/slog"
	"time"

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
	TrackPath string
}

type Info struct {
	PublisherPriority   PublisherPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
	GroupExpires        time.Duration
}

type InfoWriter struct {
	stream moq.Stream
}

func (w InfoWriter) Answer(i Info) {
	im := message.InfoMessage{
		PublisherPriority:   message.PublisherPriority(i.PublisherPriority),
		LatestGroupSequence: message.GroupSequence(i.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(i.GroupOrder),
		GroupExpires:        i.GroupExpires,
	}

	err := im.Encode(w.stream)
	if err != nil {
		slog.Error("failed to send an INFO message", slog.String("error", err.Error()))
		w.Reject(err)
		return
	}

	slog.Info("answered an info")
}

func (w InfoWriter) Reject(err error) {
	if err == nil {
		w.Close()
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

	slog.Info("rejected an info request")
}

func (w InfoWriter) Close() {
	err := w.stream.Close()
	if err != nil {
		slog.Debug("failed to close an Info Stream", slog.String("error", err.Error()))
	}
}

func readInfo(r io.Reader) (Info, error) {
	// Read an INFO message
	var im message.InfoMessage
	err := im.Decode(r)
	if err != nil {
		slog.Error("failed to read a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	info := Info{
		PublisherPriority:   PublisherPriority(im.PublisherPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}

	return info, nil
}

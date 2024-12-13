package moqt

import (
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

/*
 * Fetch Stream
 */

type FetchStream struct {
	stream moq.Stream
}

func (f FetchStream) Read(buf []byte) (int, error) {
	return f.stream.Read(buf)
}

// func (f FetchStream) Group() Group {
// 	return f.group
// }

func (f FetchStream) CancelRead(code moq.StreamErrorCode) {
	f.stream.CancelRead(code)
}

func (f FetchStream) Close() error {
	err := f.stream.Close()
	if err != nil {
		slog.Error("failed to close a Fetch Stream", slog.String("error", err.Error()))
		return err
	}

	return nil
}

/*
 * Sequence number of a group in a track
 * When this is integer more than 1, the number means the sequence number.
 * When this is 0, it indicates the sequence number is currently unknown .
 * 0 is used to specify "the latest sequence number" or "the final sequence number of an open-ended track", "the first sequence number of the default order".
 */
type GroupSequence message.GroupSequence

/***/

type FetchHandler interface {
	HandleFetch(FetchRequest, FetchResponceWriter)
}

type FetchRequest struct {
	TrackPath     string
	TrackPriority Priority
	GroupSequence GroupSequence
	FrameSequence FrameSequence
}

type FetchResponceWriter struct {
	groupSent bool
	stream    moq.Stream
}

func (w *FetchResponceWriter) SendGroup(group Group) (moq.SendStream, error) {
	if w.groupSent {
		return nil, errors.New("a Group was already sent")
	} else {
		w.groupSent = true
	}

	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(group.subscribeID),
		GroupSequence: message.GroupSequence(group.groupSequence),
		GroupPriority: message.Priority(group.GroupPriority),
	}

	err := gm.Encode(w.stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return w.stream, nil
}

func (frw FetchResponceWriter) Reject(err error) {
	if err == nil {
		frw.Close()
	}

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = moq.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	frw.stream.CancelRead(code)
	frw.stream.CancelWrite(code)

	slog.Info("rejcted the fetch request")
}

func (frw FetchResponceWriter) Close() {
	err := frw.stream.Close()
	if err != nil {
		slog.Error("catch an error when closing a Fetch Stream", slog.String("error", err.Error()))
	}
}

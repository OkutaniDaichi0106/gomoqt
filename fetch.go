package moqt

import (
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

/*
 * Fetch Stream
 */
// var _ moq.ReceiveStream = (*FetchStream)(nil)

type FetchStream struct {
	group  *Group
	stream moq.Stream
}

// func (f FetchStream) StreamID() StreamID {
// 	return f.stream.StreamID()
// }

func (f FetchStream) Read(buf []byte) (int, error) {
	return f.stream.Read(buf)
}

func (f FetchStream) Group() Group {
	return *f.group
}

func (f FetchStream) SetReadDeadline(time time.Time) error {
	return f.stream.SetDeadLine(time)
}

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
type SubscriberPriority message.SubscriberPriority

type FetchHandler interface {
	HandleFetch(FetchRequest, FetchResponceWriter)
}

type FetchRequest message.FetchMessage

type FetchResponceWriter struct {
	doneCh chan struct{}
	stream moq.Stream
}

func (w FetchResponceWriter) SendGroup(group Group, data []byte) {
	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(group.subscribeID),
		GroupSequence:     message.GroupSequence(group.groupSequence),
		PublisherPriority: message.PublisherPriority(group.PublisherPriority),
	}

	err := gm.Encode(w.stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		w.doneCh <- ErrInternalError
		return
	}

	_, err = w.stream.Write(data)
	if err != nil {
		slog.Error("failed to send the data", slog.String("error", err.Error()))
		w.doneCh <- ErrInternalError
	}

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Info("sent data")
}

func (w FetchResponceWriter) Reject(err error) {
	if err == nil {
		err := w.stream.Close()
		if err != nil {
			slog.Error("failed to close a Fetch Stream", slog.String("error", err.Error()))
		}
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

	w.stream.CancelRead(code)
	w.stream.CancelWrite(code)

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Info("rejcted the fetch request")
}

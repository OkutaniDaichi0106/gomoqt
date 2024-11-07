package moqt

import (
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Fetch Stream
 */
var _ ReceiveStream = (*FetchStream)(nil)

type FetchStream struct {
	group  *Group
	stream Stream
}

func (f FetchStream) StreamID() StreamID {
	return f.StreamID()
}

func (f FetchStream) Read(buf []byte) (int, error) {
	if f.group == nil {
		f.Group()
	}

	return f.stream.Read(buf)
}

func (f FetchStream) Group() Group {
	if f.group == nil {
		var gm message.GroupMessage
		err := gm.DeserializePayload(quicvarint.NewReader(f.stream))
		if err != nil {
			slog.Error("failed to get a GROUP message", slog.String("error", err.Error()))
			return Group{}
		}

		f.group = &Group{
			SubscribeID:       SubscribeID(gm.SubscribeID),
			GroupSequence:     GroupSequence(gm.GroupSequence),
			PublisherPriority: gm.PublisherPriority,
		}
	}

	return *f.group
}

func (f FetchStream) SetReadDeadline(time time.Time) error {
	return f.stream.SetDeadLine(time)
}

func (f FetchStream) CancelRead(code StreamErrorCode) {
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

type GroupSequence message.GroupSequence

type SubscriberPriority message.SubscriberPriority

type FetchHandler interface {
	HandleFetch(FetchRequest, FetchResponceWriter)
}

type FetchRequest message.FetchMessage

type FetchResponceWriter struct {
	doneCh chan struct{}
	stream Stream
}

func (w FetchResponceWriter) SendGroup(group Group, data []byte) {
	gm := message.GroupMessage{
		SubscribeID:       message.SubscribeID(group.SubscribeID),
		GroupSequence:     message.GroupSequence(group.GroupSequence),
		PublisherPriority: group.PublisherPriority,
	}
	_, err := w.stream.Write(gm.SerializePayload())
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

	slog.Debug("sent data")
}

func (w FetchResponceWriter) Reject(err error) {
	if err == nil {
		err := w.stream.Close()
		if err != nil {
			slog.Error("failed to close a Fetch Stream", slog.String("error", err.Error()))
		}
	}

	var code StreamErrorCode

	var strerr StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = StreamErrorCode(feterr.FetchErrorCode())
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

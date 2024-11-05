package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
)

type FetchStream Stream
type GroupSequence message.GroupSequence

type SubscriberPriority message.SubscriberPriority

type FetchHandler interface {
	HandleFetch(FetchRequest, FetchResponceWriter)
}

type FetchRequest message.FetchMessage

type FetchResponceWriter interface {
	SendGroup(BufferStream, uint64)
	Reject(FetchError)
}

var _ FetchResponceWriter = (*defaultFetchRequestWriter)(nil)

type defaultFetchRequestWriter struct {
	errCh  chan error
	stream Stream
}

func (w defaultFetchRequestWriter) SendGroup(data BufferStream, offset uint64) {
	_, err := w.stream.Write(message.GroupMessage(data.group).SerializePayload())
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		w.errCh <- err
		return
	}

	buf := make([]byte, 1<<10)
	for {
		n, err := data.ReadOffset(buf, offset)
		if err != nil && err != io.EOF {
			slog.Error("failed to read a payload", slog.String("error", err.Error()))
			w.errCh <- err
			return
		}

		_, werr := w.stream.Write(buf[:n])
		if werr != nil {
			slog.Error("failed to send a payload", slog.String("error", werr.Error()))
			w.errCh <- werr
			return
		}

		if err == io.EOF {
			break
		}

		offset += uint64(n)
	}

	w.errCh <- nil
}

func (w defaultFetchRequestWriter) Reject(err FetchError) {
	w.stream.CancelRead(StreamErrorCode(err.FetchErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.FetchErrorCode()))

	w.errCh <- err
}

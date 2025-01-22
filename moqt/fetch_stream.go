package moqt

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type SendFetchStream interface {
	// Get a fetch request
	FetchRequest() FetchRequest

	// Update the fetch
	UpdateFetch(FetchUpdate) error

	// Close the stream
	Close() error // TODO: delete if not used

	// Close the stream with an error
	CloseWithError(err error) error // TODO: delete if not used or rename to CancelFetch
}

var _ SendFetchStream = (*sendFetchStream)(nil)

type sendFetchStream struct {
	stream transport.Stream
	fetch  FetchRequest
	mu     sync.Mutex
}

func (sfs *sendFetchStream) FetchRequest() FetchRequest {
	return sfs.fetch
}

func (sfs *sendFetchStream) UpdateFetch(update FetchUpdate) error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	fetch, err := updateFetch(sfs.fetch, update)
	if err != nil {
		return err
	}

	err = writeFetchUpdate(sfs.stream, update)
	if err != nil {
		slog.Error("failed to write a fetch update message", slog.String("error", err.Error()))
		return err
	}

	sfs.fetch = fetch

	slog.Debug("updated a fetch", slog.Any("fetch", sfs.fetch))

	return nil
}

func (sfs *sendFetchStream) CloseWithError(err error) error {
	if err == nil {
		return sfs.Close()
	}

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	sfs.stream.CancelRead(code)
	sfs.stream.CancelWrite(code)

	slog.Info("closed a fetch stream with an error", slog.String("error", err.Error()))

	return nil
}

func (sfs *sendFetchStream) Close() error {
	return sfs.stream.Close()
}

type ReceiveFetchStream interface {
	// Get a SendDataStream
	SendGroupStream

	// Get a fetch request
	FetchRequest() FetchRequest

	// Close the stream
	Close() error

	// Close the stream with an error
	CloseWithError(err error) error
}

var _ ReceiveFetchStream = (*receiveFetchStream)(nil)

type receiveFetchStream struct {
	fetch  FetchRequest
	stream transport.Stream
	mu     sync.Mutex
}

func (rfs *receiveFetchStream) SubscribeID() SubscribeID {
	return rfs.fetch.SubscribeID
}

func (rfs *receiveFetchStream) GroupSequence() GroupSequence {
	return rfs.fetch.GroupSequence
}

func (rfs *receiveFetchStream) CancelWrite(code GroupErrorCode) {
	rfs.stream.CancelWrite(transport.StreamErrorCode(code))
}

func (rfs *receiveFetchStream) SetWriteDeadline(t time.Time) error {
	return rfs.stream.SetWriteDeadline(t)
}

func (rfs *receiveFetchStream) WriteFrame(buf []byte) error {
	fm := message.FrameMessage{
		Payload: buf,
	}
	err := fm.Encode(rfs.stream)
	if err != nil {
		return err
	}

	return nil
}

func (rfs *receiveFetchStream) SendDataStream() SendGroupStream {
	return sendGroupStream{
		stream:    rfs.stream,
		sequence:  rfs.fetch.GroupSequence,
		startTime: time.Now(),
	}
}

func (rfs *receiveFetchStream) FetchRequest() FetchRequest {
	return rfs.fetch
}

func (rfs *receiveFetchStream) CloseWithError(err error) error {
	rfs.mu.Lock()
	defer rfs.mu.Unlock()

	if err == nil {
		return rfs.Close()
	}

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	rfs.stream.CancelRead(code)
	rfs.stream.CancelWrite(code)

	slog.Info("rejcted the fetch request")

	return nil
}

func (rfs *receiveFetchStream) Close() error {
	return rfs.stream.Close()
}

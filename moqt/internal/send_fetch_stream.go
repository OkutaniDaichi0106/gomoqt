package internal

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/quic-go/quic-go/quicvarint"
)

func newSendFetchStream(fm *message.FetchMessage, stream transport.Stream) *SendFetchStream {
	return &SendFetchStream{
		FetchMessage: *fm,
		Stream:       stream,
		mu:           sync.Mutex{},
	}
}

type SendFetchStream struct {
	FetchMessage message.FetchMessage
	Stream       transport.Stream
	mu           sync.Mutex
}

func (sfs *SendFetchStream) UpdateFetch(fum *message.FetchUpdateMessage) error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	if fum == nil {
		fum = &message.FetchUpdateMessage{}
	}

	_, err := fum.Encode(sfs.Stream)
	if err != nil {
		slog.Error("failed to write a fetch update message", slog.String("error", err.Error()))
		return err
	}

	updateFetch(&sfs.FetchMessage, fum)

	slog.Debug("updated a fetch", slog.Any("fetch", sfs.FetchMessage))

	return nil
}

func (sfs *SendFetchStream) ReadFrame() ([]byte, error) {
	bytes, _, err := message.ReadBytes(quicvarint.NewReader(sfs.Stream))
	return bytes, err
}

func (sfs *SendFetchStream) CancelRead(code StreamErrorCode) {
	sfs.Stream.CancelRead(transport.StreamErrorCode(code))
}

func (sfs *SendFetchStream) SetReadDeadline(t time.Time) error {
	return sfs.Stream.SetReadDeadline(t)
}

func (sfs *SendFetchStream) CloseWithError(err error) error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	slog.Debug("closing a send fetch stream with an error", slog.String("error", err.Error()))

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

	sfs.Stream.CancelRead(code)
	sfs.Stream.CancelWrite(code)

	slog.Info("closed a send fetch stream with an error", slog.String("error", err.Error()))

	return nil
}

func (sfs *SendFetchStream) Close() error {
	return sfs.Stream.Close()
}

func updateFetch(fm *message.FetchMessage, fum *message.FetchUpdateMessage) {
	if fum == nil {
		return
	}

	if fm == nil {
		fm = &message.FetchMessage{}
		return
	}

	// Update all fields
	fm.TrackPriority = fum.TrackPriority
}

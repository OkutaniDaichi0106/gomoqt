package internal

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveAnnounceStream(apm *message.AnnouncePleaseMessage, stream transport.Stream) *ReceiveAnnounceStream {
	ras := &ReceiveAnnounceStream{
		AnnouncePleaseMessage: *apm,
		stream:                stream,
		anns:                  make(map[string]message.AnnounceMessage),
		liveCh:                make(chan struct{}, 1),
	}

	go ras.listenAnnouncements()

	return ras
}

type ReceiveAnnounceStream struct {
	AnnouncePleaseMessage message.AnnouncePleaseMessage
	stream                transport.Stream
	mu                    sync.RWMutex

	anns     map[string]message.AnnounceMessage
	liveAnn  message.AnnounceMessage
	liveCh   chan struct{}
	closed   bool
	closeErr error
}

// func (ras *ReceiveAnnounceStream) LiveAnnouncement() chan<- *message.AnnounceMessage {
// 	return ras.liveAnnCh
// }

func (ras *ReceiveAnnounceStream) ReceiveAnnouncements() ([]message.AnnounceMessage, error) {
	ras.mu.RLock()
	defer ras.mu.RUnlock()

	if ras.closed {
		return nil, ras.closeErr
	}

	announcements := make([]message.AnnounceMessage, 0, len(ras.anns))
	//
	for _, ann := range ras.anns {
		announcements = append(announcements, ann)
	}

	return announcements, nil
}

func (ras *ReceiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		return nil
	}

	ras.closed = true

	close(ras.liveCh)
	ras.anns = nil
	ras.liveCh = nil

	return ras.stream.Close()
}

func (ras *ReceiveAnnounceStream) CloseWithErr(err error) error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		return ras.closeErr
	}

	if err == nil {
		return ras.Close()
	}

	ras.closeErr = err
	ras.closed = true

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

	ras.stream.CancelRead(code)
	ras.stream.CancelWrite(code)

	slog.Debug("closed a receive announce stream with an error", slog.String("error", err.Error()))

	return nil
}

// func (ras *ReceiveAnnounceStream) isValidateAnnouncement(am message.AnnounceMessage) bool {
// 	if ras.anns == nil {
// 		ras.anns = make(map[string]message.AnnounceMessage)
// 	}

// 	oldAnn, ok := ras.findAnnouncement(am.TrackPathSuffix)

// 	if ok && oldAnn.AnnounceStatus == am.AnnounceStatus {
// 		slog.Debug("duplicate announcement status")
// 	}

// 	if !ok && am.AnnounceStatus == message.ENDED {
// 		slog.Debug("ended track is not announced")
// 	}

// 	return true
// }

func (ras *ReceiveAnnounceStream) findAnnouncement(trackPath []string) (message.AnnounceMessage, bool) {
	ann, exists := ras.anns[TrackPartsString(trackPath)]
	return ann, exists
}

func (ras *ReceiveAnnounceStream) storeAnnouncement(am message.AnnounceMessage) {
	ras.anns[TrackPartsString(am.TrackSuffix)] = am
}

func TrackPartsString(trackPath []string) string {
	return strings.Join(trackPath, " ")
}

func (ras *ReceiveAnnounceStream) listenAnnouncements() {
	for {
		var am message.AnnounceMessage
		_, err := am.Decode(ras.stream)
		if err != nil {
			return
		}

		switch am.AnnounceStatus {
		case message.LIVE:
			ras.liveAnn = am
			ras.liveCh <- struct{}{}
		case message.ACTIVE:
			ann, exists := ras.findAnnouncement(am.TrackSuffix)
			if exists && ann.AnnounceStatus == message.ACTIVE {
				ras.CloseWithErr(ErrProtocolViolation)
				return
			}

			ras.storeAnnouncement(am)
		case message.ENDED:
			ann, exists := ras.findAnnouncement(am.TrackSuffix)
			if !exists {
				ras.CloseWithErr(ErrProtocolViolation)
				return
			}

			if ann.AnnounceStatus == message.ENDED {
				ras.CloseWithErr(ErrProtocolViolation)
				return
			}

			ras.storeAnnouncement(am)
		}
	}
}

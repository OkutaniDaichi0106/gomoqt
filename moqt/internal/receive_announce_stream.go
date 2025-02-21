package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveAnnounceStream(apm *message.AnnouncePleaseMessage, stream transport.Stream) *ReceiveAnnounceStream {
	ras := &ReceiveAnnounceStream{
		AnnouncePleaseMessage: *apm,
		stream:                stream,
	}

	return ras
}

type ReceiveAnnounceStream struct {
	AnnouncePleaseMessage message.AnnouncePleaseMessage
	stream                transport.Stream

	mu       sync.RWMutex
	closed   bool
	closeErr error
}

func (ras *ReceiveAnnounceStream) ReadAnnounceMessage(am *message.AnnounceMessage) error {
	_, err := am.Decode(ras.stream)
	return err
}

// func (ras *ReceiveAnnounceStream) ReceiveAnnouncements() ([]message.AnnounceMessage, error) {
// 	ras.mu.RLock()
// 	defer ras.mu.RUnlock()

// 	if ras.closed {
// 		return nil, ras.closeErr
// 	}

// 	announcements := make([]message.AnnounceMessage, 0, len(ras.announcements))

// 	//
// 	for _, ann := range ras.announcements {
// 		announcements = append(announcements, ann)
// 	}

// 	return announcements, nil
// }

func (ras *ReceiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		if ras.closeErr == nil {
			return fmt.Errorf("stream has already closed due to: %v", ras.closeErr)
		}

		return errors.New("stream has already closed")
	}

	ras.closed = true

	return ras.stream.Close()
}

func (ras *ReceiveAnnounceStream) CloseWithError(err error) error {
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

	var annerr AnnounceError
	var code protocol.AnnounceErrorCode

	if errors.As(err, &annerr) {
		code = annerr.AnnounceErrorCode()
	} else {
		code = ErrInternalError.AnnounceErrorCode()
	}

	ras.stream.CancelRead(transport.StreamErrorCode(code))
	ras.stream.CancelWrite(transport.StreamErrorCode(code))

	slog.Debug("closed a receive announce stream with an error", slog.String("error", err.Error()))

	return nil
}

// func (ras *ReceiveAnnounceStream) findAnnouncement(trackPath string) (message.AnnounceMessage, bool) {
// 	ras.mu.RLock()
// 	defer ras.mu.RUnlock()

// 	ann, exists := ras.announcements[trackPath]
// 	return ann, exists
// }

// func (ras *ReceiveAnnounceStream) storeAnnouncement(am message.AnnounceMessage) {
// 	ras.mu.Lock()
// 	defer ras.mu.Unlock()
// 	ras.announcements[am.TrackSuffix] = am
// }

// func (ras *ReceiveAnnounceStream) listenAnnouncements() {
// 	for {
// 		var am message.AnnounceMessage
// 		_, err := am.Decode(ras.stream)
// 		if err != nil {
// 			return
// 		}

// 		switch am.AnnounceStatus {
// 		case message.LIVE:
// 			ras.liveAnn = am
// 			ras.liveCh <- struct{}{}
// 		case message.ACTIVE:
// 			ann, exists := ras.findAnnouncement(am.TrackSuffix)
// 			if exists && ann.AnnounceStatus == message.ACTIVE {
// 				ras.CloseWithError(ErrProtocolViolation)
// 				return
// 			}

// 			ras.storeAnnouncement(am)
// 		case message.ENDED:
// 			// Check if the announcement exists
// 			ann, exists := ras.findAnnouncement(am.TrackSuffix)
// 			if !exists {
// 				ras.CloseWithError(ErrProtocolViolation)
// 				return
// 			}

// 			// Check if the announcement is ACTIVE
// 			if ann.AnnounceStatus == message.ENDED {
// 				ras.CloseWithError(ErrProtocolViolation)
// 				return
// 			}

// 			ras.storeAnnouncement(am)
// 		}
// 	}
// }

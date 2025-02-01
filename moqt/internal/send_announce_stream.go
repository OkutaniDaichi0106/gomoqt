package internal

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendAnnounceStream(stream transport.Stream, apm message.AnnouncePleaseMessage) *SendAnnounceStream {
	return &SendAnnounceStream{
		AnnouncePleaseMessage: apm,
		Stream:                stream,
	}
}

type SendAnnounceStream struct {
	AnnouncePleaseMessage message.AnnouncePleaseMessage

	Stream transport.Stream
	mu     sync.RWMutex
}

func (sas *SendAnnounceStream) SendActiveAnnouncement(suffix []string, params message.Parameters) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending announcements", slog.Any("trackPathSuffix", suffix), slog.Any("parameters", params))

	am := message.AnnounceMessage{
		AnnounceStatus:     message.ACTIVE,
		TrackSuffix:        suffix,
		AnnounceParameters: params,
	}

	// Encode the ANNOUNCE message
	_, err := am.Encode(sas.Stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("sent announcements", slog.Any("announcements", am))

	return nil
}

func (sas *SendAnnounceStream) SendEndedAnnouncement(suffix []string, params message.Parameters) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending announcements", slog.Any("trackPathSuffix", suffix), slog.Any("parameters", params))

	// oldAnn, exists := sas.findAnnouncement(trackSuffixString(suffix))
	// if !exists {
	// 	return ErrProtocolViolation // TODO: -> ErrTrackNotAnnounced
	// }
	// if oldAnn.AnnounceStatus == message.ENDED {
	// 	return ErrProtocolViolation // TODO: -> ErrDuplicateAnnouncement
	// }

	am := message.AnnounceMessage{
		AnnounceStatus:     message.ENDED,
		TrackSuffix:        suffix,
		AnnounceParameters: params,
	}

	// Encode the ANNOUNCE message
	_, err := am.Encode(sas.Stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (sas *SendAnnounceStream) SendLiveAnnouncement(params message.Parameters) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending announcements", slog.Any("parameters", params))

	am := message.AnnounceMessage{
		AnnounceStatus:     message.LIVE,
		AnnounceParameters: params,
	}

	// Encode the ANNOUNCE message
	_, err := am.Encode(sas.Stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// func (sas *SendAnnounceStream) isValidateAnnouncement(am message.AnnounceMessage) bool {
// 	if am.AnnounceStatus != message.LIVE && am.AnnounceStatus != message.ACTIVE && am.AnnounceStatus != message.ENDED {
// 		slog.Debug("invalid announcement status")
// 		return false
// 	}

// 	oldAnn, exists := sas.findAnnouncement(trackSuffixString(am.TrackPathSuffix))
// 	if exists && oldAnn.AnnounceStatus == am.AnnounceStatus {
// 		slog.Debug("duplicate announcement status")
// 		return false
// 	}

// 	if !exists && am.AnnounceStatus == message.ENDED {
// 		slog.Debug("ended track is not announced")
// 		return false
// 	}

// 	return true
// }

// func (sas *sendAnnounceStream) writeAndStoreAnnouncement(am message.AnnounceMessage) error {
// 	switch am.AnnounceStatus {
// 	case message.ACTIVE, message.ENDED:
// 		// Verify if the track path has the track prefix

// 		// Initialize an ANNOUNCE message
// 		am = message.AnnounceMessage{
// 			AnnounceStatus:  message.AnnounceStatus(ann.AnnounceStatus),
// 			TrackPathSuffix: suffix,
// 			Parameters:      message.Parameters(ann.AnnounceParameters.paramMap),
// 		}
// 	default:
// 		return ErrProtocolViolation
// 	}

// 	sas.storeAnnouncement(ann)

// 	return nil
// }

// func (sas *SendAnnounceStream) findAnnouncement(suffix string) (message.AnnounceMessage, bool) {
// 	am, exists := sas.annMap[suffix]
// 	return am, exists
// }

// func (sas *SendAnnounceStream) storeAnnouncement(am message.AnnounceMessage) {
// 	sas.annMap[trackSuffixString(am.TrackPathSuffix)] = am
// }

// func (sas *SendAnnounceStream) deleteAnnouncement(trackSuffix string) {
// 	delete(sas.annMap, trackSuffix)
// }

func (sas *SendAnnounceStream) Close() error {
	return sas.Stream.Close()
}

func (sas *SendAnnounceStream) CloseWithError(err error) error { // TODO
	slog.Debug("closing a send announce stream with an error", slog.String("error", err.Error()))

	if err == nil {
		return sas.Stream.Close()
	}

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		annerr, ok := err.(AnnounceError)
		if ok {
			code = transport.StreamErrorCode(annerr.AnnounceErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	sas.Stream.CancelRead(code)
	sas.Stream.CancelWrite(code)

	slog.Debug("closed a send announce stream with an error", slog.String("error", err.Error()))

	return nil
}

func trackSuffixString(suffixParts []string) string {
	return strings.Join(suffixParts, " ")
}

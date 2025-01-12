package moqt

import (
	"io"
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

// type InterestHandler interface {
// 	HandleInterest(Interest, AnnounceSender)
// }

const (
	ENDED  AnnounceStatus = AnnounceStatus(message.ENDED)
	ACTIVE AnnounceStatus = AnnounceStatus(message.ACTIVE)
	LIVE   AnnounceStatus = AnnounceStatus(message.LIVE)
)

type AnnounceStatus message.AnnounceStatus

type Announcement struct {
	AnnounceStatus AnnounceStatus

	/***/
	TrackPath string

	AnnounceParameters Parameters
}

func readAnnouncement(r io.Reader, prefix string) (Announcement, error) {
	var am message.AnnounceMessage
	err := am.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Get the full track path
	trackPath := prefix + "/" + am.TrackPathSuffix

	return Announcement{
		AnnounceStatus:     AnnounceStatus(am.AnnounceStatus),
		TrackPath:          trackPath,
		AnnounceParameters: Parameters(am.Parameters),
	}, nil
}

func writeAnnouncement(w io.Writer, prefix string, ann Announcement) error {
	var am message.AnnounceMessage
	switch ann.AnnounceStatus {
	case ACTIVE, ENDED:
		// Verify if the track path has the track prefix
		if !strings.HasPrefix(ann.TrackPath, prefix) {
			return ErrInternalError
		}

		// Get a suffix part of the Track Path
		suffix := strings.TrimPrefix(ann.TrackPath, prefix+"/")

		// Initialize an ANNOUNCE message
		am = message.AnnounceMessage{
			AnnounceStatus:  message.AnnounceStatus(ann.AnnounceStatus),
			TrackPathSuffix: suffix,
			Parameters:      message.Parameters(ann.AnnounceParameters),
		}
	case LIVE:
		// Initialize an ANNOUNCE message
		am = message.AnnounceMessage{
			AnnounceStatus: message.AnnounceStatus(ann.AnnounceStatus),
		}
	default:
		return ErrProtocolViolation
	}

	// Encode the ANNOUNCE message
	err := am.Encode(w)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully announced", slog.Any("announcement", ann))

	return nil
}

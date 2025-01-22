package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

const (
	ENDED  AnnounceStatus = AnnounceStatus(message.ENDED)
	ACTIVE AnnounceStatus = AnnounceStatus(message.ACTIVE)
	LIVE   AnnounceStatus = AnnounceStatus(message.LIVE)
)

type AnnounceStatus message.AnnounceStatus

type Announcement struct {
	AnnounceStatus AnnounceStatus

	/***/
	TrackPath []string

	AnnounceParameters Parameters
}

func readAnnouncement(r io.Reader, prefix []string) (Announcement, error) {
	var am message.AnnounceMessage
	err := am.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Get the full track path
	trackPath := append(prefix, am.TrackPathSuffix...)

	return Announcement{
		AnnounceStatus:     AnnounceStatus(am.AnnounceStatus),
		TrackPath:          trackPath,
		AnnounceParameters: Parameters{am.Parameters},
	}, nil
}

func writeAnnouncement(w io.Writer, prefix []string, ann Announcement) error {
	var am message.AnnounceMessage
	switch ann.AnnounceStatus {
	case ACTIVE, ENDED:
		// Verify if the track path has the track prefix
		if !hasPrefix(ann.TrackPath, prefix) {
			return ErrInternalError
		}

		// Get a suffix part of the Track Path
		suffix := trimPrefix(ann.TrackPath, prefix)

		// Initialize an ANNOUNCE message
		am = message.AnnounceMessage{
			AnnounceStatus:  message.AnnounceStatus(ann.AnnounceStatus),
			TrackPathSuffix: suffix,
			Parameters:      message.Parameters(ann.AnnounceParameters.paramMap),
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

func hasPrefix(path, prefix []string) bool {
	if len(path) < len(prefix) {
		return false
	}

	for i, p := range prefix {
		if path[i] != p {
			return false
		}
	}

	return true
}

func trimPrefix(path, prefix []string) []string {
	if len(path) < len(prefix) {
		return path
	}

	return path[len(prefix):]
}

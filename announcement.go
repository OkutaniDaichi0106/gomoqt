package moqt

import (
	"io"
	"log/slog"

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
	/***/
	TrackPath string

	/***/
	AuthorizationInfo  string
	AnnounceParameters Parameters
}

func readAnnouncement(r io.Reader, prefix string) (AnnounceStatus, Announcement, error) {
	var am message.AnnounceMessage
	err := am.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return 0, Announcement{}, err
	}

	// Get the full track path
	trackPath := prefix + "/" + am.TrackPathSuffix

	// Initialize an Announcement
	ann := Announcement{
		TrackPath:          trackPath,
		AnnounceParameters: Parameters(am.Parameters),
	}

	// Set the AuthorizationInfo
	if auth, ok := getAuthorizationInfo(ann.AnnounceParameters); ok {
		ann.AuthorizationInfo = auth
	}

	return AnnounceStatus(am.AnnounceStatus), ann, nil
}

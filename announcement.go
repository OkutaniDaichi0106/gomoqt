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
	AnnounceStatus AnnounceStatus

	/***/
	TrackPath string

	AnnounceParameters Parameters
}

func (a Announcement) AuthorizationInfo() (string, bool) {
	return getAuthorizationInfo(a.AnnounceParameters)
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

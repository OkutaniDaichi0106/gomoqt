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
	status AnnounceStatus

	/***/
	TrackPathSuffix string
	/***/
	AuthorizationInfo string
	Parameters        Parameters
}

func readAnnouncement(r io.Reader) (Announcement, error) {
	slog.Debug("reading an announcement")
	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err := am.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Initialize an Announcement
	announcement := Announcement{
		status:          AnnounceStatus(am.AnnounceStatus),
		TrackPathSuffix: am.TrackPathSuffix,
		Parameters:      Parameters(am.Parameters),
	}

	//
	authInfo, ok := getAuthorizationInfo(announcement.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

func writeAnnouncement(w io.Writer, announcement Announcement) error {
	slog.Debug("writing an announcement")

	// Add AUTHORIZATION_INFO parameter
	if announcement.AuthorizationInfo != "" {
		announcement.Parameters.Add(AUTHORIZATION_INFO, announcement.AuthorizationInfo)
	}

	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		TrackPathSuffix: announcement.TrackPathSuffix,
		Parameters:      message.Parameters(announcement.Parameters),
	}

	// Encode the ANNOUNCE message
	err := am.Encode(w)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully announced", slog.Any("announcement", announcement))

	return nil
}

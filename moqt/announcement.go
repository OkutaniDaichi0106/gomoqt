package moqt

import (
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

const (
	ENDED  AnnounceStatus = AnnounceStatus(message.ENDED)
	ACTIVE AnnounceStatus = AnnounceStatus(message.ACTIVE)
	LIVE   AnnounceStatus = AnnounceStatus(message.LIVE)
)

type AnnounceStatus byte

func (as AnnounceStatus) String() string {
	switch as {

	case ENDED:
		return "ENDED"
	case ACTIVE:
		return "ACTIVE"
	case LIVE:
		return "LIVE"
	default:
		return "UNKNOWN"
	}
}

type Announcement struct {
	AnnounceStatus AnnounceStatus

	/*
	 *
	 */
	TrackPath TrackPath

	/*
	 *
	 */

	AnnounceParameters Parameters
}

func (a Announcement) String() string {
	var sb strings.Builder
	sb.WriteString("Announcement: {")
	sb.WriteString(" ")
	sb.WriteString("AnnounceStatus: ")
	sb.WriteString(a.AnnounceStatus.String())
	sb.WriteString(", ")
	sb.WriteString("TrackPath: ")
	sb.WriteString(a.TrackPath.String())
	sb.WriteString(", ")
	sb.WriteString("AnnounceParameters: ")
	sb.WriteString(a.AnnounceParameters.String())

	sb.WriteString(" }")
	return sb.String()
}

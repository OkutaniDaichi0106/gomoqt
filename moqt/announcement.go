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

type AnnounceStatus message.AnnounceStatus

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

func IsSamePath(path, target []string) bool {
	if len(path) != len(target) {
		return false
	}

	for i, p := range target {
		if path[i] != p {
			return false
		}
	}

	return true
}

func HasPrefix(path, prefix []string) bool {
	return hasPrefix(path, prefix)
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

package moqt

import (
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
	TrackPath string
	/***/
	AuthorizationInfo  string
	AnnounceParameters Parameters
}

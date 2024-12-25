package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type TrackPriority message.TrackPriority
type GroupPriority message.GroupPriority

type GroupOrder message.GroupOrder

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

type data interface {
	//
	SubscribeID() SubscribeID
	TrackPriority() TrackPriority
	GroupOrder() GroupOrder

	//
	GroupPriority() GroupPriority
	GroupSequence() GroupSequence
}

func schedule(a, b data) bool {
	if a.SubscribeID() != b.SubscribeID() {
		if a.TrackPriority() != b.TrackPriority() {
			return a.TrackPriority() < b.TrackPriority()
		}
	}

	if a.GroupPriority() != b.GroupPriority() {
		return a.GroupPriority() < b.GroupPriority()
	}

	switch a.GroupOrder() {
	case DEFAULT:
		return true
	case ASCENDING:
		return a.GroupSequence() < b.GroupSequence()
	case DESCENDING:
		return a.GroupSequence() > b.GroupSequence()
	default:
	}

	return false
}

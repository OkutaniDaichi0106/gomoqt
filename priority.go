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

type priorityArgs struct {
	TrackPriority TrackPriority
	groupOrder    GroupOrder
	subscribeID   SubscribeID
	groupSequence GroupSequence
	groupPriority GroupPriority
}

func comparePriority(arg1, arg2 priorityArgs) bool {
	if arg1.TrackPriority != arg2.TrackPriority {
		return arg1.TrackPriority > arg2.TrackPriority
	}

	if arg1.groupPriority != arg2.groupPriority {
		return arg1.groupPriority > arg2.groupPriority
	}

	if arg1.subscribeID != arg2.subscribeID {

		// TODO:
		return true
	}

	// if arg1.TrackPath != arg2.TrackPath {
	// 	return false // TODO: handle this situation as an error
	// }
	if arg1.groupOrder != arg2.groupOrder {
		return false // TODO: handle this situation as an error
	}

	switch arg1.groupOrder {
	case DEFAULT:
		return true
	case ASCENDING:
		return arg1.groupSequence < arg2.groupSequence
	case DESCENDING:
		return arg1.groupSequence > arg2.groupSequence
	default:
		return false
	}
}

package moqt

import "github.com/OkutaniDaichi0106/gomoqt/internal/message"

type Priority message.Priority

type GroupOrder message.GroupOrder

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

func comparePriority(s1, s2 Subscription, g1, g2 Group) bool {
	if s1.TrackPriority != s2.TrackPriority {
		return s1.TrackPriority > s2.TrackPriority
	}

	if g1.GroupPriority != g2.GroupPriority {
		return g1.GroupPriority > g2.GroupPriority
	}

	if s1.subscribeID != s2.subscribeID {
		return false // TODO: Implement an ordering
	}

	subscription := s1

	switch subscription.GroupOrder {
	case DEFAULT:

	case ASCENDING:
		return g1.groupSequence < g2.groupSequence
	case DESCENDING:
		return g1.groupSequence > g2.groupSequence
	}
	// TODO:
	return true
}

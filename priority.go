package moqt

import "github.com/OkutaniDaichi0106/gomoqt/internal/message"

type SubscriberPriority message.SubscriberPriority

type PublisherPriority message.PublisherPriority

type GroupOrder message.GroupOrder

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

func comparePriority(s1, s2 Subscription, g1, g2 Group) bool {
	if s1.SubscriberPriority != s2.SubscriberPriority {
		return s1.SubscriberPriority > s2.SubscriberPriority
	}

	if g1.PublisherPriority != g2.PublisherPriority {
		return g1.PublisherPriority > g2.PublisherPriority
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

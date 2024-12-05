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

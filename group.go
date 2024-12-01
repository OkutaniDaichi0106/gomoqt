package moqt

import "github.com/OkutaniDaichi0106/gomoqt/internal/message"

type PublisherPriority message.PublisherPriority

type Group struct {
	subscribeID SubscribeID

	groupSequence GroupSequence

	PublisherPriority PublisherPriority
}

func (g Group) SubscribeID() SubscribeID {
	return g.subscribeID
}

func (g Group) GroupSequence() GroupSequence {
	return g.groupSequence
}

type GroupDrop message.GroupDrop

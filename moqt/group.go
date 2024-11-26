package moqt

import "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"

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

type GroupErrorCode message.GroupErrorCode

const (
	group_drop_track_not_exist GroupErrorCode = 0x00
	group_drop_internal_error  GroupErrorCode = 0x01
)

type GroupDrop message.GroupDrop

package moqt

import "github.com/OkutaniDaichi0106/gomoqt/moqt/message"

type Group struct {
	SubscribeID SubscribeID

	GroupSequence GroupSequence

	PublisherPriority message.PublisherPriority
}

type GroupErrorCode message.GroupErrorCode

const (
	group_drop_track_not_exist GroupErrorCode = 0x00
	group_drop_internal_error  GroupErrorCode = 0x01
)

type GroupDrop message.GroupDrop

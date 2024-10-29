package moqt

import "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"

type Group message.GroupMessage

type GroupHander interface {
	HandleGroup(Group)
}

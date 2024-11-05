package moqt

import "github.com/OkutaniDaichi0106/gomoqt/moqt/message"

type Group message.GroupMessage

type DataHander interface {
	HandleData(Group, ReceiveStream)
}

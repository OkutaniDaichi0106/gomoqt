package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type TrackPriority message.TrackPriority
type GroupPriority message.GroupPriority

type GroupOrder message.GroupOrder

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

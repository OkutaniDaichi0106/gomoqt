package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type GroupPeriod = protocol.GroupPeriod

const (
	GroupPeriodIrregular GroupPeriod = 0
	GroupPeriodSecond    GroupPeriod = 1000
	GroupPeriodMinute    GroupPeriod = 60 * GroupPeriodSecond
	GroupPeriodHour      GroupPeriod = 60 * GroupPeriodMinute
)

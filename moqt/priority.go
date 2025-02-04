package moqt

type TrackPriority byte

type GroupOrder byte

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

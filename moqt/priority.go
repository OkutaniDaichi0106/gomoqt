package moqt

type TrackPriority byte

type GroupOrder byte

func (order GroupOrder) String() string {
	switch order {
	case DEFAULT:
		return "default"
	case ASCENDING:
		return "ascending"
	case DESCENDING:
		return "descending"
	default:
		return "undefined group order"
	}
}

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

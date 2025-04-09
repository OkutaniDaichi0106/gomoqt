package moqt

type GroupOrder byte

func (order GroupOrder) String() string {
	switch order {
	case GroupOrderDefault:
		return "default"
	case GroupOrderAscending:
		return "ascending"
	case GroupOrderDescending:
		return "descending"
	default:
		return "undefined group order"
	}
}

const (
	GroupOrderDefault    GroupOrder = 0x0
	GroupOrderAscending  GroupOrder = 0x1
	GroupOrderDescending GroupOrder = 0x2
)

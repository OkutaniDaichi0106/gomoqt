package moqt

import "github.com/OkutaniDaichi0106/gomoqt/internal/transport"

// /*
//  * data interface is implemented by dataReceiveStream and receivedDatagram.
//  */
// type data interface {
// 	Group
// 	io.Reader
// }

type dataFragment interface {

	//
	SubscribeID() SubscribeID
	TrackPriority() TrackPriority
	GroupOrder() GroupOrder

	//
	GroupPriority() GroupPriority
	GroupSequence() GroupSequence

	Payload() []byte
}

var _ dataFragment = (*streamDataFragment)(nil)

type streamDataFragment struct {
	trackPriority TrackPriority
	groupOrder    GroupOrder

	streamID transport.StreamID
	receivedGroup
	payload []byte
}

func (d streamDataFragment) StreamID() transport.StreamID {
	return d.streamID
}

func (d streamDataFragment) TrackPriority() TrackPriority {
	return d.trackPriority
}

func (d streamDataFragment) GroupOrder() GroupOrder {
	return d.groupOrder
}

func (d streamDataFragment) Payload() []byte {
	return d.payload
}

var _ dataFragment = (*datagramData)(nil)

type datagramData struct {
	trackPriority TrackPriority
	groupOrder    GroupOrder

	receivedGroup
	payload []byte
}

func (d datagramData) TrackPriority() TrackPriority {
	return d.trackPriority
}

func (d datagramData) GroupOrder() GroupOrder {
	return d.groupOrder
}

func (d datagramData) Payload() []byte {
	return d.payload
}

/*
 * dataQueue implements heap.Interface.
 */
type dataQueue []dataFragment

func (q dataQueue) Len() int {
	return len(q)
}

func (q dataQueue) Less(i, j int) bool {
	return schedule(q[i], q[j])
}

func (q dataQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *dataQueue) Push(x interface{}) {
	*q = append(*q, x.(dataFragment))
}

func (q *dataQueue) Pop() interface{} {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[:n-1]
	return x
}

func schedule(a, b dataFragment) bool {
	if a.SubscribeID() != b.SubscribeID() {
		if a.TrackPriority() != b.TrackPriority() {
			return a.TrackPriority() < b.TrackPriority()
		}
	}

	if a.GroupPriority() != b.GroupPriority() {
		return a.GroupPriority() < b.GroupPriority()
	}

	switch a.GroupOrder() {
	case DEFAULT:
		return true
	case ASCENDING:
		return a.GroupSequence() < b.GroupSequence()
	case DESCENDING:
		return a.GroupSequence() > b.GroupSequence()
	default:
	}

	return false
}

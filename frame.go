package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type FrameSequence message.FrameSequence

type frame interface {

	//
	SubscribeID() SubscribeID
	TrackPriority() TrackPriority
	GroupOrder() GroupOrder

	//
	GroupPriority() GroupPriority
	GroupSequence() GroupSequence

	Payload() []byte
}

var _ frame = (*streamFrame)(nil)

type streamFrame struct {
	trackPriority TrackPriority
	groupOrder    GroupOrder

	streamID transport.StreamID
	receivedGroup
	payload []byte
}

func (d streamFrame) StreamID() transport.StreamID {
	return d.streamID
}

func (d streamFrame) TrackPriority() TrackPriority {
	return d.trackPriority
}

func (d streamFrame) GroupOrder() GroupOrder {
	return d.groupOrder
}

func (d streamFrame) Payload() []byte {
	return d.payload
}

var _ frame = (*datagramFrame)(nil)

type datagramFrame struct {
	trackPriority TrackPriority
	groupOrder    GroupOrder

	receivedGroup
	payload []byte
}

func (d datagramFrame) TrackPriority() TrackPriority {
	return d.trackPriority
}

func (d datagramFrame) GroupOrder() GroupOrder {
	return d.groupOrder
}

func (d datagramFrame) Payload() []byte {
	return d.payload
}

/*
 * frameQueue implements heap.Interface.
 */
type frameQueue []frame

func (q frameQueue) Len() int {
	return len(q)
}

func (q frameQueue) Less(i, j int) bool {
	return schedule(q[i], q[j])
}

func (q frameQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *frameQueue) Push(x interface{}) {
	*q = append(*q, x.(frame))
}

func (q *frameQueue) Pop() interface{} {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[:n-1]
	return x
}

func schedule(a, b frame) bool {
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

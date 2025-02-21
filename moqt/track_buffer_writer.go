package moqt

import "errors"

var _ TrackWriter = (*trackBufferWriter)(nil)

type trackBufferWriter struct {
	trackBuffer *TrackBuffer
	priotity    TrackPriority
	order       GroupOrder
}

func (tw *trackBufferWriter) TrackPath() TrackPath {
	return tw.trackBuffer.TrackPath()
}

func (tw *trackBufferWriter) TrackPriority() TrackPriority {
	return tw.priotity
}

func (tw *trackBufferWriter) GroupOrder() GroupOrder {
	return tw.order
}

func (tw *trackBufferWriter) LatestGroupSequence() GroupSequence {
	return tw.trackBuffer.LatestGroupSequence()
}

func (tw *trackBufferWriter) Info() Info {
	return Info{
		TrackPriority:       tw.priotity,
		LatestGroupSequence: tw.LatestGroupSequence(),
		GroupOrder:          tw.order,
	}
}

func (tw *trackBufferWriter) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if tw.trackBuffer.closed {
		return nil, errors.New("track buffer is closed")
	}

	gb := newGroupBuffer(seq, DefaultGroupBufferSize)

	if err := tw.trackBuffer.storeGroup(gb); err != nil {
		return nil, err
	}

	return newGroupBufferWriter(gb), nil
}

func (tw *trackBufferWriter) Close() error {
	return tw.trackBuffer.Close()
}

func (tw *trackBufferWriter) CloseWithError(err error) error {
	return tw.trackBuffer.CloseWithError(err)
}

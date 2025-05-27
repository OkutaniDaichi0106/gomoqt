package moqt

// var _ TrackWriter = (*trackBufferWriter)(nil)

// func newTrackBufferWriter(tb *TrackBuffer, priority TrackPriority, order GroupOrder) *trackBufferWriter {
// 	return &trackBufferWriter{
// 		trackBuffer: tb,
// 		priotity:    priority,
// 		order:       order,
// 	}
// }

// type trackBufferWriter struct {
// 	trackBuffer *TrackBuffer
// 	priotity    TrackPriority
// 	order       GroupOrder
// }

// func (tw *trackBufferWriter) TrackPath() TrackPath {
// 	return tw.trackBuffer.TrackPath()
// }

// func (tw *trackBufferWriter) LatestGroupSequence() GroupSequence {
// 	return tw.trackBuffer.LatestGroupSequence()
// }

// func (tw *trackBufferWriter) Info() Info {
// 	return tw.trackBuffer.Info()
// }

// func (tw *trackBufferWriter) OpenGroup(seq GroupSequence) (GroupWriter, error) {
// 	if tw.trackBuffer.closed {
// 		return nil, ErrClosedTrack
// 	}

// 	gb := newGroupBuffer(seq, DefaultGroupBufferSize)

// 	err := tw.trackBuffer.storeGroup(gb)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return gb, nil
// }

// func (tw *trackBufferWriter) Close() error {
// 	return tw.trackBuffer.Close()
// }

// func (tw *trackBufferWriter) CloseWithError(err error) error {
// 	return tw.trackBuffer.CloseWithError(err)
// }

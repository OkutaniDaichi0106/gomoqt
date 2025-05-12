package moqt

import (
	"time"
)

var _ GroupReader = (*groupBufferReader)(nil)

func newGroupBufferReader(buf *GroupBuffer) *groupBufferReader {
	return &groupBufferReader{
		groupBuffer: buf,
	}
}

type groupBufferReader struct {
	groupBuffer *GroupBuffer
	readFrames  int

	canceled    bool
	canceledErr GroupError
}

func (r *groupBufferReader) GroupSequence() GroupSequence {
	return r.groupBuffer.groupSequence
}

func (r *groupBufferReader) ReadFrame() (*Frame, error) {
	r.groupBuffer.cond.L.Lock()
	defer r.groupBuffer.cond.L.Unlock()

	for r.readFrames >= len(r.groupBuffer.frames) {
		// Check for reader cancellation
		if r.groupBuffer.closed {
			if r.groupBuffer.closedErr == nil {
				return nil, r.groupBuffer.closedErr
			}
			return nil, ErrClosedGroup
		}

		r.groupBuffer.cond.Wait()
	}

	f := r.groupBuffer.frames[r.readFrames]
	r.readFrames++
	return f, nil
}

func (r *groupBufferReader) SetReadDeadline(t time.Time) error {
	d := time.Until(t)

	// If the deadline is in the past, cancel the read immediately.
	if d <= 0 {
		r.CancelRead(ErrGroupExpired)
		return nil
	}

	// Cancel the read after the deadline.
	time.AfterFunc(d, func() {
		r.CancelRead(ErrGroupExpired)
	})

	return nil
}

func (r *groupBufferReader) CancelRead(err GroupError) {
	if r.canceled {
		return
	}

	r.canceled = true
	r.canceledErr = err
}

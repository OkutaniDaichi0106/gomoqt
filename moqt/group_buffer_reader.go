package moqt

import (
	"errors"
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
	closed      bool
	closeCh     chan struct{}
	closedErr   GroupError
}

func (r *groupBufferReader) GroupSequence() GroupSequence {
	return r.groupBuffer.groupSequence
}

func (r *groupBufferReader) ReadFrame() (*Frame, error) {
	r.groupBuffer.cond.L.Lock()
	defer r.groupBuffer.cond.L.Unlock()

	select {
	case <-r.closeCh:
		return nil, r.closedErr
	default:
		for r.readFrames >= len(r.groupBuffer.frames) {
			// Check for reader cancellation
			if r.closed {
				return nil, r.closedErr
			}
			if r.groupBuffer.closed {
				return nil, r.groupBuffer.closedErr
			}
			r.groupBuffer.cond.Wait()
		}
	}

	// Check cancellation before returning the frame.
	if r.closed {
		return nil, r.closedErr
	}

	f := r.groupBuffer.frames[r.readFrames]
	r.readFrames++
	return f, nil
}

func (r *groupBufferReader) SetReadDeadline(t time.Time) error {
	d := time.Until(t)

	// If the deadline is in the past, cancel the read immediately.
	if d <= 0 {
		r.groupBuffer.cond.L.Lock()
		defer r.groupBuffer.cond.L.Unlock()
		if r.closed {
			return nil
		}
		var grperr GroupError
		if errors.As(ErrGroupExpired, &grperr) {
			r.CancelRead(grperr)
		}
		return nil
	}

	//
	time.AfterFunc(d, func() {
		r.groupBuffer.cond.L.Lock()
		defer r.groupBuffer.cond.L.Unlock()
		if !r.closed {
			var grperr GroupError
			if errors.As(ErrGroupExpired, &grperr) {
				r.CancelRead(grperr)
			}
		}
	})

	return nil
}

func (r *groupBufferReader) CancelRead(err GroupError) {
	if r.closed {
		return
	}

	r.closed = true
	r.closedErr = err
	close(r.closeCh)
}

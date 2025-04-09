package moqt

// import (
// 	"errors"
// 	"fmt"
// 	"time"
// )

// var _ GroupWriter = (*groupBufferWriter)(nil)

// func newGroupBufferWriter(buf *GroupBuffer) *groupBufferWriter {
// 	return &groupBufferWriter{
// 		groupBuffer: buf,
// 	}
// }

// type groupBufferWriter struct {
// 	groupBuffer *GroupBuffer
// 	closed      bool
// 	closedErr   error
// }

// func (w *groupBufferWriter) GroupSequence() GroupSequence {
// 	return w.groupBuffer.groupSequence
// }

// func (w *groupBufferWriter) Close() error {
// 	w.groupBuffer.cond.L.Lock()
// 	defer w.groupBuffer.cond.L.Unlock()

// 	w.groupBuffer.cond.Broadcast()

// 	return nil
// }

// func (w *groupBufferWriter) WriteFrame(frame Frame) error {
// 	w.groupBuffer.cond.L.Lock()
// 	defer w.groupBuffer.cond.L.Unlock()

// 	if w.closed {
// 		if w.closedErr != nil {
// 			return w.closedErr
// 		}
// 		return ErrClosedGroup
// 	}

// 	w.groupBuffer.frames = append(w.groupBuffer.frames, frame)

// 	w.groupBuffer.cond.Broadcast()

// 	return nil
// }

// // New method: SetWriteDeadline schedules cancellation at the given time.
// func (w *groupBufferWriter) SetWriteDeadline(t time.Time) error {
// 	d := time.Until(t)

// 	// If the deadline is in the past, cancel the write immediately.
// 	if d <= 0 {
// 		return w.CloseWithError(ErrGroupExpired)
// 	}

// 	// Cancel the write after the deadline.
// 	time.AfterFunc(d, func() {
// 		if !w.closed {
// 			w.CloseWithError(ErrGroupExpired)
// 		}
// 	})
// 	return nil
// }

// func (w *groupBufferWriter) CloseWithError(err error) error {
// 	w.groupBuffer.cond.L.Lock()
// 	defer w.groupBuffer.cond.L.Unlock()

// 	if w.closed {
// 		if w.closedErr != nil {
// 			return fmt.Errorf("group has already closed due to: %w", w.closedErr)
// 		}
// 		return errors.New("group has already closed")
// 	}

// 	w.closed = true
// 	w.closedErr = err

// 	w.groupBuffer.cond.Broadcast()

// 	return nil
// }

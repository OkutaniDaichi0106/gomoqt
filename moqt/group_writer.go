package moqt

import "time"

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame(frame []byte) error
	CancelWrite(GroupErrorCode)
	SetWriteDeadline(time.Time) error
	Close() error
}

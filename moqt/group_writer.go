package moqt

import "time"

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame(*Frame) error
	CancelWrite(GroupErrorCode)
	SetWriteDeadline(time.Time) error
	Close() error
}

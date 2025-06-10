package moqt

import "time"

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame(*Frame) error
	CloseWithError(GroupErrorCode) error
	SetWriteDeadline(time.Time) error
	Close() error
}

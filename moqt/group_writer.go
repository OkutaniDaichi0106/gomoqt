package moqt

import "time"

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame(Frame) error
	CloseWithError(error) error
	SetWriteDeadline(time.Time) error
	Close() error
}

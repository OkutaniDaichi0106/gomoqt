package moqt

import "time"

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame(frame []byte) error
	CloseWithError(error) error
	SetWriteDeadline(time.Time) error
	Close() error
}

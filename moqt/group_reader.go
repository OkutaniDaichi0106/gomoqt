package moqt

import "time"

/*
 * Group Reader
 */
type GroupReader interface {
	GroupSequence() GroupSequence
	ReadFrame() ([]byte, error)
	CancelRead(GroupError)
	SetReadDeadline(time.Time) error
}

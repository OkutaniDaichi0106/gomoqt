package moqt

import "time"

/*
 * Group Reader
 */
type GroupReader interface {
	GroupSequence() GroupSequence
	ReadFrame() (*Frame, error)
	CancelRead(GroupError)
	SetReadDeadline(time.Time) error
}

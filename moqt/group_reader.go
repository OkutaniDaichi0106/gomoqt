package moqt

import "time"

/*
 * Group Reader
 */
type GroupReader interface {
	GroupSequence() GroupSequence
	ReadFrame() (*Frame, error)
	CancelRead(GroupErrorCode)
	SetReadDeadline(time.Time) error
}

package moqt

import "time"

/*
 * Group Reader
 */
type GroupReader interface {
	GroupSequence() GroupSequence
	ReadFrame() ([]byte, error)
	CancelRead(GroupErrorCode)
	SetReadDeadline(time.Time) error
}

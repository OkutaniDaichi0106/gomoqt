package moqt

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame([]byte) error
	Close() error
}

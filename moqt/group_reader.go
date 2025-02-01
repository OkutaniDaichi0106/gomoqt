package moqt

/*
 * Group Reader
 */
type GroupReader interface {
	GroupSequence() GroupSequence
	ReadFrame() ([]byte, error)
}

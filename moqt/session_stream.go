package moqt

type SessionStream interface {
	UpdateSession(bitrate uint64) error
}

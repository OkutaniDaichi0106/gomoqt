package moqt

// type TrackManager struct {
// 	LocalTrack  []string
// 	RemoteTrack []string
// }

type RelayManager interface {
	LocalTrack() string
	RemoteTrack() []string
}

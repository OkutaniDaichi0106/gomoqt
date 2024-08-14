package gomoq

// Moqt version
type Version int

const (
	Draft01  Version = 0xff000001
	Draft02  Version = 0xff000002
	Draft03  Version = 0xff000003
	Draft04  Version = 0xff000004
	Draft05  Version = 0xff000005
	Stable01 Version = 0x00000001
)

func DefaultVersion() Version {
	return Draft05
}

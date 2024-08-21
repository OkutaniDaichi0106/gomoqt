package gomoq

// Moqt version
type Version int

const (
	Draft01  Version = 0xff000001 /* Not implemented */
	Draft02  Version = 0xff000002 /* Not implemented */
	Draft03  Version = 0xff000003 /* Not implemented */
	Draft04  Version = 0xff000004 /* Not implemented */
	Draft05  Version = 0xff000005 /* Implemented */
	Stable01 Version = 0x00000001 /* Not implemented */
)

func DefaultVersion() Version {
	return Draft05
}

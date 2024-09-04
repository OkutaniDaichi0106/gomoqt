package gomoq

// Moqt version
type Version int

const (
	INVALID_VERSION Version = 0x0
	Draft01         Version = 0xff000001 /* Not implemented */
	Draft02         Version = 0xff000002 /* Not implemented */
	Draft03         Version = 0xff000003 /* Not implemented */
	Draft04         Version = 0xff000004 /* Not implemented */
	Draft05         Version = 0xff000005 /* Partly Implemented */
	LATEST          Version = 0xffffffff /* Partly Implemented */
	Stable01        Version = 0x00000001 /* Not implemented */
)

func DefaultVersion() Version {
	return Draft05
}

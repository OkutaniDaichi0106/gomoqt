package moqt

// Development and draft versions for MOQ.
// These can be used to represent supported protocol versions in a client or server.
const (
	// Default Version
	Default Version = Development

	// MoQ Lite Draft Versions
	LiteDraft01 Version = 0xff0dad01
	LiteDraft02 Version = 0xff0dad02

	// This implement version
	Development Version = 0xfeedbabe
)

// DefaultClientVersions lists the versions offered by default by a client during session setup.
var DefaultClientVersions []Version = []Version{Default}

// DefaultServerVersion is the version a server selects by default when accepting a session.
// It is selected when the server's configuration does not otherwise specify the selection behavior.
var DefaultServerVersion Version = Default

// Version identifies a protocol version for MOQ transports and negotiation.
type Version uint64

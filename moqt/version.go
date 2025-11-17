package moqt

// Development and draft versions for MOQ.
// These can be used to represent supported protocol versions in a client or server.
const (
	/*
	 * Develop Version
	 * These versions start with 0xffffff...
	 */
	Develop Version = 0xffffff00

	/*
	 * MOQTransfork draft versions
	 * These versions start with 0xff0bad...
	 */
	Draft01 Version = 0xff0bad01
	Draft02 Version = 0xff0bad02
	Draft03 Version = 0xff0bad03
)

// DefaultClientVersions lists the versions offered by default by a client during session setup.
var DefaultClientVersions []Version = []Version{Develop}

// DefaultServerVersion is the version a server selects by default when accepting a session.
// It is selected when the server's configuration does not otherwise specify the selection behavior.
var DefaultServerVersion Version = Develop

// Version identifies a protocol version for MOQ transports and negotiation.
type Version uint64

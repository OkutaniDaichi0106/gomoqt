package moqt

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

var DefaultClientVersions []Version = []Version{Develop}

var DefaultServerVersion Version = Develop

type Version uint64

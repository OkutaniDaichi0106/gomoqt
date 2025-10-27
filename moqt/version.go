package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
)

var DefaultClientVersions []Version = []Version{internal.Develop}

var DefaultServerVersion Version = internal.Develop

type Version = internal.Version

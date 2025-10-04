package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

var DefaultClientVersions []Version = []Version{protocol.Develop}

var DefaultServerVersion Version = protocol.Develop

type Version = protocol.Version

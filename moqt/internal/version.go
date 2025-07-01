package internal

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

var DefaultClientVersions = []protocol.Version{protocol.Develop}

var DefaultServerVersion = protocol.Develop

// func ContainVersion(version protocol.Version, versions []protocol.Version) bool {
// 	versionMap := make(map[protocol.Version]bool, len(versions))
// 	for _, v := range versions {
// 		versionMap[v] = true
// 	}

// 	return versionMap[version]
// }

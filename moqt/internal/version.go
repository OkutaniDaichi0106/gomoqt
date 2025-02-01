package internal

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

var DefaultClientVersions = []protocol.Version{Default}

var DefaultServerVersion = Default

const (
	Default = protocol.Develop
	//Draft01 Version = Version(protocol.Draft01)
	Develop protocol.Version = protocol.Develop
)

func ContainVersion(version protocol.Version, versions []protocol.Version) bool {
	versionMap := make(map[protocol.Version]bool, len(versions))
	for _, v := range versions {
		versionMap[v] = true
	}

	return versionMap[version]
}

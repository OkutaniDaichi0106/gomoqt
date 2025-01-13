package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
)

type Version protocol.Version

var DefaultClientVersions = []Version{Default}

var DefaultServerVersion = Default

const (
	Default = Develop
	//Draft01 Version = Version(protocol.Draft01)
	Develop Version = Version(protocol.Develop)
)

func ContainVersion(version Version, versions []Version) bool {
	versionMap := make(map[Version]bool, len(versions))
	for _, v := range versions {
		versionMap[v] = true
	}

	return versionMap[version]
}

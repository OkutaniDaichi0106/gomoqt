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

// /*
//  * Select a latest moqt version from a pair of version sets
//  */
// func SelectLatestVersion(vs1, vs2 []Version) (Version, bool) {
// 	// Get version-bool mapping
// 	versionMap := make(map[Version]bool, len(vs1))
// 	for _, v := range vs1 {
// 		versionMap[v] = true
// 	}

// 	// Get common versions
// 	commonVersions := []Version{}
// 	for _, v := range vs2 {
// 		if versionMap[v] {
// 			commonVersions = append(commonVersions, v)
// 		}
// 	}

// 	// Verify if common versions between the sets exist
// 	if len(commonVersions) < 1 {
// 		return 0, false
// 	}

// 	// Select the latest Vesion
// 	latestVersion := commonVersions[0]
// 	for _, version := range commonVersions[1:] {
// 		if latestVersion < version {
// 			latestVersion = version
// 		}
// 	}

// 	return latestVersion, true
// }

func ContainVersion(version Version, versions []Version) bool {
	versionMap := make(map[Version]bool, len(versions))
	for _, v := range versions {
		versionMap[v] = true
	}

	return versionMap[version]
}

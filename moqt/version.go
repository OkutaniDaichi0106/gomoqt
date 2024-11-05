package moqt

import (
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type Version protocol.Version

const (
	Default         = Draft01
	Draft01 Version = 0xffffff01
	Devlop  Version = 0xffffff00
)

/*
 * Select a latest moqt version from a pair of version sets
 */
func SelectLatestVersion(vs1, vs2 []Version) (Version, error) {
	// Get version-bool mapping
	versionMap := make(map[Version]bool, len(vs1))
	for _, v := range vs1 {
		versionMap[v] = true
	}

	// Get common versions
	commonVersions := []Version{}
	for _, v := range vs2 {
		if versionMap[v] {
			commonVersions = append(commonVersions, v)
		}
	}

	// Verify if common versions between the sets exist
	if len(commonVersions) < 1 {
		return 0, errors.New("no common versions")
	}

	// Select the latest Vesion
	latestVersion := commonVersions[0]
	for _, version := range commonVersions[1:] {
		if latestVersion < version {
			latestVersion = version
		}
	}

	return latestVersion, nil
}

func ContainVersion(version Version, versions []Version) bool {
	versionMap := make(map[Version]bool, len(versions))
	for _, v := range versions {
		versionMap[v] = true
	}

	return versionMap[version]
}

func getVersions(vs []uint64) []Version {
	versions := make([]Version, len(vs))
	for _, v := range vs {
		versions = append(versions, Version(v))
	}

	return versions
}

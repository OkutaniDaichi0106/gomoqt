package moqtmessage

import (
	"errors"
)

type Version int

const (
	FoalkDraft01 Version = 0xffffff01
)

/*
 * Select a latest moqt version from a pair of version sets
 */
func SelectLatestVersion(vs1, vs2 []Version) (Version, error) {
	// Get common Versions
	versionMap := make(map[Version]bool, len(vs1))
	for _, v := range vs1 {
		versionMap[v] = true
	}

	commonVersions := []Version{}
	for _, v := range vs2 {
		if versionMap[v] {
			commonVersions = append(commonVersions, v)
			delete(versionMap, v)
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
	for _, v := range versions {
		if v == version {
			return true
		}
	}

	return false
}

//var ErrVersionNotFound = errors.New("version not found")

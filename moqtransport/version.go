package moqtransport

import (
	"errors"
)

// Moqt version
type Version int

const (
	INVALID_VERSION Version = 0x0
	Draft01         Version = 0xff000001 /* Not implemented */
	Draft02         Version = 0xff000002 /* Not implemented */
	Draft03         Version = 0xff000003 /* Not implemented */
	Draft04         Version = 0xff000004 /* Not implemented */
	Draft05         Version = 0xff000005 /* Partly Implemented */
	LATEST          Version = 0xffffffff /* Partly Implemented */
	Stable01        Version = 0x00000001 /* Not implemented */
)

func DefaultVersion() Version {
	return Draft05
}

/*
 * Select a newest moqt version from a pair of version sets
 */
func selectVersion(vs1, vs2 []Version) (Version, error) {
	// Register a slice of Versions as map
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
	if len(commonVersions) < 1 {
		// Throw an error if there are no common versions between the sets
		return INVALID_VERSION, errors.New("no valid versions")
	}

	// Set the first Version as the latest version
	latestVersion := commonVersions[0]
	if len(commonVersions) > 1 {
		// Select latest Version
		for _, version := range commonVersions {
			if latestVersion < version {
				latestVersion = version
			}
		}
	}
	return latestVersion, nil
}

func contain(version Version, versions []Version) error {
	for _, v := range versions {
		if v != version {
			continue
		}
		return nil
	}
	return ErrVersionNotFound
}

var ErrVersionNotFound = errors.New("version not found")

package moqt

import (
	"strings"
)

type TrackPath string

func (tp TrackPath) String() string {
	return string(tp)
}

func (tp TrackPath) HasPrefix(prefix string) bool {
	// If path length is shorter than prefix, return false
	if len(tp) < len(prefix) {
		return false
	}

	return strings.HasPrefix(string(tp), prefix+"/")
}

func (tp TrackPath) HasSuffix(suffix string) bool {
	// If path length is shorter than suffix, return false
	if len(tp) < len(suffix) {
		return false
	}

	return strings.HasSuffix(string(tp), "/"+suffix)
}

func (tp TrackPath) Equal(target TrackPath) bool {
	return tp == target
}

func (tp TrackPath) Match(pattern string) bool {
	return matchGlob(pattern, string(tp))
}

// matchGlob checks if the given wildcard pattern matches the path
// * matches any string within a single segment
// ** matches multiple segments (directory levels)
func matchGlob(pattern, pathStr string) bool {
	if pattern == "" && pathStr == "" {
		return true
	}
	if pattern == "" || pathStr == "" {
		return false
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(pathStr, "/")

	return matchParts(patternParts, pathParts)
}

// matchParts compares pattern parts and path parts arrays to determine a match
func matchParts(patternParts, pathParts []string) bool {
	var i, j int

	for i < len(patternParts) && j < len(pathParts) {
		// For **, check for matches across multiple segments
		if patternParts[i] == "**" {
			// If ** is the last part of pattern, consider all remaining path as matching
			if i == len(patternParts)-1 {
				return true
			}

			// Try to match the pattern after ** with remaining paths
			for k := j; k < len(pathParts); k++ {
				if matchParts(patternParts[i+1:], pathParts[k:]) {
					return true
				}
			}
			return false
		} else if patternParts[i] == "*" {
			// * matches any string within a single segment
			i++
			j++
		} else if patternParts[i] == pathParts[j] {
			// If strings match exactly
			i++
			j++
		} else {
			return false
		}
	}

	// Only consider a match if both arrays are fully processed
	return i == len(patternParts) && j == len(pathParts)
}

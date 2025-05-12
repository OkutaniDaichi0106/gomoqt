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
	return matchGlob(pattern, string(tp), nil)
}

func (tp TrackPath) IsRoot() bool {
	return tp == "/"
}

func (tp TrackPath) ExtractParameters(pattern string) []string {
	// Count the number of * or ** in the pattern
	count := strings.Count(pattern, "*") - strings.Count(pattern, "**")

	if count == 0 {
		return nil
	}

	params := make([]string, 0, count)

	matched := matchGlob(pattern, string(tp), &params)
	if matched {
		return params
	}

	return nil
}

// matchGlob checks if the given wildcard pattern matches the path
// * matches any string within a single segment
// ** matches multiple segments (directory levels)
func matchGlob(pattern, pathStr string, variables *[]string) bool {
	if pattern == "" && pathStr == "" {
		return true
	}
	if pattern == "" || pathStr == "" {
		return false
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(pathStr, "/")

	return matchParts(patternParts, pathParts, variables)
}

// matchParts compares pattern parts and path parts arrays to determine a match
func matchParts(patternParts, pathParts []string, variables *[]string) bool {
	var i, j int

	for i < len(patternParts) && j < len(pathParts) {
		// For **, check for matches across multiple segments
		if patternParts[i] == "**" {
			// If ** is the last part of pattern, consider all remaining path as matching
			if i == len(patternParts)-1 {
				if variables != nil {
					*variables = append(*variables, strings.Join(pathParts[j:], "/"))
				}
				return true
			}

			// Try to match the pattern after ** with remaining paths
			for k := j; k < len(pathParts); k++ {
				if matchParts(patternParts[i+1:], pathParts[k:], variables) {
					if variables != nil {
						*variables = append(*variables, strings.Join(pathParts[j:k], "/"))
					}
					return true
				}
			}
			return false
		} else if patternParts[i] == "*" {
			// * matches any string within a single segment
			if variables != nil {
				*variables = append(*variables, pathParts[j])
			}

			// Increment both pointers
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

func NewTrackPath(pattern string, segments ...string) TrackPath {
	count := strings.Count(pattern, "*") - strings.Count(pattern, "**")

	if count > len(segments) {
		if count > cap(segments) {
			old := segments
			segments = make([]string, count)
			copy(segments, old)
		} else {
			len := len(segments)
			for i := range len - count {
				segments[i] = ""
			}
		}
	} else if count < len(segments) {
		segments = segments[:count]
	}

	var pathStr string

	var prefix, after string
	var ok bool

	after = pattern
	for i := range count {
		prefix, after, ok = strings.Cut(after, "*")
		if !ok {
			break
		}

		after = strings.TrimPrefix(after, "*")

		pathStr += prefix + segments[i]

		if i == count-1 {
			pathStr += after
		}
	}

	return TrackPath(pathStr)
}

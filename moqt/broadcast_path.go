package moqt

import (
	"strings"
)

// BroadcastPath represents a hierarchical path used to identify a group of related tracks.
// Paths use forward slashes as separators, similar to URL paths (e.g., "live/camera1").
type BroadcastPath string

// String returns the string representation of the broadcast path.
func (bc BroadcastPath) String() string {
	return string(bc)
}

// HasPrefix checks if the broadcast path starts with the given prefix.
func (bc BroadcastPath) HasPrefix(prefix string) bool {
	// If path length is shorter than prefix, return false
	if len(bc) < len(prefix) {
		return false
	}
	return strings.HasPrefix(string(bc), prefix)
}

// GetSuffix returns the path suffix after removing the given prefix.
// Returns empty string and false if the path doesn't have the prefix.
func (bc BroadcastPath) GetSuffix(prefix string) (string, bool) {
	if !bc.HasPrefix(prefix) {
		return "", false
	}

	return strings.TrimPrefix(string(bc), prefix), true
}

// Extension returns the file extension of the path (e.g., ".mp4") if present.
func (bc BroadcastPath) Extension() string {
	if i := strings.LastIndex(string(bc), "."); i >= 0 {
		return string(bc)[i:]
	}

	return ""
}

// Equal checks if two broadcast paths are identical.
func (bc BroadcastPath) Equal(target BroadcastPath) bool {
	return bc == target
}

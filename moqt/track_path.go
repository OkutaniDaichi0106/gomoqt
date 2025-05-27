package moqt

import (
	"strings"
)

type BroadcastPath string

func (bc BroadcastPath) String() string {
	return string(bc)
}

func (bc BroadcastPath) HasPrefix(prefix string) bool {
	// If path length is shorter than prefix, return false
	if len(bc) < len(prefix) {
		return false
	}

	return strings.HasPrefix(string(bc), prefix+"/")
}

func (bc BroadcastPath) HasSuffix(suffix string) bool {
	// If path length is shorter than suffix, return false
	if len(bc) < len(suffix) {
		return false
	}

	return strings.HasSuffix(string(bc), "/"+suffix)
}

func (bc BroadcastPath) GetSuffix(prefix string) (string, bool) {
	if !bc.HasPrefix(prefix) {
		return "", false
	}

	return strings.TrimPrefix(string(bc), prefix+"/"), true
}

func (bc BroadcastPath) Extension() string {
	if i := strings.LastIndex(string(bc), "."); i >= 0 {
		return string(bc)[i:]
	}

	return ""
}

func (bc BroadcastPath) Equal(target BroadcastPath) bool {
	return bc == target
}

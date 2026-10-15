package moqt

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidPathPattern is returned for a path pattern that is not a
// broadcast path or a prefix of them. Received on an Auth Stream, it is a
// protocol violation.
var ErrInvalidPathPattern = errors.New("moqt: invalid path pattern")

// PathPattern matches broadcast paths the way an http.ServeMux pattern
// without wildcards matches URL paths, and the way TrackMux and
// Session.AcceptAnnounce read paths and prefixes: "/room/1/alice" matches
// that path alone, and a pattern ending in "/", such as "/room/1/", matches
// every path beneath it. "/" matches every path.
type PathPattern struct {
	raw string
}

// ParsePathPattern parses s: a path starting with "/", with no empty segment
// before its last. "{" and "}" are reserved for wildcards and rejected.
func ParsePathPattern(s string) (PathPattern, error) {
	if !isValidPath(BroadcastPath(s)) || strings.Contains(s, "//") || strings.ContainsAny(s, "{}") {
		return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
	}
	return PathPattern{raw: s}, nil
}

// MustParsePathPattern is like ParsePathPattern but panics on an invalid
// pattern. It is meant for patterns written in code.
func MustParsePathPattern(s string) PathPattern {
	p, err := ParsePathPattern(s)
	if err != nil {
		panic(err)
	}
	return p
}

// String returns the pattern as written.
func (p PathPattern) String() string {
	return p.raw
}

// Match reports whether path is in the pattern's set. The zero PathPattern
// matches nothing.
func (p PathPattern) Match(path BroadcastPath) bool {
	if isValidPrefix(p.raw) {
		return path.HasPrefix(p.raw)
	}
	return p.raw != "" && string(path) == p.raw
}

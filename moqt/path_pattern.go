package moqt

import (
	"errors"
	"fmt"
	"go/token"
	"strings"
)

// ErrInvalidPathPattern is returned for a path pattern that is not a
// broadcast path or a prefix of them. Received on an Auth Stream, it is a
// protocol violation.
var ErrInvalidPathPattern = errors.New("moqt: invalid path pattern")

// PathPattern matches broadcast paths the way an http.ServeMux pattern
// matches URL paths, and the way TrackMux and Session.AcceptAnnounce read
// paths and prefixes: "/room/1/alice" matches that path alone, and a pattern
// ending in "/", such as "/room/1/", matches every path beneath it. "/"
// matches every path.
//
// A segment may be a wildcard. "{name}" matches any one non-empty segment, so
// "/room/{id}/host" matches "/room/1/host", and "/room/{id}/" matches every
// path beneath any room. "{name...}" as the last segment matches the rest of
// the path, even an empty rest, so "/room/{rest...}" matches what "/room/"
// does. "{$}" as the last segment matches only the "/" before it:
// "/room/{$}" matches "/room/" alone. Names are Go identifiers, unique within
// a pattern.
type PathPattern struct {
	raw string
	// segs holds the segments of a pattern with a wildcard, without the
	// leading "/" and a trailing "/".
	segs []string
}

// ParsePathPattern parses s: a path starting with "/" whose wildcards are
// whole segments. An empty segment, as in "/room//x", is a literal one.
func ParsePathPattern(s string) (PathPattern, error) {
	if !isValidPath(BroadcastPath(s)) {
		return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
	}
	if !strings.ContainsAny(s, "{}") {
		return PathPattern{raw: s}, nil
	}

	segs := strings.Split(strings.TrimSuffix(s[1:], "/"), "/")
	names := make(map[string]bool, len(segs))
	for i, seg := range segs {
		if !strings.ContainsAny(seg, "{}") {
			continue
		}
		if seg == anchorSegment {
			if i != len(segs)-1 || strings.HasSuffix(s, "/") {
				return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
			}
			continue
		}
		name, rest, ok := parseWildcard(seg)
		if !ok || names[name] || rest && (i != len(segs)-1 || strings.HasSuffix(s, "/")) {
			return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
		}
		names[name] = true
	}
	return PathPattern{raw: s, segs: segs}, nil
}

// anchorSegment ends a pattern that matches its own trailing "/" alone.
const anchorSegment = "{$}"

// parseWildcard reads a "{name}" or "{name...}" segment.
func parseWildcard(seg string) (name string, rest, ok bool) {
	inner, ok := strings.CutPrefix(seg, "{")
	if !ok {
		return "", false, false
	}
	if inner, ok = strings.CutSuffix(inner, "}"); !ok {
		return "", false, false
	}
	name, rest = strings.CutSuffix(inner, "...")
	return name, rest, token.IsIdentifier(name)
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
	return p.match(path, nil)
}

// Values reports whether path is in the pattern's set and, if it is, returns
// the value each named wildcard matched, like http.Request.PathValue: for
// "/room/{id}/{rest...}", the path "/room/1/a/b" gives id "1" and rest "a/b".
// A pattern without a named wildcard gives an empty map.
func (p PathPattern) Values(path BroadcastPath) (map[string]string, bool) {
	values := make(map[string]string)
	if !p.match(path, values) {
		return nil, false
	}
	return values, true
}

// match reports whether path is in the pattern's set, recording wildcard
// values into values unless it is nil.
func (p PathPattern) match(path BroadcastPath, values map[string]string) bool {
	if p.segs != nil {
		return p.matchSegments(path, values)
	}
	if isValidPrefix(p.raw) {
		return path.HasPrefix(p.raw)
	}
	return p.raw != "" && string(path) == p.raw
}

// matchSegments matches path against a pattern with a wildcard.
func (p PathPattern) matchSegments(path BroadcastPath, values map[string]string) bool {
	if !isValidPath(path) {
		return false
	}
	prefix := strings.HasSuffix(p.raw, "/")
	rest := string(path[1:])
	for i, seg := range p.segs {
		if seg == anchorSegment {
			return rest == ""
		}
		if name, ok := strings.CutSuffix(seg, "...}"); ok {
			if values != nil {
				values[name[1:]] = rest
			}
			return true
		}
		got, after, more := strings.Cut(rest, "/")
		if strings.HasPrefix(seg, "{") {
			if got == "" {
				return false
			}
			if values != nil {
				values[seg[1:len(seg)-1]] = got
			}
		} else if got != seg {
			return false
		}
		if !more {
			return i == len(p.segs)-1 && !prefix
		}
		rest = after
	}
	return prefix
}

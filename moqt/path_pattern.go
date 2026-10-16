package moqt

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// ErrInvalidPathPattern is returned for a path pattern that does not follow
// the syntax documented on Grant: one that is not a broadcast path or a prefix
// of them, or whose wildcard is malformed, misplaced or named twice. Received
// on an Auth Stream, it is a protocol violation.
var ErrInvalidPathPattern = errors.New("moqt: invalid path pattern")

// PathPattern is a parsed path pattern, a set of broadcast paths; Grant
// documents the syntax. An application that grants by the same rules as a
// session enforces, such as for requests outside MoQ, matches with it.
type PathPattern struct {
	raw string
	// segs holds the segments of a pattern with a wildcard, and is nil for
	// one without.
	segs []patternSeg
	// trailing is whether a pattern with a wildcard ends in "/".
	trailing bool
}

// patternSeg is a segment of a pattern with a wildcard.
type patternSeg struct {
	kind segKind
	// lit is the text of a literal segment, and empty for a wildcard.
	lit string
}

// ParsePathPattern parses s, an absolute pattern: a path starting with "/"
// whose wildcards are whole segments. An empty segment, as in "/room//x", is
// a literal one. ParseGrantPattern reads one relative to a root, as a Grant's
// are.
func ParsePathPattern(s string) (PathPattern, error) {
	if !isValidPath(BroadcastPath(s)) {
		return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
	}
	if !strings.ContainsAny(s, "{}") {
		return PathPattern{raw: s}, nil
	}

	p := PathPattern{raw: s, trailing: strings.HasSuffix(s, "/")}
	var names []string
	for segs := strings.TrimSuffix(s[1:], "/"); ; {
		seg, next, more := strings.Cut(segs, "/")
		name, kind := parseWildcard(seg)
		last := !more && !p.trailing
		switch {
		case kind == segInvalid,
			kind == segAnchor && !last,
			kind == segRest && !last,
			name != "" && slices.Contains(names, name):
			return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
		}
		if name != "" {
			names = append(names, name)
		}
		ps := patternSeg{kind: kind}
		if kind == segLiteral {
			ps.lit = seg
		}
		p.segs = append(p.segs, ps)
		if !more {
			return p, nil
		}
		segs = next
	}
}

// segKind is the kind of a pattern segment.
type segKind int

const (
	segInvalid  segKind = iota
	segLiteral          // a literal segment
	segWildcard         // "{name}"
	segRest             // "{name...}"
	segAnchor           // "{$}"
)

// parseWildcard reads a segment: its kind, and the wildcard's name.
func parseWildcard(seg string) (name string, kind segKind) {
	inner, ok := strings.CutPrefix(seg, "{")
	if !ok {
		if strings.ContainsAny(seg, "{}") {
			return "", segInvalid
		}
		return "", segLiteral
	}
	if inner, ok = strings.CutSuffix(inner, "}"); !ok {
		return "", segInvalid
	}
	if inner == "$" {
		return "", segAnchor
	}
	kind = segWildcard
	if name, ok = strings.CutSuffix(inner, "..."); ok {
		kind = segRest
	} else {
		name = inner
	}
	if !isWildcardName(name) {
		return "", segInvalid
	}
	return name, kind
}

// isWildcardName reports whether s is a wildcard name as net/http reads it:
// letters, digits and "_", not starting with a digit.
func isWildcardName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}

// Match reports whether path is in the pattern's set. The zero PathPattern
// matches nothing.
func (p PathPattern) Match(path BroadcastPath) bool {
	if p.segs != nil {
		return p.matchSegments(path)
	}
	if isValidPrefix(p.raw) {
		return path.HasPrefix(p.raw)
	}
	return p.raw != "" && string(path) == p.raw
}

// matchSegments matches path against a pattern with a wildcard.
func (p PathPattern) matchSegments(path BroadcastPath) bool {
	if !isValidPath(path) {
		return false
	}
	rest := string(path[1:])
	for i, seg := range p.segs {
		switch seg.kind {
		case segAnchor:
			return rest == ""
		case segRest:
			return true
		}
		got, after, more := strings.Cut(rest, "/")
		if seg.kind == segWildcard {
			if got == "" {
				return false
			}
		} else if got != seg.lit {
			return false
		}
		if i == len(p.segs)-1 {
			// The pattern's segments are used up: an exact pattern needs the
			// path to end here, and a prefix needs it to go on.
			return more == p.trailing
		}
		if !more {
			return false
		}
		rest = after
	}
	return false
}

// String returns the pattern as it was parsed.
func (p PathPattern) String() string {
	return p.raw
}

// ParseGrantPattern parses pattern relative to root, as a Grant's patterns
// are read: with root "/room/1/", "/alice" is "/room/1/alice". root must be a
// prefix, ending in "/", and pattern start with "/".
func ParseGrantPattern(root, pattern string) (PathPattern, error) {
	if !isValidPrefix(root) || strings.ContainsAny(root, "{}") || !strings.HasPrefix(pattern, "/") {
		return PathPattern{}, fmt.Errorf("%w: %q under %q", ErrInvalidPathPattern, pattern, root)
	}
	return ParsePathPattern(strings.TrimSuffix(root, "/") + pattern)
}

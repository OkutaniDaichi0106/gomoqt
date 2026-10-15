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

// pathPattern is a parsed path pattern; Grant documents the syntax.
type pathPattern struct {
	raw string
	// segs holds the segments of a pattern with a wildcard, and is nil for
	// one without.
	segs []patternSeg
	// trailing is whether the pattern ends in "/".
	trailing bool
}

// patternSeg is a segment of a pattern with a wildcard.
type patternSeg struct {
	kind int
	// lit is the text of a literal segment.
	lit string
}

// parsePathPattern parses s: a path starting with "/" whose wildcards are
// whole segments. An empty segment, as in "/room//x", is a literal one.
func parsePathPattern(s string) (pathPattern, error) {
	if !isValidPath(BroadcastPath(s)) {
		return pathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
	}
	if !strings.ContainsAny(s, "{}") {
		return pathPattern{raw: s}, nil
	}

	p := pathPattern{raw: s, trailing: strings.HasSuffix(s, "/")}
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
			return pathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
		}
		if name != "" {
			names = append(names, name)
		}
		p.segs = append(p.segs, patternSeg{kind: kind, lit: seg})
		if !more {
			return p, nil
		}
		segs = next
	}
}

// Kinds of pattern segment.
const (
	segInvalid  = iota
	segLiteral  // a literal segment
	segWildcard // "{name}"
	segRest     // "{name...}"
	segAnchor   // "{$}"
)

// parseWildcard reads a segment: its kind, and the wildcard's name.
func parseWildcard(seg string) (name string, kind int) {
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

// match reports whether path is in the pattern's set. The zero pathPattern
// matches nothing.
func (p pathPattern) match(path BroadcastPath) bool {
	if p.segs != nil {
		return p.matchSegments(path)
	}
	if isValidPrefix(p.raw) {
		return path.HasPrefix(p.raw)
	}
	return p.raw != "" && string(path) == p.raw
}

// matchSegments matches path against a pattern with a wildcard.
func (p pathPattern) matchSegments(path BroadcastPath) bool {
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

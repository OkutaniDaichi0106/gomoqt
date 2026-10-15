package moqt

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// ErrInvalidPathPattern is returned for a path pattern that is not a
// broadcast path or a prefix of them. Received on an Auth Stream, it is a
// protocol violation.
var ErrInvalidPathPattern = errors.New("moqt: invalid path pattern")

// pathPattern is a parsed path pattern; Grant documents the syntax.
type pathPattern struct {
	raw string
	// wild is whether a segment is a wildcard.
	wild bool
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

	trailing := strings.HasSuffix(s, "/")
	var names []string
	for segs := strings.TrimSuffix(s[1:], "/"); ; {
		seg, next, more := strings.Cut(segs, "/")
		if strings.ContainsAny(seg, "{}") {
			name, kind := parseWildcard(seg)
			last := !more && !trailing
			switch {
			case kind == segInvalid,
				kind == segAnchor && !last,
				kind == segRest && !last,
				slices.Contains(names, name):
				return pathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
			}
			if name != "" {
				names = append(names, name)
			}
		}
		if !more {
			return pathPattern{raw: s, wild: true}, nil
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
	if p.wild {
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
	trailing := strings.HasSuffix(p.raw, "/")
	rest := string(path[1:])
	for segs := strings.TrimSuffix(p.raw[1:], "/"); ; {
		seg, nextSegs, moreSegs := strings.Cut(segs, "/")
		_, kind := parseWildcard(seg)
		switch kind {
		case segAnchor:
			return rest == ""
		case segRest:
			return true
		}
		got, after, more := strings.Cut(rest, "/")
		if kind == segWildcard {
			if got == "" {
				return false
			}
		} else if got != seg {
			return false
		}
		if !moreSegs {
			// The pattern's segments are used up: an exact pattern needs the
			// path to end here, and a prefix needs it to go on.
			return more == trailing
		}
		if !more {
			return false
		}
		segs, rest = nextSegs, after
	}
}

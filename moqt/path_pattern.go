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
// "/room/{$}" matches "/room/" alone. A name is letters, digits and "_", not
// starting with a digit, as in net/http, and unique within a pattern.
//
// PathPatterns are comparable: two are equal when they were parsed from the
// same string.
type PathPattern struct {
	raw string
	// wild is whether a segment is a wildcard.
	wild bool
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
				return PathPattern{}, fmt.Errorf("%w: %q", ErrInvalidPathPattern, s)
			}
			if name != "" {
				names = append(names, name)
			}
		}
		if !more {
			return PathPattern{raw: s, wild: true}, nil
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
	if p.wild {
		return p.matchSegments(path, nil)
	}
	if isValidPrefix(p.raw) {
		return path.HasPrefix(p.raw)
	}
	return p.raw != "" && string(path) == p.raw
}

// Values reports whether path is in the pattern's set and, if it is, returns
// the value each named wildcard matched, like http.Request.PathValue: for
// "/room/{id}/{rest...}", the path "/room/1/a/b" gives id "1" and rest "a/b".
// A pattern without a named wildcard gives a nil map.
func (p PathPattern) Values(path BroadcastPath) (map[string]string, bool) {
	if !p.wild {
		return nil, p.Match(path)
	}
	var values map[string]string
	if !p.matchSegments(path, &values) {
		return nil, false
	}
	return values, true
}

// matchSegments matches path against a pattern with a wildcard, recording
// what each named wildcard matched into *values unless values is nil.
func (p PathPattern) matchSegments(path BroadcastPath, values *map[string]string) bool {
	if !isValidPath(path) {
		return false
	}
	record := func(name, value string) {
		if values == nil {
			return
		}
		if *values == nil {
			*values = make(map[string]string)
		}
		(*values)[name] = value
	}

	trailing := strings.HasSuffix(p.raw, "/")
	rest := string(path[1:])
	for segs := strings.TrimSuffix(p.raw[1:], "/"); ; {
		seg, nextSegs, moreSegs := strings.Cut(segs, "/")
		name, kind := parseWildcard(seg)
		switch kind {
		case segAnchor:
			return rest == ""
		case segRest:
			record(name, rest)
			return true
		}
		got, after, more := strings.Cut(rest, "/")
		if kind == segWildcard {
			if got == "" {
				return false
			}
			record(name, got)
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

// patternRelation is how the sets of paths two patterns match relate, as
// http.ServeMux relates its patterns.
type patternRelation int

const (
	relEquivalent   patternRelation = iota // the same set
	relMoreSpecific                        // a strict subset
	relMoreGeneral                         // a strict superset
	relOverlaps                            // neither, but they share a path
	relDisjoint                            // they share no path
)

// compare relates the set of paths p matches to the set q matches.
func (p PathPattern) compare(q PathPattern) patternRelation {
	a, b := p.elems(), q.elems()
	rel := relEquivalent
	for i := 0; ; i++ {
		if i == len(a) || i == len(b) {
			if len(a) == len(b) {
				return rel
			}
			return relDisjoint
		}
		ea, eb := a[i], b[i]
		switch {
		case ea.multi && eb.multi:
			return rel
		case ea.multi:
			return combineRelations(rel, relMoreGeneral)
		case eb.multi:
			return combineRelations(rel, relMoreSpecific)
		}
		if rel = combineRelations(rel, ea.compare(eb)); rel == relDisjoint {
			return rel
		}
	}
}

// combineRelations relates two patterns from how two consecutive parts of
// them relate.
func combineRelations(r1, r2 patternRelation) patternRelation {
	switch {
	case r1 == relDisjoint || r2 == relDisjoint:
		return relDisjoint
	case r1 == relEquivalent:
		return r2
	case r2 == relEquivalent || r1 == r2:
		return r1
	default:
		return relOverlaps
	}
}

// patternElem is one segment of a pattern: a literal, a "{name}" wildcard,
// or the rest of a path ("{name...}" or what a trailing "/" leaves).
type patternElem struct {
	literal  string
	wildcard bool
	multi    bool
}

// compare relates the segments e and f match.
func (e patternElem) compare(f patternElem) patternRelation {
	switch {
	case e.wildcard && f.wildcard:
		return relEquivalent
	case e.wildcard:
		// "{name}" matches no empty segment.
		if f.literal == "" {
			return relDisjoint
		}
		return relMoreGeneral
	case f.wildcard:
		if e.literal == "" {
			return relDisjoint
		}
		return relMoreSpecific
	case e.literal == f.literal:
		return relEquivalent
	default:
		return relDisjoint
	}
}

// elems returns the segments of p. "{$}" is the empty literal segment it
// matches.
func (p PathPattern) elems() []patternElem {
	if p.raw == "" {
		return nil
	}
	var elems []patternElem
	if p.raw != "/" {
		for _, seg := range strings.Split(strings.TrimSuffix(p.raw[1:], "/"), "/") {
			switch _, kind := parseWildcard(seg); kind {
			case segAnchor:
				elems = append(elems, patternElem{})
			case segRest:
				return append(elems, patternElem{multi: true})
			case segWildcard:
				elems = append(elems, patternElem{wildcard: true})
			default:
				elems = append(elems, patternElem{literal: seg})
			}
		}
	}
	if strings.HasSuffix(p.raw, "/") {
		elems = append(elems, patternElem{multi: true})
	}
	return elems
}

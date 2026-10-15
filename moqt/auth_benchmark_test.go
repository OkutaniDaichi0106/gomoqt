package moqt

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// benchPatterns are grant patterns of each kind, with a path each one matches
// and one it does not.
var benchPatterns = []struct {
	name, pattern string
	hit, miss     BroadcastPath
}{
	{"exact", "/room/123/alice", "/room/123/alice", "/room/123/bob"},
	{"prefix", "/room/123/", "/room/123/alice", "/room/124/alice"},
	{"wildcard", "/room/{id}/host", "/room/123/host", "/room/123/guest"},
	{"rest", "/room/{id}/{rest...}", "/room/123/a/b", "/lobby/123/a"},
}

func BenchmarkParsePathPattern(b *testing.B) {
	for _, bp := range benchPatterns {
		b.Run(bp.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := parsePathPattern(bp.pattern); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPathPattern_match(b *testing.B) {
	for _, bp := range benchPatterns {
		p, err := parsePathPattern(bp.pattern)
		if err != nil {
			b.Fatal(err)
		}
		for _, c := range []struct {
			name string
			path BroadcastPath
			want bool
		}{{"hit", bp.hit, true}, {"miss", bp.miss, false}} {
			b.Run(bp.name+"/"+c.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if p.match(c.path) != c.want {
						b.Fatal("unexpected match result")
					}
				}
			})
		}
	}
}

// benchScope returns a scope holding one grant per room, each a prefix like
// the ticket of one room, or a wildcard pattern shared by every room.
func benchScope(b *testing.B, grants int, pattern func(i int) string) *authScope {
	b.Helper()
	scope := newAuthScope(func() {})
	for i := range grants {
		w := &AuthWriter{scope: scope}
		if err := w.Grant(Grant{Subscribe: []string{pattern(i)}}); err != nil {
			b.Fatal(err)
		}
	}
	return scope
}

func roomPrefix(i int) string { return fmt.Sprintf("/room/%d/", i) }

func roomWildcard(i int) string { return fmt.Sprintf("/room/{id}/host%d", i) }

// BenchmarkAuthScope_canSubscribe checks a path no grant covers, so every
// grant is scanned: the cost of a refused request, and the upper bound of an
// allowed one.
func BenchmarkAuthScope_canSubscribe(b *testing.B) {
	for _, kind := range []struct {
		name    string
		pattern func(int) string
	}{{"prefix", roomPrefix}, {"wildcard", roomWildcard}} {
		for _, n := range []int{1, 10, 100, 1000} {
			b.Run(fmt.Sprintf("%s/grants=%d", kind.name, n), func(b *testing.B) {
				scope := benchScope(b, n, kind.pattern)
				b.ReportAllocs()
				for b.Loop() {
					if scope.canSubscribe("/lobby/1/host") {
						b.Fatal("unexpected grant")
					}
				}
			})
		}
	}
}

func BenchmarkAuthScope_canSubscribe_Parallel(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("grants=%d", n), func(b *testing.B) {
			scope := benchScope(b, n, roomPrefix)
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if !scope.canSubscribe("/room/0/host") {
						b.Fatal("missing grant")
					}
				}
			})
		})
	}
}

func BenchmarkAuthWriter_Grant(b *testing.B) {
	grants := map[string]Grant{
		"prefix":   {Subscribe: []string{"/room/123/"}, Publish: []string{"/room/123/alice"}},
		"wildcard": {Subscribe: []string{"/room/{id}/host"}, Publish: []string{"/room/{id}/{rest...}"}},
	}
	for name, g := range grants {
		// Each grant replaces one with other patterns, so the scope changes.
		b.Run(name+"/change", func(b *testing.B) {
			w := &AuthWriter{scope: newAuthScope(func() {})}
			other := Grant{Subscribe: []string{"/other/"}}
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				next := g
				if i%2 == 1 {
					next = other
				}
				if err := w.Grant(next); err != nil {
					b.Fatal(err)
				}
			}
		})
		// The same grant sent again with a max age, as a refresh does.
		b.Run(name+"/extend", func(b *testing.B) {
			w := &AuthWriter{scope: newAuthScope(func() {})}
			g := g
			g.MaxAge = time.Minute
			b.ReportAllocs()
			for b.Loop() {
				if err := w.Grant(g); err != nil {
					b.Fatal(err)
				}
			}
			w.mu.Lock()
			w.end()
			w.mu.Unlock()
		})
	}
}

// BenchmarkAnnouncementWriter_rescope measures a pass over an announce stream
// after the scope changed in a way that leaves every announcement allowed, as
// when a grant for another room is added: each announcement is checked
// against every grant, and nothing is sent.
func BenchmarkAnnouncementWriter_rescope(b *testing.B) {
	for _, anns := range []int{10, 100, 1000} {
		for _, grants := range []int{1, 10} {
			b.Run(fmt.Sprintf("announcements=%d/grants=%d", anns, grants), func(b *testing.B) {
				// Rooms 0..grants-1 are granted, and every announcement is under
				// the last one, so each check scans every grant.
				scope := benchScope(b, grants, roomPrefix)
				aw := newAnnouncementWriter(&FakeQUICStream{}, "/", 0, 0, nil)
				aw.allow = scope.canSubscribe
				close(aw.initDone)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				for i := range anns {
					ann, _ := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/room/%d/user%d", grants-1, i)))
					aw.actives[suffix(fmt.Sprintf("room/%d/user%d", grants-1, i))] = &activeAnnouncement{
						announcement: ann,
						stop:         func() bool { return true },
					}
				}
				b.ReportAllocs()
				for b.Loop() {
					aw.rescope()
				}
			})
		}
	}
}

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

// benchScope returns a scope holding one grant per i, each with the pattern
// pattern(i) returns.
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

// roomWildcard returns a distinct wildcard pattern for each i.
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

// BenchmarkAuthScope_canSubscribe_Parallel checks the same refused path as
// BenchmarkAuthScope_canSubscribe from every core at once, as the requests of
// one session may.
func BenchmarkAuthScope_canSubscribe_Parallel(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("prefix/grants=%d", n), func(b *testing.B) {
			scope := benchScope(b, n, roomPrefix)
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if scope.canSubscribe("/lobby/1/host") {
						b.Error("unexpected grant")
						return
					}
				}
			})
		})
	}
}

func BenchmarkAuthWriter_Grant(b *testing.B) {
	grants := []struct {
		name string
		// a and b are grants of the same kind with other patterns.
		a, b Grant
	}{
		{
			"prefix",
			Grant{Subscribe: []string{"/room/123/"}, Publish: []string{"/room/123/alice"}},
			Grant{Subscribe: []string{"/room/124/"}, Publish: []string{"/room/124/alice"}},
		},
		{
			"wildcard",
			Grant{Subscribe: []string{"/room/{id}/host"}, Publish: []string{"/room/{id}/{rest...}"}},
			Grant{Subscribe: []string{"/room/{id}/guest"}, Publish: []string{"/room/{id}/{tail...}"}},
		},
	}
	for _, gc := range grants {
		// Each grant replaces one with other patterns, so the scope changes.
		b.Run(gc.name+"/change", func(b *testing.B) {
			w := &AuthWriter{scope: newAuthScope(func() {})}
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				next := gc.a
				if i%2 == 1 {
					next = gc.b
				}
				if err := w.Grant(next); err != nil {
					b.Fatal(err)
				}
			}
		})
		// The same grant sent again with a max age, as a refresh does.
		b.Run(gc.name+"/extend", func(b *testing.B) {
			w := &AuthWriter{scope: newAuthScope(func() {})}
			ext := gc.a
			ext.MaxAge = time.Minute
			b.ReportAllocs()
			for b.Loop() {
				if err := w.Grant(ext); err != nil {
					b.Fatal(err)
				}
			}
			w.end()
		})
	}
}

// BenchmarkAnnouncementWriter_rescope measures a pass over an announce stream
// after the scope changed in a way that changes nothing on it, as when a grant
// for another room is added: each active announcement stays allowed and each
// held-back one stays refused, so each is checked against every grant and
// nothing is sent. The writer's state is set directly, as init and
// SendAnnouncement would leave it, to keep the stream out of the measurement.
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
				aw.mu.Lock()
				// Half are active and allowed, half held back and refused.
				for i := range anns / 2 {
					ann, _ := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/room/%d/user%d", grants-1, i)))
					aw.actives[suffix(fmt.Sprintf("room/%d/user%d", grants-1, i))] = &activeAnnouncement{
						announcement: ann,
						stop:         func() bool { return true },
					}
					held, _ := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/lobby/user%d", i)))
					aw.hideLocked(held)
				}
				aw.mu.Unlock()
				b.ReportAllocs()
				for b.Loop() {
					aw.rescope()
				}
			})
		}
	}
}

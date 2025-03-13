package moqt

import (
	"testing"
)

type dummyHandler struct{}

func (d *dummyHandler) ServeTrack(w TrackWriter, r SubscribeConfig) {}
func (d *dummyHandler) ServeAnnouncement(w AnnouncementWriter)      {}
func (d *dummyHandler) ServeInfo(ch chan<- Info, r InfoRequest)     {}

func TestFindHandler(t *testing.T) {
	mux := NewTrackMux()
	handler := &dummyHandler{}

	// Register handlers for specific paths
	mux.Handle("/a/b/c/d", handler)
	mux.Handle("/a/b", handler)
	mux.Handle("/x/y/z", handler)

	// Test cases with correct glob pattern expectations
	testCases := []struct {
		pattern     string // The pattern to search for
		targetPath  string // Path we're trying to match
		shouldMatch bool   // Whether we expect a match
	}{
		// Basic exact matches
		{"/a/b/c/d", "/a/b/c/d", true},
		{"/a/b", "/a/b", true},
		{"/x/y/z", "/x/y/z", true},

		// Single wildcard matches
		{"/a/*/c/d", "/a/b/c/d", true}, // * matches 'b'
		{"/a/b/*/d", "/a/b/c/d", true}, // * matches 'c'
		{"/a/b/*/*", "/a/b/c/d", true}, // * matches 'c' and 'd'
		{"/*/b/c/d", "/a/b/c/d", true}, // * matches 'a'

		// Double wildcard matches
		{"/**/d", "/a/b/c/d", true},   // ** matches 'a/b/c'
		{"/a/**", "/a/b/c/d", true},   // ** matches 'b/c/d'
		{"/a/**/d", "/a/b/c/d", true}, // ** matches 'b/c'

		// Non-matches
		{"/b/*", "/a/b", false},       // Different base path
		{"/a/*/d", "/a/b/c/d", false}, // Missing segment
		{"/a/b/c", "/a/b/c/d", false}, // Path too short
		{"/*", "/a/b", false},         // Single * only matches one segment
	}

	for _, tc := range testCases {
		t.Run(tc.pattern+"→"+tc.targetPath, func(t *testing.T) {
			// Create pattern from the target path we're trying to match
			// targetPattern := newPattern(tc.targetPath)

			// Find the node using the glob pattern
			node := mux.findRoutingNode(newPattern(tc.pattern))

			if tc.shouldMatch {
				if node == nil || node.handler != handler {
					t.Errorf("Pattern %q should match path %q but did not", tc.pattern, tc.targetPath)
				} else {
					t.Logf("Pattern %q correctly matched path %q", tc.pattern, tc.targetPath)
				}
			} else {
				if node != nil && node.handler == handler {
					t.Errorf("Pattern %q should NOT match path %q but did", tc.pattern, tc.targetPath)
				} else {
					t.Logf("Pattern %q correctly did not match path %q", tc.pattern, tc.targetPath)
				}
			}

			// Also test the reverse: direct glob matching function
			if matchGlob(tc.pattern, tc.targetPath) != tc.shouldMatch {
				t.Errorf("matchGlob(%q, %q) returned %v, expected %v",
					tc.pattern, tc.targetPath, !tc.shouldMatch, tc.shouldMatch)
			}
		})
	}
}

package moqt

import (
	"testing"
)

type dummyHandler struct{}

func (d *dummyHandler) ServeTrack(w TrackWriter, r SubscribeConfig)              {}
func (d *dummyHandler) ServeAnnouncement(w AnnouncementWriter, r AnnounceConfig) {}
func (d *dummyHandler) ServeInfo(ch chan<- Info, r InfoRequest)                  {}

func TestFindHandler(t *testing.T) {
	mux := NewServeMux()
	dummy := &dummyHandler{}

	// Register "/a/b/c/d"
	mux.Handle("/a/b/c/d", dummy)
	// Manually set dummy on the leaf node corresponding to "/a/b/c/d"
	{
		// Traverse: parts => ["", "a", "b", "c", "d"]
		node, ok := mux.tree.children[""]
		if !ok {
			t.Fatal("root empty node not found")
		}
		node, ok = node.children["a"]
		if !ok {
			t.Fatal("node 'a' not found")
		}
		node, ok = node.children["b"]
		if !ok {
			t.Fatal("node 'b' not found")
		}
		node, ok = node.children["c"]
		if !ok {
			t.Fatal("node 'c' not found")
		}
		node, ok = node.children["d"]
		if !ok {
			t.Fatal("node 'd' not found")
		}
		node.handler = dummy
	}

	// Test patterns to match against "/a/b/c/d"
	testCases := []struct {
		pattern  string
		expected bool
	}{
		{"*/c/d", true},
		{"*/d", true},
		{"/a/*", true},
		{"*/b/*", true},
		{"*/b/c/*", true},
	}

	for _, tc := range testCases {
		h := mux.findRoutingNode(newPattern(tc.pattern)).handler
		if (h == dummy) != tc.expected {
			t.Errorf("pattern %q: expected match %v, got %v", tc.pattern, tc.expected, h)
		} else {
			t.Logf("pattern %q matched as expected", tc.pattern)
		}
	}
}

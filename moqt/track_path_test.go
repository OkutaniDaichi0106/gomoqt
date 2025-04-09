package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestTrackPath_String(t *testing.T) {
	tests := map[string]struct {
		path     moqt.TrackPath
		expected string
	}{
		"empty path": {
			path:     moqt.TrackPath(""),
			expected: "",
		},
		"simple path": {
			path:     moqt.TrackPath("/test/path"),
			expected: "/test/path",
		},
		"complex path": {
			path:     moqt.TrackPath("/test/path/with/multiple/segments"),
			expected: "/test/path/with/multiple/segments",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.path.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackPath_HasPrefix(t *testing.T) {
	tests := map[string]struct {
		path     moqt.TrackPath
		prefix   string
		expected bool
	}{
		"empty path and prefix": {
			path:     moqt.TrackPath(""),
			prefix:   "",
			expected: false,
		},
		"path shorter than prefix": {
			path:     moqt.TrackPath("/test"),
			prefix:   "/test/path",
			expected: false,
		},
		"matching prefix": {
			path:     moqt.TrackPath("/test/path/segment"),
			prefix:   "/test",
			expected: true,
		},
		"non-matching prefix": {
			path:     moqt.TrackPath("/test/path"),
			prefix:   "/other",
			expected: false,
		},
		"prefix without trailing slash": {
			path:     moqt.TrackPath("/test/path"),
			prefix:   "/test",
			expected: true,
		},
		"exact match is not a prefix": {
			path:     moqt.TrackPath("/test/path"),
			prefix:   "/test/path",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.path.HasPrefix(tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackPath_HasSuffix(t *testing.T) {
	tests := map[string]struct {
		path     moqt.TrackPath
		suffix   string
		expected bool
	}{
		"empty path and suffix": {
			path:     moqt.TrackPath(""),
			suffix:   "",
			expected: false,
		},
		"path shorter than suffix": {
			path:     moqt.TrackPath("/test"),
			suffix:   "/test/path",
			expected: false,
		},
		"matching suffix": {
			path:     moqt.TrackPath("/segment/test/path"),
			suffix:   "path",
			expected: true,
		},
		"non-matching suffix": {
			path:     moqt.TrackPath("/test/path"),
			suffix:   "other",
			expected: false,
		},
		"suffix without leading slash": {
			path:     moqt.TrackPath("/test/path"),
			suffix:   "path",
			expected: true,
		},
		"exact match is not a suffix": {
			path:     moqt.TrackPath("/test/path"),
			suffix:   "/test/path",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.path.HasSuffix(tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackPath_Equal(t *testing.T) {
	tests := map[string]struct {
		path     moqt.TrackPath
		target   moqt.TrackPath
		expected bool
	}{
		"empty paths": {
			path:     moqt.TrackPath(""),
			target:   moqt.TrackPath(""),
			expected: true,
		},
		"identical paths": {
			path:     moqt.TrackPath("/test/path"),
			target:   moqt.TrackPath("/test/path"),
			expected: true,
		},
		"different paths": {
			path:     moqt.TrackPath("/test/path1"),
			target:   moqt.TrackPath("/test/path2"),
			expected: false,
		},
		"case sensitive comparison": {
			path:     moqt.TrackPath("/Test/Path"),
			target:   moqt.TrackPath("/test/path"),
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.path.Equal(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackPath_Match(t *testing.T) {
	tests := map[string]struct {
		path     moqt.TrackPath
		pattern  string
		expected bool
	}{
		"empty path and pattern": {
			path:     moqt.TrackPath(""),
			pattern:  "",
			expected: true,
		},
		"exact match": {
			path:     moqt.TrackPath("/test/path"),
			pattern:  "/test/path",
			expected: true,
		},
		"single wildcard": {
			path:     moqt.TrackPath("/test/path"),
			pattern:  "/test/*",
			expected: true,
		},
		"double wildcard": {
			path:     moqt.TrackPath("/test/path/segment"),
			pattern:  "/test/**",
			expected: true,
		},
		"double wildcard in middle": {
			path:     moqt.TrackPath("/test/path/segment/end"),
			pattern:  "/test/**/end",
			expected: true,
		},
		"non-matching pattern": {
			path:     moqt.TrackPath("/test/path"),
			pattern:  "/other/*",
			expected: false,
		},
		"complex pattern match": {
			path:     moqt.TrackPath("/test/path/segment/end"),
			pattern:  "/test/*/segment/*",
			expected: true,
		},
		"complex pattern non-match": {
			path:     moqt.TrackPath("/test/path/wrong/end"),
			pattern:  "/test/*/segment/*",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.path.Match(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

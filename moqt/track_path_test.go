package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestTrackPath_String(t *testing.T) {
	tests := map[string]struct {
		path     moqt.BroadcastPath
		expected string
	}{
		"empty path": {
			path:     moqt.BroadcastPath(""),
			expected: "",
		},
		"simple path": {
			path:     moqt.BroadcastPath("/test/path"),
			expected: "/test/path",
		},
		"complex path": {
			path:     moqt.BroadcastPath("/test/path/with/multiple/segments"),
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
		path     moqt.BroadcastPath
		prefix   string
		expected bool
	}{
		"empty path and prefix": {
			path:     moqt.BroadcastPath(""),
			prefix:   "",
			expected: false,
		},
		"path shorter than prefix": {
			path:     moqt.BroadcastPath("/test"),
			prefix:   "/test/path",
			expected: false,
		},
		"matching prefix": {
			path:     moqt.BroadcastPath("/test/path/segment"),
			prefix:   "/test",
			expected: true,
		},
		"non-matching prefix": {
			path:     moqt.BroadcastPath("/test/path"),
			prefix:   "/other",
			expected: false,
		},
		"prefix without trailing slash": {
			path:     moqt.BroadcastPath("/test/path"),
			prefix:   "/test",
			expected: true,
		},
		"exact match is not a prefix": {
			path:     moqt.BroadcastPath("/test/path"),
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
		path     moqt.BroadcastPath
		suffix   string
		expected bool
	}{
		"empty path and suffix": {
			path:     moqt.BroadcastPath(""),
			suffix:   "",
			expected: false,
		},
		"path shorter than suffix": {
			path:     moqt.BroadcastPath("/test"),
			suffix:   "/test/path",
			expected: false,
		},
		"matching suffix": {
			path:     moqt.BroadcastPath("/segment/test/path"),
			suffix:   "path",
			expected: true,
		},
		"non-matching suffix": {
			path:     moqt.BroadcastPath("/test/path"),
			suffix:   "other",
			expected: false,
		},
		"suffix without leading slash": {
			path:     moqt.BroadcastPath("/test/path"),
			suffix:   "path",
			expected: true,
		},
		"exact match is not a suffix": {
			path:     moqt.BroadcastPath("/test/path"),
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
		path     moqt.BroadcastPath
		target   moqt.BroadcastPath
		expected bool
	}{
		"empty paths": {
			path:     moqt.BroadcastPath(""),
			target:   moqt.BroadcastPath(""),
			expected: true,
		},
		"identical paths": {
			path:     moqt.BroadcastPath("/test/path"),
			target:   moqt.BroadcastPath("/test/path"),
			expected: true,
		},
		"different paths": {
			path:     moqt.BroadcastPath("/test/path1"),
			target:   moqt.BroadcastPath("/test/path2"),
			expected: false,
		},
		"case sensitive comparison": {
			path:     moqt.BroadcastPath("/Test/Path"),
			target:   moqt.BroadcastPath("/test/path"),
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
		path     moqt.BroadcastPath
		pattern  string
		expected bool
	}{
		"empty path and pattern": {
			path:     moqt.BroadcastPath(""),
			pattern:  "",
			expected: true,
		},
		"exact match": {
			path:     moqt.BroadcastPath("/test/path"),
			pattern:  "/test/path",
			expected: true,
		},
		"single wildcard": {
			path:     moqt.BroadcastPath("/test/path"),
			pattern:  "/test/*",
			expected: true,
		},
		"double wildcard": {
			path:     moqt.BroadcastPath("/test/path/segment"),
			pattern:  "/test/**",
			expected: true,
		},
		"double wildcard in middle": {
			path:     moqt.BroadcastPath("/test/path/segment/end"),
			pattern:  "/test/**/end",
			expected: true,
		},
		"non-matching pattern": {
			path:     moqt.BroadcastPath("/test/path"),
			pattern:  "/other/*",
			expected: false,
		},
		"complex pattern match": {
			path:     moqt.BroadcastPath("/test/path/segment/end"),
			pattern:  "/test/*/segment/*",
			expected: true,
		},
		"complex pattern non-match": {
			path:     moqt.BroadcastPath("/test/path/wrong/end"),
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

func TestTrackPath_ExtractParameters(t *testing.T) {
	tests := map[string]struct {
		path           moqt.BroadcastPath
		pattern        string
		expectedParams []string
		expectedMatch  bool
	}{
		"empty path and pattern": {
			path:           moqt.BroadcastPath(""),
			pattern:        "",
			expectedParams: nil,
			expectedMatch:  false,
		},
		"pattern with no wildcards": {
			path:           moqt.BroadcastPath("/test/path"),
			pattern:        "/test/path",
			expectedParams: nil,
			expectedMatch:  false,
		},
		"single wildcard - successful match": {
			path:           moqt.BroadcastPath("/test/value"),
			pattern:        "/test/*",
			expectedParams: []string{"value"},
			expectedMatch:  true,
		},
		"single wildcard - no match": {
			path:           moqt.BroadcastPath("/other/value"),
			pattern:        "/test/*",
			expectedParams: nil,
			expectedMatch:  false,
		},
		"multiple single wildcards": {
			path:           moqt.BroadcastPath("/users/123/posts/456"),
			pattern:        "/users/*/posts/*",
			expectedParams: []string{"123", "456"},
			expectedMatch:  true,
		},
		"double wildcard - successful match": {
			path:           moqt.BroadcastPath("/api/v1/users/123/profile"),
			pattern:        "/api/**/profile",
			expectedParams: []string{"v1/users/123"},
			expectedMatch:  true,
		},
		"double wildcard - no match": {
			path:           moqt.BroadcastPath("/api/v1/users/123/settings"),
			pattern:        "/api/**/profile",
			expectedParams: nil,
			expectedMatch:  false,
		},
		"mixed wildcards": {
			path:           moqt.BroadcastPath("/api/v1/users/123/posts/456/comments"),
			pattern:        "/api/*/users/**/comments",
			expectedParams: []string{"v1", "123/posts/456"},
			expectedMatch:  true,
		},
		"wildcard at beginning": {
			path:           moqt.BroadcastPath("/tenant1/resources"),
			pattern:        "/*/resources",
			expectedParams: []string{"tenant1"},
			expectedMatch:  true,
		},
		"wildcard in middle": {
			path:           moqt.BroadcastPath("/api/tenant1/resources"),
			pattern:        "/api/*/resources",
			expectedParams: []string{"tenant1"},
			expectedMatch:  true,
		},
		"wildcard at end": {
			path:           moqt.BroadcastPath("/api/resources/latest"),
			pattern:        "/api/resources/*",
			expectedParams: []string{"latest"},
			expectedMatch:  true,
		},
		"empty segment in path": {
			path:           moqt.BroadcastPath("/api//resources"),
			pattern:        "/api/*/resources",
			expectedParams: []string{""},
			expectedMatch:  true,
		},
		"special characters in path": {
			path:           moqt.BroadcastPath("/api/user@example.com/profile"),
			pattern:        "/api/*/profile",
			expectedParams: []string{"user@example.com"},
			expectedMatch:  true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.path.ExtractParameters(tt.pattern)
			assert.Equal(t, tt.expectedParams, params, "Extracted parameters should match expected values")
		})
	}
}

func TestNewTrackPath(t *testing.T) {
	tests := map[string]struct {
		pattern  string
		segments []string
		expected moqt.BroadcastPath
	}{
		"empty pattern and segments": {
			pattern:  "",
			segments: []string{},
			expected: moqt.BroadcastPath(""),
		},
		"empty segments": {
			pattern:  "/test/audio",
			segments: []string{},
			expected: moqt.BroadcastPath("/test/audio"),
		},
		"single wildcard": {
			pattern:  "/test/*",
			segments: []string{"value"},
			expected: moqt.BroadcastPath("/test/value"),
		},
		"multiple wildcards": {
			pattern:  "/users/*/posts/*",
			segments: []string{"123", "456"},
			expected: moqt.BroadcastPath("/users/123/posts/456"),
		},
		"wildcard at beginning": {
			pattern:  "/*/resources",
			segments: []string{"tenant1"},
			expected: moqt.BroadcastPath("/tenant1/resources"),
		},
		"wildcard in middle": {
			pattern:  "/api/*/resources",
			segments: []string{"tenant1"},
			expected: moqt.BroadcastPath("/api/tenant1/resources"),
		},
		"wildcard at end": {
			pattern:  "/api/resources/*",
			segments: []string{"latest"},
			expected: moqt.BroadcastPath("/api/resources/latest"),
		},
		"segment count mismatch fewer": {
			pattern:  "/users/*/posts/*",
			segments: []string{"123"},
			expected: moqt.BroadcastPath("/users/123/posts/"),
		},
		"segment count mismatch more": {
			pattern:  "/users/*",
			segments: []string{"123", "456"},
			expected: moqt.BroadcastPath("/users/123"),
		},
		"pattern with double wildcard": {
			pattern:  "/api/**/profile",
			segments: []string{},
			expected: moqt.BroadcastPath("/api//profile"),
		},
		"special characters in segments": {
			pattern:  "/api/*/profile",
			segments: []string{"user@example.com"},
			expected: moqt.BroadcastPath("/api/user@example.com/profile"),
		},
		"multiple double wildcards": {
			pattern:  "/api/**/profile/**",
			segments: []string{"v1/123", "456"},
			expected: moqt.BroadcastPath("/api/v1/123/profile/456"),
		},
		"single and double wildcards": {
			pattern:  "/api/*/users/**/comments",
			segments: []string{"v1", "posts/123"},
			expected: moqt.BroadcastPath("/api/v1/users/posts/123/comments"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := moqt.NewTrackPath(tt.pattern, tt.segments...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroadcastPath_String(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		expected string
	}{
		"empty path": {
			path:     BroadcastPath(""),
			expected: "",
		},
		"simple path": {
			path:     BroadcastPath("/test/path"),
			expected: "/test/path",
		},
		"complex path": {
			path:     BroadcastPath("/test/path/with/multiple/segments"),
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

func TestBroadcastPath_HasPrefix(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		prefix   string
		expected bool
	}{
		"empty path and prefix": {
			path:     BroadcastPath(""),
			prefix:   "",
			expected: false,
		},
		"path shorter than prefix": {
			path:     BroadcastPath("/test"),
			prefix:   "/test/path",
			expected: false,
		},
		"matching prefix": {
			path:     BroadcastPath("/test/path/segment"),
			prefix:   "/test",
			expected: true,
		},
		"non-matching prefix": {
			path:     BroadcastPath("/test/path"),
			prefix:   "/other",
			expected: false,
		},
		"prefix without trailing slash": {
			path:     BroadcastPath("/test/path"),
			prefix:   "/test",
			expected: true,
		},
		"exact match is not a prefix": {
			path:     BroadcastPath("/test/path"),
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

func TestBroadcastPath_HasSuffix(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		suffix   string
		expected bool
	}{
		"empty path and suffix": {
			path:     BroadcastPath(""),
			suffix:   "",
			expected: false,
		},
		"path shorter than suffix": {
			path:     BroadcastPath("/test"),
			suffix:   "/test/path",
			expected: false,
		},
		"matching suffix": {
			path:     BroadcastPath("/segment/test/path"),
			suffix:   "path",
			expected: true,
		},
		"non-matching suffix": {
			path:     BroadcastPath("/test/path"),
			suffix:   "other",
			expected: false,
		},
		"suffix without leading slash": {
			path:     BroadcastPath("/test/path"),
			suffix:   "path",
			expected: true,
		},
		"exact match is not a suffix": {
			path:     BroadcastPath("/test/path"),
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

func TestBroadcastPath_Equal(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		target   BroadcastPath
		expected bool
	}{
		"empty paths": {
			path:     BroadcastPath(""),
			target:   BroadcastPath(""),
			expected: true,
		},
		"identical paths": {
			path:     BroadcastPath("/test/path"),
			target:   BroadcastPath("/test/path"),
			expected: true,
		},
		"different paths": {
			path:     BroadcastPath("/test/path1"),
			target:   BroadcastPath("/test/path2"),
			expected: false,
		},
		"case sensitive comparison": {
			path:     BroadcastPath("/Test/Path"),
			target:   BroadcastPath("/test/path"),
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

func TestBroadcastPath_Extension(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		expected string
	}{
		"no extension": {
			path:     BroadcastPath("/test/path"),
			expected: "",
		},
		"with extension": {
			path:     BroadcastPath("/test/path.mp4"),
			expected: ".mp4",
		},
		"multiple dots": {
			path:     BroadcastPath("/test/path.backup.mp4"),
			expected: ".mp4",
		},
		"hidden file with extension": {
			path:     BroadcastPath("/test/.hidden.txt"),
			expected: ".txt",
		},
		"path ending with dot": {
			path:     BroadcastPath("/test/path."),
			expected: ".",
		},
		"empty path": {
			path:     BroadcastPath(""),
			expected: "",
		},
		"only filename with extension": {
			path:     BroadcastPath("file.txt"),
			expected: ".txt",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.path.Extension()
			assert.Equal(t, tt.expected, result)
		})
	}
}

package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestBroadcastPath_String(t *testing.T) {
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

func TestBroadcastPath_HasPrefix(t *testing.T) {
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

func TestBroadcastPath_HasSuffix(t *testing.T) {
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

func TestBroadcastPath_Equal(t *testing.T) {
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

func TestBroadcastPath_Extension(t *testing.T) {
	tests := map[string]struct {
		path     moqt.BroadcastPath
		expected string
	}{
		"no extension": {
			path:     moqt.BroadcastPath("/test/path"),
			expected: "",
		},
		"with extension": {
			path:     moqt.BroadcastPath("/test/path.mp4"),
			expected: ".mp4",
		},
		"multiple dots": {
			path:     moqt.BroadcastPath("/test/path.backup.mp4"),
			expected: ".mp4",
		},
		"hidden file with extension": {
			path:     moqt.BroadcastPath("/test/.hidden.txt"),
			expected: ".txt",
		},
		"path ending with dot": {
			path:     moqt.BroadcastPath("/test/path."),
			expected: ".",
		},
		"empty path": {
			path:     moqt.BroadcastPath(""),
			expected: "",
		},
		"only filename with extension": {
			path:     moqt.BroadcastPath("file.txt"),
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

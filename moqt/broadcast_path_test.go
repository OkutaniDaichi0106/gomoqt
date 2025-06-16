package moqt

import (
	"strings"
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
		"empty path with root prefix": {
			path:     BroadcastPath(""),
			prefix:   "/",
			expected: false,
		},
		"path shorter than prefix": {
			path:     BroadcastPath("/test"),
			prefix:   "/test/path/",
			expected: false,
		},
		"matching prefix with trailing slash": {
			path:     BroadcastPath("/test/path/segment"),
			prefix:   "/test/",
			expected: true,
		},
		"non-matching prefix": {
			path:     BroadcastPath("/test/path"),
			prefix:   "/other/",
			expected: false,
		},
		"root prefix matches all": {
			path:     BroadcastPath("/test/path"),
			prefix:   "/",
			expected: true,
		},
		"exact match with trailing slash": {
			path:     BroadcastPath("/test/path"),
			prefix:   "/test/path/",
			expected: false,
		},
		"multi-level prefix match": {
			path:     BroadcastPath("/room/alice/stream1"),
			prefix:   "/room/alice/",
			expected: true,
		},
		"partial segment should not match": {
			path:     BroadcastPath("/testroom/alice"),
			prefix:   "/test/",
			expected: false,
		},
		"nested path structure": {
			path:     BroadcastPath("/broadcast/room/conference/alice"),
			prefix:   "/broadcast/room/",
			expected: true,
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

func TestBroadcastPath_GetSuffix(t *testing.T) {
	tests := map[string]struct {
		path           BroadcastPath
		prefix         string
		expectedSuffix string
		expectedOk     bool
	}{
		"valid prefix match": {
			path:           BroadcastPath("/root/path/to/file"),
			prefix:         "/root/path/",
			expectedSuffix: "to/file",
			expectedOk:     true,
		},
		"prefix not found": {
			path:           BroadcastPath("/root/path/to/file"),
			prefix:         "/other/path/",
			expectedSuffix: "",
			expectedOk:     false,
		},
		"exact match with trailing slash": {
			path:           BroadcastPath("/root/path"),
			prefix:         "/root/path/",
			expectedSuffix: "",
			expectedOk:     false,
		},
		"empty path": {
			path:           BroadcastPath(""),
			prefix:         "/root/",
			expectedSuffix: "",
			expectedOk:     false,
		},
		"root prefix": {
			path:           BroadcastPath("/test/file"),
			prefix:         "/",
			expectedSuffix: "test/file",
			expectedOk:     true,
		},
		"single segment suffix": {
			path:           BroadcastPath("/room/alice"),
			prefix:         "/room/",
			expectedSuffix: "alice",
			expectedOk:     true,
		},
		"multi-segment suffix": {
			path:           BroadcastPath("/broadcast/room/alice/stream1"),
			prefix:         "/broadcast/room/",
			expectedSuffix: "alice/stream1",
			expectedOk:     true,
		},
		"prefix longer than path": {
			path:           BroadcastPath("/test"),
			prefix:         "/test/longer/path/",
			expectedSuffix: "",
			expectedOk:     false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			suffix, ok := tt.path.GetSuffix(tt.prefix)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedSuffix, suffix)
		})
	}
}

func TestBroadcastPath_Boundary(t *testing.T) {
	tests := map[string]struct {
		path        BroadcastPath
		description string
	}{
		"empty path": {
			path:        BroadcastPath(""),
			description: "empty string should be handled gracefully",
		},
		"single character": {
			path:        BroadcastPath("a"),
			description: "single character path",
		},
		"single slash": {
			path:        BroadcastPath("/"),
			description: "root path only",
		},
		"very long path": {
			path:        BroadcastPath("/" + strings.Repeat("segment/", 100) + "end"),
			description: "very long path with many segments",
		},
		"path with unicode": {
			path:        BroadcastPath("/こんにちは/世界"),
			description: "path with unicode characters",
		},
		"path with special chars": {
			path:        BroadcastPath("/test@#$%/path-_+/file.ext"),
			description: "path with special characters",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Test String() method
			result := tt.path.String()
			assert.Equal(t, string(tt.path), result)

			// Test Equal() method
			assert.True(t, tt.path.Equal(tt.path))
			assert.False(t, tt.path.Equal(BroadcastPath("different")))
		})
	}
}

func TestBroadcastPath_EdgeCases(t *testing.T) {
	tests := map[string]struct {
		method      string
		path        BroadcastPath
		input       string
		expected    interface{}
		description string
	}{
		"hasPrefix_exact_match": {
			method:      "HasPrefix",
			path:        BroadcastPath("/test/path"),
			input:       "/test/path/",
			expected:    false,
			description: "exact match should return false for HasPrefix",
		},
		"hasSuffix_exact_match": {
			method:      "HasSuffix",
			path:        BroadcastPath("/test/path"),
			input:       "/test/path",
			expected:    false,
			description: "exact match should return false for HasSuffix",
		},
		"hasPrefix_longer_than_path": {
			method:      "HasPrefix",
			path:        BroadcastPath("/test"),
			input:       "/test/longer/path/",
			expected:    false,
			description: "prefix longer than path should return false",
		},
		"hasSuffix_longer_than_path": {
			method:      "HasSuffix",
			path:        BroadcastPath("/test"),
			input:       "/test/longer/path",
			expected:    false,
			description: "suffix longer than path should return false",
		},
		"hasPrefix_root_prefix": {
			method:      "HasPrefix",
			path:        BroadcastPath("/test/path"),
			input:       "/",
			expected:    true,
			description: "root prefix should match any non-empty path",
		},
		"hasSuffix_empty_input": {
			method:      "HasSuffix",
			path:        BroadcastPath("/test/path"),
			input:       "",
			expected:    false,
			description: "empty suffix should return false",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var result bool
			switch tt.method {
			case "HasPrefix":
				result = tt.path.HasPrefix(tt.input)
			case "HasSuffix":
				result = tt.path.HasSuffix(tt.input)
			}
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestBroadcastPath_CaseSensitivity(t *testing.T) {
	path := BroadcastPath("/Test/Path/Segment")

	tests := map[string]struct {
		method   string
		input    string
		expected bool
	}{
		"prefix_case_sensitive": {
			method:   "HasPrefix",
			input:    "/test/",
			expected: false,
		},
		"suffix_case_sensitive": {
			method:   "HasSuffix",
			input:    "segment",
			expected: false,
		},
		"prefix_exact_case": {
			method:   "HasPrefix",
			input:    "/Test/",
			expected: true,
		},
		"suffix_exact_case": {
			method:   "HasSuffix",
			input:    "Segment",
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var result bool
			switch tt.method {
			case "HasPrefix":
				result = path.HasPrefix(tt.input)
			case "HasSuffix":
				result = path.HasSuffix(tt.input)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBroadcastPath_PathSeparators(t *testing.T) {
	tests := map[string]struct {
		path        BroadcastPath
		prefix      string
		suffix      string
		hasPrefix   bool
		hasSuffix   bool
		description string
	}{
		"multiple_consecutive_slashes": {
			path:        BroadcastPath("/test//path///segment"),
			prefix:      "/test/",
			suffix:      "segment",
			hasPrefix:   true,
			hasSuffix:   true,
			description: "multiple consecutive slashes should be handled",
		},
		"trailing_slash": {
			path:        BroadcastPath("/test/path/"),
			prefix:      "/test/",
			suffix:      "path",
			hasPrefix:   true,
			hasSuffix:   false,
			description: "trailing slash affects suffix matching",
		},
		"leading_slash_in_suffix": {
			path:        BroadcastPath("/test/path/segment"),
			prefix:      "/test/",
			suffix:      "/segment",
			hasPrefix:   true,
			hasSuffix:   false,
			description: "leading slash in suffix parameter",
		},
		"no_separator_in_prefix": {
			path:        BroadcastPath("/testpath/segment"),
			prefix:      "/test/",
			suffix:      "segment",
			hasPrefix:   false,
			hasSuffix:   true,
			description: "prefix without proper separator",
		},
		"root_prefix_match": {
			path:        BroadcastPath("/any/path/segment"),
			prefix:      "/",
			suffix:      "segment",
			hasPrefix:   true,
			hasSuffix:   true,
			description: "root prefix should match any path",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			prefixResult := tt.path.HasPrefix(tt.prefix)
			suffixResult := tt.path.HasSuffix(tt.suffix)

			assert.Equal(t, tt.hasPrefix, prefixResult, "HasPrefix: "+tt.description)
			assert.Equal(t, tt.hasSuffix, suffixResult, "HasSuffix: "+tt.description)
		})
	}
}

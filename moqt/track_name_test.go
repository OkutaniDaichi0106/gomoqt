package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackName_Type(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"basic string": {
			input:    "test-track",
			expected: "test-track",
		},
		"empty string": {
			input:    "",
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			trackName := TrackName(tt.input)
			assert.Equal(t, TrackName(tt.expected), trackName)
			assert.Equal(t, tt.expected, string(trackName))
		})
	}
}

func TestTrackName_Empty(t *testing.T) {
	tests := map[string]struct {
		expected TrackName
	}{
		"zero value": {
			expected: TrackName(""),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var trackName TrackName
			assert.Equal(t, tt.expected, trackName)
			assert.Equal(t, "", string(trackName))
		})
	}
}

func TestTrackName_StringConversion(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"simple name": {
			input: "video",
		},
		"name with dashes": {
			input: "audio-high-quality",
		},
		"name with underscores": {
			input: "metadata_stream",
		},
		"name with numbers": {
			input: "track123",
		},
		"complex name": {
			input: "live-stream/camera-1/video/high",
		},
		"unicode name": {
			input: "トラック名",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			trackName := TrackName(tt.input)
			assert.Equal(t, tt.input, string(trackName))
		})
	}
}

func TestTrackName_Comparison(t *testing.T) {
	tests := map[string]struct {
		name1    TrackName
		name2    TrackName
		name3    TrackName
		expected bool
	}{
		"equal names": {
			name1:    TrackName("track-a"),
			name2:    TrackName("track-a"),
			name3:    TrackName("track-b"),
			expected: true,
		},
		"case sensitive": {
			name1:    TrackName("track"),
			name2:    TrackName("TRACK"),
			name3:    TrackName("track"),
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if name == "case sensitive" {
				// Test case sensitivity - these should be different
				assert.NotEqual(t, tt.name1, tt.name2)
				assert.Equal(t, tt.name1, tt.name3)
			} else {
				// Test equality
				assert.Equal(t, tt.name1, tt.name2)
				assert.NotEqual(t, tt.name1, tt.name3)
			}
		})
	}
}

func TestTrackName_Length(t *testing.T) {
	tests := map[string]struct {
		input  string
		length int
	}{
		"empty": {
			input:  "",
			length: 0,
		},
		"single char": {
			input:  "a",
			length: 1,
		},
		"normal length": {
			input:  "video-track",
			length: 11,
		},
		"long name": {
			input:  "very-long-track-name-with-many-components",
			length: 41,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			trackName := TrackName(tt.input)
			assert.Equal(t, tt.length, len(string(trackName)))
		})
	}
}

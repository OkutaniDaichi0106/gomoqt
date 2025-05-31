package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestTrackName_Type(t *testing.T) {
	// Test basic string operations
	name := moqt.TrackName("test-track")
	assert.Equal(t, moqt.TrackName("test-track"), name)
	assert.Equal(t, "test-track", string(name))
}

func TestTrackName_Empty(t *testing.T) {
	// Test empty track name
	var name moqt.TrackName
	assert.Equal(t, moqt.TrackName(""), name)
	assert.Equal(t, "", string(name))
}

func TestTrackName_StringConversion(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple name",
			input: "video",
		},
		{
			name:  "name with dashes",
			input: "audio-high-quality",
		},
		{
			name:  "name with underscores",
			input: "metadata_stream",
		},
		{
			name:  "name with numbers",
			input: "track123",
		},
		{
			name:  "complex name",
			input: "live-stream/camera-1/video/high",
		},
		{
			name:  "unicode name",
			input: "トラック名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackName := moqt.TrackName(tt.input)
			assert.Equal(t, tt.input, string(trackName))
		})
	}
}

func TestTrackName_Comparison(t *testing.T) {
	name1 := moqt.TrackName("track-a")
	name2 := moqt.TrackName("track-a")
	name3 := moqt.TrackName("track-b")

	// Test equality
	assert.Equal(t, name1, name2)
	assert.NotEqual(t, name1, name3)

	// Test with different cases
	nameLower := moqt.TrackName("track")
	nameUpper := moqt.TrackName("TRACK")
	assert.NotEqual(t, nameLower, nameUpper)
}

func TestTrackName_Length(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		length int
	}{
		{
			name:   "empty",
			input:  "",
			length: 0,
		},
		{
			name:   "single char",
			input:  "a",
			length: 1,
		},
		{
			name:   "normal length",
			input:  "video-track",
			length: 11,
		},
		{
			name:   "long name",
			input:  "very-long-track-name-with-many-components",
			length: 41,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackName := moqt.TrackName(tt.input)
			assert.Equal(t, tt.length, len(string(trackName)))
		})
	}
}

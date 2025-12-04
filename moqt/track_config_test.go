package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackConfig(t *testing.T) {
	tests := map[string]struct {
		trackPriority TrackPriority
	}{
		"default values": {
			trackPriority: TrackPriority(0),
		},
		"specific priority": {
			trackPriority: TrackPriority(128),
		},
		"high priority": {
			trackPriority: TrackPriority(255),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			config := TrackConfig{
				TrackPriority: tt.trackPriority,
			}

			assert.Equal(t, tt.trackPriority, config.TrackPriority)
		})
	}
}

func TestTrackConfig_String(t *testing.T) {
	tests := map[string]struct {
		config           TrackConfig
		expectedContains string
	}{
		"default config": {
			config:           TrackConfig{TrackPriority: 0},
			expectedContains: "track_priority: 0",
		},
		"custom priority": {
			config:           TrackConfig{TrackPriority: 100},
			expectedContains: "track_priority: 100",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.config.String()
			assert.Contains(t, result, tt.expectedContains, "String() should contain TrackPriority")
		})
	}
}

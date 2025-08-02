package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackConfig(t *testing.T) {
	tests := map[string]struct {
		trackPriority    TrackPriority
		minGroupSequence GroupSequence
		maxGroupSequence GroupSequence
	}{
		"default values": {
			trackPriority:    TrackPriority(0),
			minGroupSequence: GroupSequence(0),
			maxGroupSequence: GroupSequence(0),
		},
		"specific range": {
			trackPriority:    TrackPriority(128),
			minGroupSequence: GroupSequence(10),
			maxGroupSequence: GroupSequence(100),
		}, "high priority": {
			trackPriority:    TrackPriority(255),
			minGroupSequence: GroupSequence(1),
			maxGroupSequence: GroupSequence(1000),
		},
		"maximum boundary values": {
			trackPriority:    TrackPriority(255),
			minGroupSequence: GroupSequenceFirst,
			maxGroupSequence: MaxGroupSequence,
		}}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			config := TrackConfig{
				TrackPriority:    tt.trackPriority,
				MinGroupSequence: tt.minGroupSequence,
				MaxGroupSequence: tt.maxGroupSequence,
			}

			assert.Equal(t, tt.trackPriority, config.TrackPriority)
			assert.Equal(t, tt.minGroupSequence, config.MinGroupSequence)
			assert.Equal(t, tt.maxGroupSequence, config.MaxGroupSequence)
		})
	}
}

func TestTrackConfig_IsInRange(t *testing.T) {
	tests := map[string]struct {
		config   TrackConfig
		seq      GroupSequence
		expected bool
	}{
		"both min and max not specified": {
			config: TrackConfig{
				MinGroupSequence: GroupSequenceNotSpecified,
				MaxGroupSequence: GroupSequenceNotSpecified,
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		"only min not specified": {
			config: TrackConfig{
				MinGroupSequence: GroupSequenceNotSpecified,
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		"only min not specified - above max": {
			config: TrackConfig{
				MinGroupSequence: GroupSequenceNotSpecified,
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(150),
			expected: false,
		},
		"only max not specified": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequenceNotSpecified,
			},
			seq:      GroupSequence(50),
			expected: true},
		"only max not specified - below min": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequenceNotSpecified,
			},
			seq:      GroupSequence(5),
			expected: false,
		},
		"in range": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		"at min boundary": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(10),
			expected: true,
		},
		"at max boundary": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(100),
			expected: true,
		},
		"below min": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(5),
			expected: false,
		},
		"above max": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(150),
			expected: false,
		},
		"single value range": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(50),
				MaxGroupSequence: GroupSequence(50),
			},
			seq:      GroupSequence(50),
			expected: true,
		}, "single value range - different value": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(50),
				MaxGroupSequence: GroupSequence(50),
			},
			seq:      GroupSequence(51),
			expected: false,
		},
		"max boundary with MaxGroupSequence": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(1),
				MaxGroupSequence: MaxGroupSequence,
			},
			seq:      MaxGroupSequence,
			expected: true,
		},
		"max boundary with MaxGroupSequence - above max": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(1),
				MaxGroupSequence: MaxGroupSequence - 1,
			},
			seq:      MaxGroupSequence,
			expected: false,
		}, "GroupSequenceFirst boundary": {
			config: TrackConfig{
				MinGroupSequence: GroupSequenceFirst,
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequenceFirst,
			expected: true,
		},
		"invalid range - min greater than max": {
			config: TrackConfig{
				MinGroupSequence: GroupSequence(100),
				MaxGroupSequence: GroupSequence(50),
			},
			seq:      GroupSequence(75),
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.config.IsInRange(tt.seq)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackConfig_String(t *testing.T) {
	tests := map[string]struct {
		config   TrackConfig
		expected string
	}{
		"default values": {
			config: TrackConfig{
				TrackPriority:    TrackPriority(0),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(0),
			},
			expected: "{ track_priority: 0, min_group_sequence: 0, max_group_sequence: 0 }",
		},
		"specific values": {
			config: TrackConfig{
				TrackPriority:    TrackPriority(128),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			expected: "{ track_priority: 128, min_group_sequence: 10, max_group_sequence: 100 }",
		}, "high values": {
			config: TrackConfig{
				TrackPriority:    TrackPriority(255),
				MinGroupSequence: GroupSequence(1000),
				MaxGroupSequence: GroupSequence(9999),
			},
			expected: "{ track_priority: 255, min_group_sequence: 1000, max_group_sequence: 9999 }",
		}, "maximum boundary values": {
			config: TrackConfig{
				TrackPriority:    TrackPriority(255),
				MinGroupSequence: GroupSequenceFirst,
				MaxGroupSequence: MaxGroupSequence,
			},
			expected: "{ track_priority: 255, min_group_sequence: 1, max_group_sequence: 4611686018427387903 }",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.config.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackConfig_ZeroValue(t *testing.T) {
	var config TrackConfig

	assert.Equal(t, TrackPriority(0), config.TrackPriority)
	assert.Equal(t, GroupSequence(0), config.MinGroupSequence)
	assert.Equal(t, GroupSequence(0), config.MaxGroupSequence)

	// Test that zero values behave correctly with IsInRange
	// Since both min and max are 0 (GroupSequenceNotSpecified), it should accept any sequence
	assert.True(t, config.IsInRange(GroupSequence(0)))
	assert.True(t, config.IsInRange(GroupSequence(50)))
	assert.True(t, config.IsInRange(MaxGroupSequence))
}

func TestTrackConfig_Comparison(t *testing.T) {
	config1 := TrackConfig{
		TrackPriority:    TrackPriority(128),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(100),
	}

	config2 := TrackConfig{
		TrackPriority:    TrackPriority(128),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(100),
	}

	config3 := TrackConfig{
		TrackPriority:    TrackPriority(64),
		MinGroupSequence: GroupSequence(20),
		MaxGroupSequence: GroupSequence(200),
	}

	assert.Equal(t, config1, config2)
	assert.NotEqual(t, config1, config3)
}

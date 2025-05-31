package moqt

import (
	"testing"
)

func TestSubscribeConfig(t *testing.T) {
	tests := []struct {
		name             string
		trackPriority    TrackPriority
		minGroupSequence GroupSequence
		maxGroupSequence GroupSequence
	}{
		{
			name:             "default values",
			trackPriority:    TrackPriority(0),
			minGroupSequence: GroupSequence(0),
			maxGroupSequence: GroupSequence(0),
		},
		{
			name:             "specific range",
			trackPriority:    TrackPriority(128),
			minGroupSequence: GroupSequence(10),
			maxGroupSequence: GroupSequence(100),
		},
		{
			name:             "high priority",
			trackPriority:    TrackPriority(255),
			minGroupSequence: GroupSequence(1),
			maxGroupSequence: GroupSequence(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SubscribeConfig{
				TrackPriority:    tt.trackPriority,
				MinGroupSequence: tt.minGroupSequence,
				MaxGroupSequence: tt.maxGroupSequence,
			}

			if config.TrackPriority != tt.trackPriority {
				t.Errorf("TrackPriority = %v, want %v", config.TrackPriority, tt.trackPriority)
			}

			if config.MinGroupSequence != tt.minGroupSequence {
				t.Errorf("MinGroupSequence = %v, want %v", config.MinGroupSequence, tt.minGroupSequence)
			}

			if config.MaxGroupSequence != tt.maxGroupSequence {
				t.Errorf("MaxGroupSequence = %v, want %v", config.MaxGroupSequence, tt.maxGroupSequence)
			}
		})
	}
}

func TestSubscribeConfigIsInRange(t *testing.T) {
	tests := []struct {
		name     string
		config   SubscribeConfig
		seq      GroupSequence
		expected bool
	}{
		{
			name: "both min and max not specified",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequenceNotSpecified,
				MaxGroupSequence: GroupSequenceNotSpecified,
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		{
			name: "only min not specified",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequenceNotSpecified,
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		{
			name: "only min not specified - above max",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequenceNotSpecified,
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(150),
			expected: false,
		},
		{
			name: "only max not specified",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequenceNotSpecified,
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		{
			name: "only max not specified - below min",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequenceNotSpecified,
			},
			seq:      GroupSequence(5),
			expected: false,
		},
		{
			name: "in range",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		{
			name: "at min boundary",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(10),
			expected: true,
		},
		{
			name: "at max boundary",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(100),
			expected: true,
		},
		{
			name: "below min",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(5),
			expected: false,
		},
		{
			name: "above max",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			seq:      GroupSequence(150),
			expected: false,
		},
		{
			name: "single value range",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(50),
				MaxGroupSequence: GroupSequence(50),
			},
			seq:      GroupSequence(50),
			expected: true,
		},
		{
			name: "single value range - different value",
			config: SubscribeConfig{
				MinGroupSequence: GroupSequence(50),
				MaxGroupSequence: GroupSequence(50),
			},
			seq:      GroupSequence(51),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsInRange(tt.seq)
			if result != tt.expected {
				t.Errorf("IsInRange(%v) = %v, want %v", tt.seq, result, tt.expected)
			}
		})
	}
}

func TestSubscribeConfigString(t *testing.T) {
	tests := []struct {
		name     string
		config   SubscribeConfig
		expected string
	}{
		{
			name: "default values",
			config: SubscribeConfig{
				TrackPriority:    TrackPriority(0),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(0),
			},
			expected: "SubscribeConfig: { TrackPriority: 0, MinGroupSequence: 0, MaxGroupSequence: 0 }",
		},
		{
			name: "specific values",
			config: SubscribeConfig{
				TrackPriority:    TrackPriority(128),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			expected: "SubscribeConfig: { TrackPriority: 128, MinGroupSequence: 10, MaxGroupSequence: 100 }",
		},
		{
			name: "high values",
			config: SubscribeConfig{
				TrackPriority:    TrackPriority(255),
				MinGroupSequence: GroupSequence(1000),
				MaxGroupSequence: GroupSequence(9999),
			},
			expected: "SubscribeConfig: { TrackPriority: 255, MinGroupSequence: 1000, MaxGroupSequence: 9999 }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSubscribeConfigZeroValue(t *testing.T) {
	var config SubscribeConfig

	if config.TrackPriority != 0 {
		t.Errorf("zero value TrackPriority = %v, want 0", config.TrackPriority)
	}

	if config.MinGroupSequence != 0 {
		t.Errorf("zero value MinGroupSequence = %v, want 0", config.MinGroupSequence)
	}

	if config.MaxGroupSequence != 0 {
		t.Errorf("zero value MaxGroupSequence = %v, want 0", config.MaxGroupSequence)
	}

	// Test that zero values behave correctly with IsInRange
	// Since both min and max are 0 (not GroupSequenceNotSpecified), it should only accept sequence 0
	if !config.IsInRange(GroupSequence(0)) {
		t.Error("zero value config should accept sequence 0")
	}

	// Note: This behavior depends on whether 0 is considered GroupSequenceNotSpecified
	// Based on the IsInRange implementation, it seems like GroupSequenceNotSpecified is a specific value
}

func TestSubscribeConfigComparison(t *testing.T) {
	config1 := SubscribeConfig{
		TrackPriority:    TrackPriority(128),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(100),
	}

	config2 := SubscribeConfig{
		TrackPriority:    TrackPriority(128),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(100),
	}

	config3 := SubscribeConfig{
		TrackPriority:    TrackPriority(64),
		MinGroupSequence: GroupSequence(20),
		MaxGroupSequence: GroupSequence(200),
	}

	if config1 != config2 {
		t.Error("identical configs should be equal")
	}

	if config1 == config3 {
		t.Error("different configs should not be equal")
	}
}

package internal

import "testing"

func TestTrackPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority TrackPriority
		expected uint8
	}{
		{"zero", TrackPriority(0), 0},
		{"one", TrackPriority(1), 1},
		{"max", TrackPriority(255), 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint8(tt.priority) != tt.expected {
				t.Errorf("TrackPriority(%d) = %d, want %d", tt.priority, uint8(tt.priority), tt.expected)
			}
		})
	}
}

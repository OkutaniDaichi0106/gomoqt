package protocol

import "testing"

func TestGroupSequence(t *testing.T) {
	tests := []struct {
		name     string
		seq      GroupSequence
		expected uint64
	}{
		{"zero", GroupSequence(0), 0},
		{"one", GroupSequence(1), 1},
		{"max", GroupSequence(^uint64(0)), ^uint64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint64(tt.seq) != tt.expected {
				t.Errorf("GroupSequence(%d) = %d, want %d", tt.seq, uint64(tt.seq), tt.expected)
			}
		})
	}
}

func TestGroupSequenceNext(t *testing.T) {
	tests := []struct {
		name     string
		seq      GroupSequence
		expected GroupSequence
	}{
		{"not specified", GroupSequenceNotSpecified, 1},
		{"first", GroupSequenceFirst, 2},
		{"normal", GroupSequence(5), 6},
		{"max", MaxGroupSequence, 1}, // wrap around
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.seq.Next()
			if result != tt.expected {
				t.Errorf("GroupSequence(%d).Next() = %d, want %d", tt.seq, result, tt.expected)
			}
		})
	}
}

package protocol

import "testing"

func TestSubscribeID(t *testing.T) {
	tests := []struct {
		name     string
		id       SubscribeID
		expected uint64
	}{
		{"zero", SubscribeID(0), 0},
		{"one", SubscribeID(1), 1},
		{"max", SubscribeID(^uint64(0)), ^uint64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint64(tt.id) != tt.expected {
				t.Errorf("SubscribeID(%d) = %d, want %d", tt.id, uint64(tt.id), tt.expected)
			}
		})
	}
}

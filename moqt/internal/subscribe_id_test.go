package internal

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

func TestSubscribeIDString(t *testing.T) {
	tests := []struct {
		name     string
		id       SubscribeID
		expected string
	}{
		{"zero", SubscribeID(0), "0"},
		{"one", SubscribeID(1), "1"},
		{"large", SubscribeID(12345), "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.String()
			if result != tt.expected {
				t.Errorf("SubscribeID(%d).String() = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

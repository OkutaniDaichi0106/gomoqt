package internal

import "testing"

func TestVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected uint64
	}{
		{"zero", Version(0), 0},
		{"one", Version(1), 1},
		{"draft07", Version(0xff000007), 0xff000007},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint64(tt.version) != tt.expected {
				t.Errorf("Version(%d) = %d, want %d", tt.version, uint64(tt.version), tt.expected)
			}
		})
	}
}

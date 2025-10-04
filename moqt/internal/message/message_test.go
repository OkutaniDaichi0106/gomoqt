package message

import (
	"testing"
)

func TestVarintLen(t *testing.T) {
	tests := []struct {
		input    uint64
		expected int
	}{
		{0, 1},
		{maxVarInt1, 1},
		{maxVarInt1 + 1, 2},
		{maxVarInt2, 2},
		{maxVarInt2 + 1, 4},
		{maxVarInt4, 4},
		{maxVarInt4 + 1, 8},
		{maxVarInt8, 8},
	}

	for _, tt := range tests {
		result := VarintLen(tt.input)
		if result != tt.expected {
			t.Errorf("VarintLen(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestVarintLenPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("VarintLen should panic for large values")
		}
	}()
	VarintLen(maxVarInt8 + 1)
}

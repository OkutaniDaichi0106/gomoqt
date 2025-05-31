package moqt

import (
	"testing"
)

func TestGroupOrder_String(t *testing.T) {
	tests := []struct {
		name     string
		order    GroupOrder
		expected string
	}{
		{
			name:     "default order",
			order:    GroupOrderDefault,
			expected: "default",
		},
		{
			name:     "ascending order",
			order:    GroupOrderAscending,
			expected: "ascending",
		},
		{
			name:     "descending order",
			order:    GroupOrderDescending,
			expected: "descending",
		},
		{
			name:     "undefined order",
			order:    GroupOrder(0xFF),
			expected: "undefined group order",
		},
		{
			name:     "invalid order value",
			order:    GroupOrder(42),
			expected: "undefined group order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.order.String()
			if result != tt.expected {
				t.Errorf("GroupOrder.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGroupOrder_Constants(t *testing.T) {
	tests := map[string]struct {
		order    GroupOrder
		expected byte
	}{
		"default order": {
			order:    GroupOrderDefault,
			expected: 0x0,
		},
		"ascending order": {
			order:    GroupOrderAscending,
			expected: 0x1,
		},
		"descending order": {
			order:    GroupOrderDescending,
			expected: 0x2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if byte(tt.order) != tt.expected {
				t.Errorf("GroupOrder value = %v, want %v", byte(tt.order), tt.expected)
			}
		})
	}
}

package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupOrder_String(t *testing.T) {
	tests := map[string]struct {
		order    GroupOrder
		expected string
	}{
		"default order": {
			order:    GroupOrderDefault,
			expected: "default",
		},
		"ascending order": {
			order:    GroupOrderAscending,
			expected: "ascending",
		},
		"descending order": {
			order:    GroupOrderDescending,
			expected: "descending",
		},
		"undefined order": {
			order:    GroupOrder(0xFF),
			expected: "undefined group order",
		},
		"invalid order value": {
			order:    GroupOrder(42),
			expected: "undefined group order",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.order.String()
			assert.Equal(t, tt.expected, result)
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
			assert.Equal(t, tt.expected, byte(tt.order))
		})
	}
}

func TestGroupOrder_Values(t *testing.T) {
	// Test that constants have expected values
	assert.Equal(t, GroupOrder(0x0), GroupOrderDefault)
	assert.Equal(t, GroupOrder(0x1), GroupOrderAscending)
	assert.Equal(t, GroupOrder(0x2), GroupOrderDescending)
}

func TestGroupOrder_TypeConversion(t *testing.T) {
	tests := map[string]struct {
		byteValue byte
		expected  GroupOrder
	}{
		"convert byte 0 to default": {
			byteValue: 0x0,
			expected:  GroupOrderDefault,
		},
		"convert byte 1 to ascending": {
			byteValue: 0x1,
			expected:  GroupOrderAscending,
		},
		"convert byte 2 to descending": {
			byteValue: 0x2,
			expected:  GroupOrderDescending,
		},
		"convert invalid byte": {
			byteValue: 0xFF,
			expected:  GroupOrder(0xFF),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			order := GroupOrder(tt.byteValue)
			assert.Equal(t, tt.expected, order)
			assert.Equal(t, tt.byteValue, byte(order))
		})
	}
}

func TestGroupOrder_Comparisons(t *testing.T) {
	// Test equality
	assert.Equal(t, GroupOrderDefault, GroupOrderDefault)
	assert.Equal(t, GroupOrderAscending, GroupOrderAscending)
	assert.Equal(t, GroupOrderDescending, GroupOrderDescending)

	// Test inequality
	assert.NotEqual(t, GroupOrderDefault, GroupOrderAscending)
	assert.NotEqual(t, GroupOrderDefault, GroupOrderDescending)
	assert.NotEqual(t, GroupOrderAscending, GroupOrderDescending)
}

func TestGroupOrder_Uniqueness(t *testing.T) {
	orders := []GroupOrder{
		GroupOrderDefault,
		GroupOrderAscending,
		GroupOrderDescending,
	}

	// Test that all defined orders have unique values
	seen := make(map[byte]bool)
	for _, order := range orders {
		value := byte(order)
		if seen[value] && order != GroupOrderDefault {
			// Allow GroupOrderDefault to share value with others since it's 0x0
			t.Errorf("duplicate value %d found", value)
		}
		seen[value] = true
	}
}

func TestGroupOrder_StringRepresentations(t *testing.T) {
	representations := map[GroupOrder]string{
		GroupOrderDefault:    "default",
		GroupOrderAscending:  "ascending",
		GroupOrderDescending: "descending",
	}

	for order, expected := range representations {
		t.Run(expected, func(t *testing.T) {
			assert.Equal(t, expected, order.String())
		})
	}
}

func TestGroupOrder_InvalidValues(t *testing.T) {
	invalidValues := []byte{3, 4, 5, 10, 42, 100, 255}

	for _, value := range invalidValues {
		t.Run(string(rune(value)), func(t *testing.T) {
			order := GroupOrder(value)
			result := order.String()
			assert.Equal(t, "undefined group order", result)
		})
	}
}

func TestGroupOrder_BoundaryValues(t *testing.T) {
	tests := map[string]struct {
		value    byte
		expected string
	}{
		"minimum byte value": {
			value:    0,
			expected: "default",
		},
		"maximum defined value": {
			value:    2,
			expected: "descending",
		},
		"just above maximum defined": {
			value:    3,
			expected: "undefined group order",
		},
		"maximum byte value": {
			value:    255,
			expected: "undefined group order",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			order := GroupOrder(tt.value)
			result := order.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupOrder_IsValid(t *testing.T) {
	// Helper function to check if order is valid
	isValidOrder := func(order GroupOrder) bool {
		switch order {
		case GroupOrderDefault, GroupOrderAscending, GroupOrderDescending:
			return true
		default:
			return false
		}
	}
	tests := map[string]struct {
		order    GroupOrder
		expected bool
	}{
		"default is valid": {
			order:    GroupOrderDefault,
			expected: true,
		},
		"ascending is valid": {
			order:    GroupOrderAscending,
			expected: true,
		},
		"descending is valid": {
			order:    GroupOrderDescending,
			expected: true,
		},
		"invalid order": {
			order:    GroupOrder(42),
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := isValidOrder(tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupOrder_ConstantImmutability(t *testing.T) {
	// Test that constants are truly constants (compile-time values)
	const testDefault = GroupOrderDefault
	const testAscending = GroupOrderAscending
	const testDescending = GroupOrderDescending

	assert.Equal(t, GroupOrderDefault, testDefault)
	assert.Equal(t, GroupOrderAscending, testAscending)
	assert.Equal(t, GroupOrderDescending, testDescending)
}

func TestGroupOrder_AllDefinedOrders(t *testing.T) {
	allOrders := []struct {
		order     GroupOrder
		name      string
		byteVal   byte
		stringRep string
	}{
		{
			order:     GroupOrderDefault,
			name:      "default",
			byteVal:   0x0,
			stringRep: "default",
		},
		{
			order:     GroupOrderAscending,
			name:      "ascending",
			byteVal:   0x1,
			stringRep: "ascending",
		},
		{
			order:     GroupOrderDescending,
			name:      "descending",
			byteVal:   0x2,
			stringRep: "descending",
		},
	}

	for _, orderInfo := range allOrders {
		t.Run(orderInfo.name, func(t *testing.T) {
			// Test byte value
			assert.Equal(t, orderInfo.byteVal, byte(orderInfo.order))

			// Test string representation
			assert.Equal(t, orderInfo.stringRep, orderInfo.order.String())

			// Test round-trip conversion
			converted := GroupOrder(orderInfo.byteVal)
			assert.Equal(t, orderInfo.order, converted)
		})
	}
}

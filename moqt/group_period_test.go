package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupPeriod_Constants(t *testing.T) {
	tests := map[string]struct {
		period   GroupPeriod
		expected uint64
	}{
		"irregular": {
			period:   GroupPeriodIrregular,
			expected: 0,
		},
		"second": {
			period:   GroupPeriodSecond,
			expected: 1000,
		},
		"minute": {
			period:   GroupPeriodMinute,
			expected: 60000,
		},
		"hour": {
			period:   GroupPeriodHour,
			expected: 3600000,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, uint64(tt.period))
		})
	}
}

func TestGroupPeriod_Uniqueness(t *testing.T) {
	orders := []GroupPeriod{
		GroupPeriodIrregular,
		GroupPeriodSecond,
		GroupPeriodMinute,
		GroupPeriodHour,
	}

	// Test that all defined orders have unique values
	seen := make(map[byte]bool)
	for _, order := range orders {
		value := byte(order)
		if seen[value] && order != GroupPeriodIrregular {
			// Allow GroupPeriodIrregular to share value with others since it's 0
			t.Errorf("duplicate value %d found", value)
		}
		seen[value] = true
	}
}

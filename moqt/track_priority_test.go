package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackPriority_Range(t *testing.T) {
	tests := map[string]struct {
		priority byte
	}{
		"minimum priority": {
			priority: 0,
		},
		"low priority": {
			priority: 1,
		},
		"medium priority": {
			priority: 128,
		},
		"high priority": {
			priority: 200,
		},
		"maximum priority": {
			priority: 255,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			priority := TrackPriority(tt.priority)
			assert.Equal(t, tt.priority, byte(priority))
		})
	}
}

func TestTrackPriority_Comparison(t *testing.T) {
	low := TrackPriority(10)
	medium := TrackPriority(100)
	high := TrackPriority(200)

	// Test ordering (lower values should be "less than" higher values)
	assert.True(t, low < medium)
	assert.True(t, medium < high)
	assert.True(t, low < high)

	// Test equality
	same := TrackPriority(100)
	assert.Equal(t, medium, same)
	assert.False(t, medium < same)
	assert.False(t, same < medium)
}

func TestTrackPriority_Arithmetic(t *testing.T) {
	priority := TrackPriority(100)

	// Test increment
	priority++
	assert.Equal(t, TrackPriority(101), priority)

	// Test decrement
	priority--
	assert.Equal(t, TrackPriority(100), priority)

	// Test addition
	result := priority + TrackPriority(50)
	assert.Equal(t, TrackPriority(150), result)

	// Test subtraction
	result = priority - TrackPriority(30)
	assert.Equal(t, TrackPriority(70), result)
}

func TestTrackPriority_Overflow(t *testing.T) {
	// Test overflow behavior (should wrap around)
	maxPriority := TrackPriority(255)
	overflow := maxPriority + 1
	assert.Equal(t, TrackPriority(0), overflow)

	// Test underflow behavior
	minPriority := TrackPriority(0)
	underflow := minPriority - 1
	assert.Equal(t, TrackPriority(255), underflow)
}

package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestTrackPriority_Range(t *testing.T) {
	tests := []struct {
		name     string
		priority byte
	}{
		{
			name:     "minimum priority",
			priority: 0,
		},
		{
			name:     "low priority",
			priority: 1,
		},
		{
			name:     "medium priority",
			priority: 128,
		},
		{
			name:     "high priority",
			priority: 200,
		},
		{
			name:     "maximum priority",
			priority: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := moqt.TrackPriority(tt.priority)
			assert.Equal(t, tt.priority, byte(priority))
		})
	}
}

func TestTrackPriority_Comparison(t *testing.T) {
	low := moqt.TrackPriority(10)
	medium := moqt.TrackPriority(100)
	high := moqt.TrackPriority(200)

	// Test ordering (lower values should be "less than" higher values)
	assert.True(t, low < medium)
	assert.True(t, medium < high)
	assert.True(t, low < high)

	// Test equality
	same := moqt.TrackPriority(100)
	assert.Equal(t, medium, same)
	assert.False(t, medium < same)
	assert.False(t, same < medium)
}

func TestTrackPriority_Arithmetic(t *testing.T) {
	priority := moqt.TrackPriority(100)

	// Test increment
	priority++
	assert.Equal(t, moqt.TrackPriority(101), priority)

	// Test decrement
	priority--
	assert.Equal(t, moqt.TrackPriority(100), priority)

	// Test addition
	result := priority + moqt.TrackPriority(50)
	assert.Equal(t, moqt.TrackPriority(150), result)

	// Test subtraction
	result = priority - moqt.TrackPriority(30)
	assert.Equal(t, moqt.TrackPriority(70), result)
}

func TestTrackPriority_Overflow(t *testing.T) {
	// Test overflow behavior (should wrap around)
	maxPriority := moqt.TrackPriority(255)
	overflow := maxPriority + 1
	assert.Equal(t, moqt.TrackPriority(0), overflow)

	// Test underflow behavior
	minPriority := moqt.TrackPriority(0)
	underflow := minPriority - 1
	assert.Equal(t, moqt.TrackPriority(255), underflow)
}

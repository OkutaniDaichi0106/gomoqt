package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeID_String(t *testing.T) {
	tests := map[string]struct {
		id   SubscribeID
		want string
	}{
		"zero id": {
			id:   0,
			want: "0",
		},
		"small id": {
			id:   42,
			want: "42",
		},
		"large id": {
			id:   18446744073709551615, // max uint64
			want: "18446744073709551615",
		},
		"typical id": {
			id:   12345,
			want: "12345",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.id.String()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSubscribeID_Type(t *testing.T) {
	// Test that SubscribeID is based on uint64
	var id SubscribeID = 100

	// Test assignment and comparison
	assert.Equal(t, SubscribeID(100), id)

	// Test arithmetic operations
	id++
	assert.Equal(t, SubscribeID(101), id)

	id--
	assert.Equal(t, SubscribeID(100), id)
}

func TestSubscribeID_ZeroValue(t *testing.T) {
	// Test zero value behavior
	var id SubscribeID
	assert.Equal(t, SubscribeID(0), id)
	assert.Equal(t, "0", id.String())
}

func TestSubscribeID_MaxValue(t *testing.T) {
	// Test maximum value
	var maxID SubscribeID = ^SubscribeID(0) // max uint64
	assert.Equal(t, "18446744073709551615", maxID.String())
}

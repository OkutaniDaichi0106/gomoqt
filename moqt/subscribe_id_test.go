package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeID_String(t *testing.T) {
	tests := []struct {
		name string
		id   moqt.SubscribeID
		want string
	}{
		{
			name: "zero id",
			id:   0,
			want: "SubscribeID: 0",
		},
		{
			name: "small id",
			id:   42,
			want: "SubscribeID: 42",
		},
		{
			name: "large id",
			id:   18446744073709551615, // max uint64
			want: "SubscribeID: 18446744073709551615",
		},
		{
			name: "typical id",
			id:   12345,
			want: "SubscribeID: 12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.String()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSubscribeID_Type(t *testing.T) {
	// Test that SubscribeID is based on uint64
	var id moqt.SubscribeID = 100

	// Test assignment and comparison
	assert.Equal(t, moqt.SubscribeID(100), id)

	// Test arithmetic operations
	id++
	assert.Equal(t, moqt.SubscribeID(101), id)

	id--
	assert.Equal(t, moqt.SubscribeID(100), id)
}

func TestSubscribeID_ZeroValue(t *testing.T) {
	// Test zero value behavior
	var id moqt.SubscribeID
	assert.Equal(t, moqt.SubscribeID(0), id)
	assert.Equal(t, "SubscribeID: 0", id.String())
}

func TestSubscribeID_MaxValue(t *testing.T) {
	// Test maximum value
	var maxID moqt.SubscribeID = ^moqt.SubscribeID(0) // max uint64
	assert.Equal(t, "SubscribeID: 18446744073709551615", maxID.String())
}

package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamTypeValues(t *testing.T) {
	// Test that stream type constants have expected values
	assert.Equal(t, byte(0x0), byte(stream_type_session))
	assert.Equal(t, byte(0x1), byte(stream_type_announce))
	assert.Equal(t, byte(0x2), byte(stream_type_subscribe))
	assert.Equal(t, byte(0x0), byte(stream_type_group))
}

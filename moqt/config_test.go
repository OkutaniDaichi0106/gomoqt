package moqt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Clone(t *testing.T) {
	tests := map[string]struct {
		config *Config
	}{
		"config with all fields": {
			config: &Config{
				SetupTimeout: 30 * time.Second,
			},
		},
		"config with nil fields": {
			config: &Config{
				SetupTimeout: 10 * time.Second,
			},
		},
		"config with zero values": {
			config: &Config{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			original := tt.config
			cloned := original.Clone()

			assert.NotNil(t, cloned, "Cloned config should not be nil")

			// Check if both are nil or both are non-nil for function fields
			assert.Equal(t, original.SetupTimeout, cloned.SetupTimeout, "Timeout should be equal")
		})
	}
}

func TestConfig_setupTimeout(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		var c *Config
		timeout := c.setupTimeout()
		assert.Equal(t, 5*time.Second, timeout, "nil config should return default 5s timeout")
	})

	t.Run("zero timeout returns default", func(t *testing.T) {
		c := &Config{
			SetupTimeout: 0,
		}
		timeout := c.setupTimeout()
		assert.Equal(t, 5*time.Second, timeout, "zero timeout should return default 5s")
	})

	t.Run("negative timeout returns default", func(t *testing.T) {
		c := &Config{
			SetupTimeout: -1 * time.Second,
		}
		timeout := c.setupTimeout()
		assert.Equal(t, 5*time.Second, timeout, "negative timeout should return default 5s")
	})

	t.Run("positive timeout returns configured value", func(t *testing.T) {
		c := &Config{
			SetupTimeout: 30 * time.Second,
		}
		timeout := c.setupTimeout()
		assert.Equal(t, 30*time.Second, timeout, "should return configured timeout")
	})
}

// TestConfig_setupTimeout_Boundary verifies timeout edge cases
func TestConfig_setupTimeout_Boundary(t *testing.T) {
	t.Run("very small positive timeout", func(t *testing.T) {
		c := &Config{
			SetupTimeout: 1 * time.Millisecond,
		}
		timeout := c.setupTimeout()
		assert.Equal(t, 1*time.Millisecond, timeout, "should accept very small timeout")
	})

	t.Run("very large timeout", func(t *testing.T) {
		c := &Config{
			SetupTimeout: 5 * time.Minute,
		}
		timeout := c.setupTimeout()
		assert.Equal(t, 5*time.Minute, timeout, "should accept large timeout")
	})
}

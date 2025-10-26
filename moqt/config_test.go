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

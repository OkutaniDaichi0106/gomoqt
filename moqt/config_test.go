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
				ClientSetupExtensions: func() *Parameters {
					return &Parameters{}
				},
				ServerSetupExtensions: func(clientParams *Parameters) (serverParams *Parameters, err error) {
					return &Parameters{}, nil
				},

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
			assert.Equal(t, original.ClientSetupExtensions == nil, cloned.ClientSetupExtensions == nil, "ClientSetupExtensions nil status should be equal")
			assert.Equal(t, original.ServerSetupExtensions == nil, cloned.ServerSetupExtensions == nil, "ServerSetupExtensions nil status should be equal")
			assert.Equal(t, original.SetupTimeout, cloned.SetupTimeout, "Timeout should be equal")
		})
	}
}

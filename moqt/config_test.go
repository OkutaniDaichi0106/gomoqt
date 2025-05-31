package moqt_test

import (
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Clone(t *testing.T) {
	// Create original config with various settings
	originalConfig := &moqt.Config{
		ClientSetupExtensions: func() *moqt.Parameters {
			return &moqt.Parameters{}
		},
		ServerSetupExtensions: func(clientParams *moqt.Parameters) (serverParams *moqt.Parameters, err error) {
			return &moqt.Parameters{}, nil
		},
		Tracer: func() *moqtrace.SessionTracer {
			return &moqtrace.SessionTracer{}
		},
		Timeout: 30 * time.Second,
	}

	// Clone the config
	clonedConfig := originalConfig.Clone()

	// Verify the clone is not the same object
	assert.NotSame(t, originalConfig, clonedConfig)

	// Verify the fields are correctly copied
	assert.Equal(t, originalConfig.Timeout, clonedConfig.Timeout)
	assert.NotNil(t, clonedConfig.ClientSetupExtensions)
	assert.NotNil(t, clonedConfig.ServerSetupExtensions)
	assert.NotNil(t, clonedConfig.Tracer)

	// Test that function fields work correctly
	params := clonedConfig.ClientSetupExtensions()
	assert.NotNil(t, params)

	rspParams, err := clonedConfig.ServerSetupExtensions(&moqt.Parameters{})
	assert.NoError(t, err)
	assert.NotNil(t, rspParams)

	tracer := clonedConfig.Tracer()
	assert.Nil(t, tracer) // Our test tracer returns nil
}

func TestConfig_CloneWithNilFields(t *testing.T) {
	// Test cloning config with nil function fields
	originalConfig := &moqt.Config{
		Timeout: 10 * time.Second,
	}

	clonedConfig := originalConfig.Clone()

	assert.NotSame(t, originalConfig, clonedConfig)
	assert.Equal(t, originalConfig.Timeout, clonedConfig.Timeout)
	assert.Nil(t, clonedConfig.ClientSetupExtensions)
	assert.Nil(t, clonedConfig.ServerSetupExtensions)
	assert.Nil(t, clonedConfig.Tracer)
}

func TestConfig_CloneZeroValues(t *testing.T) {
	// Test config with zero values
	config := &moqt.Config{}
	clonedConfig := config.Clone()

	assert.NotSame(t, config, clonedConfig)
	assert.Zero(t, clonedConfig.Timeout)
	assert.Nil(t, clonedConfig.ClientSetupExtensions)
	assert.Nil(t, clonedConfig.ServerSetupExtensions)
	assert.Nil(t, clonedConfig.Tracer)
}

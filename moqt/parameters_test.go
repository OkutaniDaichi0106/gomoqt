package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParameters(t *testing.T) {
	params := NewExtension()
	assert.NotNil(t, params, "NewParameters should return a non-nil value")
}

func TestParameters_String(t *testing.T) {
	tests := map[string]struct {
		setup    func() *Extension
		expected string
	}{"empty parameters": {
		setup: func() *Extension {
			return NewExtension()
		},
		expected: "{ }",
	},
		"with one parameter": {
			setup: func() *Extension {
				p := NewExtension()
				p.SetString(1, "test")
				return p
			},
			expected: "{ 1: [116 101 115 116], }",
		},
		"with multiple parameters": {
			setup: func() *Extension {
				p := NewExtension()
				p.SetString(1, "test1")
				p.SetString(2, "test2")
				p.SetUint(3, 42)
				return p
			},
			// The order of parameters in the string representation might vary
			// so we just check that it contains all the expected parts
			expected: "{",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.setup()
			result := params.String()

			if name == "with multiple parameters" {
				assert.Contains(t, result, "{")
				assert.Contains(t, result, "1: [116 101 115 116 49]")
				assert.Contains(t, result, "2: [116 101 115 116 50]")
				assert.Contains(t, result, "3: [42]")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParameters_ByteArrayParameter(t *testing.T) {
	tests := map[string]struct {
		setup         func() *Extension
		key           ExtensionKey
		value         []byte
		expectedError bool
	}{
		"set and get byte array": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         []byte{1, 2, 3, 4},
			expectedError: false,
		},
		"get non-existent key": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         nil,
			expectedError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.setup()

			if tt.value != nil {
				params.SetByteArray(tt.key, tt.value)
			}

			result, err := params.GetByteArray(tt.key)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, ErrParameterNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, result)
			}
		})
	}
}

func TestParameters_StringParameter(t *testing.T) {
	tests := map[string]struct {
		setup         func() *Extension
		key           ExtensionKey
		value         string
		expectedError bool
	}{
		"set and get string": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         "test string",
			expectedError: false,
		},
		"get non-existent key": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         "",
			expectedError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.setup()

			if tt.value != "" || !tt.expectedError {
				params.SetString(tt.key, tt.value)
			}

			result, err := params.GetString(tt.key)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, ErrParameterNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, result)
			}
		})
	}
}

func TestParameters_UintParameter(t *testing.T) {
	tests := map[string]struct {
		setup         func() *Extension
		key           ExtensionKey
		value         uint64
		expectedError bool
	}{
		"set and get uint": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         42,
			expectedError: false,
		},
		"set and get large uint": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         1<<32 - 1, // max uint32
			expectedError: false,
		},
		"get non-existent key": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         0,
			expectedError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.setup()

			if !tt.expectedError {
				params.SetUint(tt.key, tt.value)
			}

			result, err := params.GetUint(tt.key)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, ErrParameterNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, result)
			}
		})
	}
}

func TestParameters_BoolParameter(t *testing.T) {
	tests := map[string]struct {
		setup         func() *Extension
		key           ExtensionKey
		value         bool
		expectedError bool
	}{
		"set and get true": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         true,
			expectedError: false,
		},
		"set and get false": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         false,
			expectedError: false,
		},
		"get non-existent key": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:           1,
			value:         false,
			expectedError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.setup()

			if !tt.expectedError {
				params.SetBool(tt.key, tt.value)
			}

			result, err := params.GetBool(tt.key)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, ErrParameterNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, result)
			}
		})
	}
}

func TestParameters_Remove(t *testing.T) {
	tests := map[string]struct {
		setup    func() *Extension
		key      ExtensionKey
		expected bool
	}{
		"remove existing key": {
			setup: func() *Extension {
				p := NewExtension()
				p.SetString(1, "test")
				return p
			},
			key:      1,
			expected: true,
		},
		"remove non-existent key": {
			setup: func() *Extension {
				return NewExtension()
			},
			key:      1,
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.setup()

			// Check if key exists before removal
			_, beforeErr := params.GetByteArray(tt.key)
			beforeExists := beforeErr == nil

			// Remove the key
			params.Remove(tt.key)

			// Check if key exists after removal
			_, afterErr := params.GetByteArray(tt.key)
			afterExists := afterErr == nil

			assert.Equal(t, tt.expected, beforeExists, "Key existence before removal should match expected")
			assert.False(t, afterExists, "Key should not exist after removal")
		})
	}
}

func TestParameters_GetBool_InvalidValue(t *testing.T) {
	tests := map[string]struct {
		value         uint64
		expectedError string
	}{
		"value 2": {
			value:         2,
			expectedError: "invalid value as bool",
		},
		"value 255": {
			value:         255,
			expectedError: "invalid value as bool",
		},
		"value 100": {
			value:         100,
			expectedError: "invalid value as bool",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := NewExtension()
			params.SetUint(1, tt.value)

			_, err := params.GetBool(1)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestParameters_NilMap(t *testing.T) {
	// Create a Parameters with nil map
	var params Extension

	// Test all getters with nil map
	_, err := params.GetByteArray(1)
	assert.Equal(t, ErrParameterNotFound, err)

	_, err = params.GetString(1)
	assert.Equal(t, ErrParameterNotFound, err)

	_, err = params.GetUint(1)
	assert.Equal(t, ErrParameterNotFound, err)

	_, err = params.GetBool(1)
	assert.Equal(t, ErrParameterNotFound, err)

	// Test all setters with nil map (should initialize the map)
	params.SetByteArray(1, []byte{1, 2, 3})
	params.SetString(2, "test")
	params.SetUint(4, 42)
	params.SetBool(5, true)

	// Verify the values were set correctly
	val1, err := params.GetByteArray(1)
	assert.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, val1)

	val2, err := params.GetString(2)
	assert.NoError(t, err)
	assert.Equal(t, "test", val2)

	val4, err := params.GetUint(4)
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), val4)

	val5, err := params.GetBool(5)
	assert.NoError(t, err)
	assert.Equal(t, true, val5)
}

func TestParameters_Clone(t *testing.T) {
	original := NewExtension()
	original.SetString(1, "test")
	original.SetUint(2, 42)
	original.SetBool(3, true)

	cloned := original.Clone()
	assert.NotNil(t, cloned, "Clone should return a non-nil value")
	assert.NotSame(t, original, cloned, "Clone should return a different instance")

	// Check that the cloned parameters have the same values
	val1, err := cloned.GetString(1)
	assert.NoError(t, err)
	assert.Equal(t, "test", val1)

	val2, err := cloned.GetUint(2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), val2)

	val3, err := cloned.GetBool(3)
	assert.NoError(t, err)
	assert.Equal(t, true, val3)

	// Modify the original and ensure the clone is unaffected
	original.SetString(1, "modified")
	val1Original, _ := original.GetString(1)
	val1Cloned, _ := cloned.GetString(1)
	assert.Equal(t, "modified", val1Original)
	assert.Equal(t, "test", val1Cloned)
}

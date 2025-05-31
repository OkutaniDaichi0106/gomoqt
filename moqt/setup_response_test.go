package moqt

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
)

func TestSetupResponse_String(t *testing.T) {
	tests := []struct {
		name     string
		response SetupResponse
		expected string
	}{
		{
			name: "setup response with parameters",
			response: SetupResponse{
				Parameters:      &Parameters{},
				selectedVersion: protocol.Version(1),
			},
			expected: "SetupResponse: { SelectedVersion: 1, Parameters:  }",
		},
		{
			name: "setup response with different version",
			response: SetupResponse{
				Parameters:      &Parameters{},
				selectedVersion: protocol.Version(42),
			},
			expected: "SetupResponse: { SelectedVersion: 42, Parameters:  }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.String()
			assert.Contains(t, result, "SetupResponse:")
			assert.Contains(t, result, "SelectedVersion:")
			assert.Contains(t, result, "Parameters:")
		})
	}
}

func TestSetupResponseFields(t *testing.T) {
	params := &Parameters{}
	version := protocol.Version(5)

	response := SetupResponse{
		Parameters:      params,
		selectedVersion: version,
	}

	assert.Equal(t, params, response.Parameters)
	assert.Equal(t, version, response.selectedVersion)
}

func TestSetupResponseWithNilParameters(t *testing.T) {
	response := SetupResponse{
		Parameters:      nil,
		selectedVersion: protocol.Version(1),
	}

	// This should not panic when calling String()
	str := response.String()
	assert.Contains(t, str, "SetupResponse:")
}

package moqt

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSetupRequest(t *testing.T) {
	tests := []struct {
		name        string
		urlStr      string
		expectError bool
		expectedURL string
	}{
		{
			name:        "valid http URL",
			urlStr:      "http://example.com",
			expectError: false,
			expectedURL: "http://example.com",
		},
		{
			name:        "valid https URL",
			urlStr:      "https://example.com:8080/path",
			expectError: false,
			expectedURL: "https://example.com:8080/path",
		},
		{
			name:        "valid URL with query parameters",
			urlStr:      "https://example.com/path?param=value",
			expectError: false,
			expectedURL: "https://example.com/path?param=value",
		},
		{
			name:        "invalid URL",
			urlStr:      "://invalid-url",
			expectError: true,
		},
		{
			name:        "empty URL",
			urlStr:      "",
			expectError: false,
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewSetupRequest(tt.urlStr)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, req)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)
				assert.Equal(t, tt.expectedURL, req.URI.String())
				assert.Nil(t, req.Parameters)
			}
		})
	}
}

func TestSetupRequest_String(t *testing.T) {
	tests := []struct {
		name     string
		req      SetupRequest
		expected string
	}{
		{
			name: "setup request with no parameters",
			req: SetupRequest{
				URI:        mustParseURL("https://example.com"),
				Parameters: nil,
			},
			expected: "SetupRequest: { URI: https://example.com, Parameters: No Parameters }",
		},
		{
			name: "setup request with parameters",
			req: SetupRequest{
				URI:        mustParseURL("https://example.com"),
				Parameters: &Parameters{}, // Assuming Parameters has a String() method
			},
			expected: "SetupRequest: { URI: https://example.com, Parameters:  }", // Empty parameters string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.req.String()
			assert.Contains(t, result, "SetupRequest:")
			assert.Contains(t, result, "URI: https://example.com")
			if tt.req.Parameters == nil {
				assert.Contains(t, result, "No Parameters")
			} else {
				assert.Contains(t, result, "Parameters:")
			}
		})
	}
}

func TestSetupRequestFields(t *testing.T) {
	uri := mustParseURL("https://example.com/test")
	params := &Parameters{}

	req := &SetupRequest{
		URI:        uri,
		Parameters: params,
	}

	assert.Equal(t, uri, req.URI)
	assert.Equal(t, params, req.Parameters)
}

// Helper function for tests
func mustParseURL(urlStr string) *url.URL {
	uri, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return uri
}

package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionClientMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SessionClientMessage
		wantErr bool
	}{
		"valid message": {
			input: message.SessionClientMessage{
				SupportedVersions: []protocol.Version{0, 1},
				Parameters: message.Parameters{
					1: []byte("value1"),
					2: []byte("value2"),
				},
			},
		},
		"empty parameters": {
			input: message.SessionClientMessage{
				SupportedVersions: []protocol.Version{0},
				Parameters:        message.Parameters{},
			},
		},
		"max values": {
			input: message.SessionClientMessage{
				SupportedVersions: []protocol.Version{protocol.Version(^byte(0))},
				Parameters: message.Parameters{
					^uint64(0): bytes.Repeat([]byte("a"), 1024),
				},
			},
		},
		"nil values": {
			input: message.SessionClientMessage{
				SupportedVersions: []protocol.Version{0},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer

			// Encode
			en, err := tc.input.Encode(&buf)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Decode
			var decoded message.SessionClientMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

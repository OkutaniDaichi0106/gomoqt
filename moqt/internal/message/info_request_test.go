package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoRequestMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.InfoRequestMessage
		wantErr bool
	}{
		"valid message": {
			input: message.InfoRequestMessage{
				TrackPath: "path/to/track",
			},
		},
		"empty track path": {
			input: message.InfoRequestMessage{
				TrackPath: "",
			},
		},
		"long track path": {
			input: message.InfoRequestMessage{
				TrackPath: "very/long/path/with/many/segments",
			},
		},
		// "nil track path": {
		// 	input: message.InfoRequestMessage{},
		// },
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
			var decoded message.InfoRequestMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

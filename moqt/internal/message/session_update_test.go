package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionUpdateMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SessionUpdateMessage
		wantErr bool
	}{
		"valid bitrate": {
			input: message.SessionUpdateMessage{
				Bitrate: 12345,
			},
		},
		"zero bitrate": {
			input: message.SessionUpdateMessage{
				Bitrate: 0,
			},
		},
		// "max bitrate": {
		// 	input: message.SessionUpdateMessage{
		// 		Bitrate: ^uint64(0),
		// 	},
		// },
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			// Encode
			en, err := tc.input.Encode(&buf)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Decode
			var decoded message.SessionUpdateMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

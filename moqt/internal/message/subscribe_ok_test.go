package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeOkMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SubscribeOkMessage
		wantErr bool
	}{
		"valid message with 0 group frequency": {
			input: message.SubscribeOkMessage{
				GroupPeriod: 0,
			},
		},
		"valid message with valid group frequency": {
			input: message.SubscribeOkMessage{
				GroupPeriod: message.GroupPeriod(1),
			},
		},
		"valid message with big group frequency": {
			input: message.SubscribeOkMessage{
				GroupPeriod: message.GroupPeriod(255),
			},
		},
		"zero value": {
			input: message.SubscribeOkMessage{
				GroupPeriod: 0,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer

			// Encode
			err := tc.input.Encode(&buf)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Decode
			var decoded message.SubscribeOkMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

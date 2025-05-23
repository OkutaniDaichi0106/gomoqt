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
		"valid message with default group order": {
			input: message.SubscribeOkMessage{
				GroupOrder: message.GroupOrderDefault,
			},
		},
		"valid message with ascending group order": {
			input: message.SubscribeOkMessage{
				GroupOrder: message.GroupOrderAscending,
			},
		},
		"valid message with descending group order": {
			input: message.SubscribeOkMessage{
				GroupOrder: message.GroupOrderDescending,
			},
		},
		"zero value": {
			input: message.SubscribeOkMessage{
				GroupOrder: 0,
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
			var decoded message.SubscribeOkMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

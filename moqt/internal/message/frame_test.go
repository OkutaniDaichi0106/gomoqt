package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.FrameMessage
		wantErr bool
	}{
		"valid payload": {
			input: message.FrameMessage{
				Payload: []byte{1, 2},
			},
		},
		"empty payload": {
			input: message.FrameMessage{
				Payload: []byte{},
			},
		},
		"string payload": {
			input: message.FrameMessage{
				Payload: []byte("bar"),
			},
		},
		"large payload": {
			input: message.FrameMessage{
				Payload: bytes.Repeat([]byte("a"), 1024),
			},
		},
		"nil payload": {
			input: message.FrameMessage{
				Payload: nil,
			},
		},
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
			var decoded message.FrameMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

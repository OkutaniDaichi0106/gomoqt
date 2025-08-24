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
			input: []byte{1, 2},
		},
		"empty payload": {
			input: []byte{},
		},
		"string payload": {
			input: []byte("bar"),
		},
		"large payload": {
			input: bytes.Repeat([]byte("a"), 1024),
		},
		// "nil payload": {
		// 	input: message.FrameMessage{
		// 		Payload: nil,
		// 	},
		// },
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			// Encode
			err := tc.input.Encode(&buf)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Decode
			decoded := message.FrameMessage{}
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

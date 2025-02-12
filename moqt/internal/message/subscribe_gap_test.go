package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeGapMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SubscribeGapMessage
		wantErr bool
	}{
		"valid message": {
			input: message.SubscribeGapMessage{
				GapStartSequence: 100,
				GapCount:         200,
				GroupErrorCode:   1,
			},
		},
		"zero values": {
			input: message.SubscribeGapMessage{
				GapStartSequence: 0,
				GapCount:         0,
				GroupErrorCode:   0,
			},
		},
		// "max values": {
		// 	input: message.SubscribeGapMessage{
		// 		MinGapSequence: message.GroupSequence(^uint64(0)),
		// 		MaxGapSequence: message.GroupSequence(^uint64(0)),
		// 		GroupErrorCode: message.GroupErrorCode(^uint32(0)),
		// 	},
		// },
		"min greater than max": {
			input: message.SubscribeGapMessage{
				GapStartSequence: 200,
				GapCount:         100,
				GroupErrorCode:   1,
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
			var decoded message.SubscribeGapMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

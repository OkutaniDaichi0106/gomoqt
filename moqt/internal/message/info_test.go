package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.InfoMessage
		wantErr bool
	}{
		"valid message": {
			input: message.InfoMessage{
				TrackPriority:       1,
				LatestGroupSequence: 2,
				GroupOrder:          3,
			},
		},
		"zero values": {
			input: message.InfoMessage{
				TrackPriority:       0,
				LatestGroupSequence: 0,
				GroupOrder:          0,
			},
		},
		"max values": {
			input: message.InfoMessage{
				TrackPriority:       message.TrackPriority(^byte(0)),
				LatestGroupSequence: message.GroupSequence(^uint64(0)),
				GroupOrder:          message.GroupOrder(^byte(0)),
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
			var decoded message.InfoMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

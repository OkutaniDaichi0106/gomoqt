package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeUpdateMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SubscribeUpdateMessage
		wantErr bool
	}{
		"valid message": {
			input: message.SubscribeUpdateMessage{
				TrackPriority:             5,
				MinGroupSequence:          10,
				MaxGroupSequence:          20,
				GroupOrder:                1,
				SubscribeUpdateParameters: message.Parameters{1: []byte("value1")},
			},
		},
		"empty parameters": {
			input: message.SubscribeUpdateMessage{
				TrackPriority:    5,
				MinGroupSequence: 10,
				MaxGroupSequence: 20,
				GroupOrder:       1,
			},
		},
		"max values": {
			input: message.SubscribeUpdateMessage{
				TrackPriority:             message.TrackPriority(^byte(0)),
				MinGroupSequence:          message.GroupSequence(^uint64(0)),
				MaxGroupSequence:          message.GroupSequence(^uint64(0)),
				GroupOrder:                message.GroupOrder(^byte(0)),
				SubscribeUpdateParameters: message.Parameters{^uint64(0): bytes.Repeat([]byte("a"), 1024)},
			},
		},
		"zero values": {
			input: message.SubscribeUpdateMessage{
				TrackPriority:    0,
				MinGroupSequence: 0,
				MaxGroupSequence: 0,
				GroupOrder:       0,
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
			var decoded message.SubscribeUpdateMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

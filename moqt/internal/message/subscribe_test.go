package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SubscribeMessage
		wantErr bool
	}{
		"valid message": {
			input: message.SubscribeMessage{
				SubscribeID:         1,
				TrackPath:           []string{"path", "to", "track"},
				TrackPriority:       5,
				GroupOrder:          1,
				MinGroupSequence:    10,
				MaxGroupSequence:    20,
				SubscribeParameters: message.Parameters{1: []byte("value")},
			},
		},
		"empty track path": {
			input: message.SubscribeMessage{
				SubscribeID:         1,
				TrackPath:           []string{},
				TrackPriority:       5,
				GroupOrder:          1,
				MinGroupSequence:    10,
				MaxGroupSequence:    20,
				SubscribeParameters: message.Parameters{1: []byte("value")},
			},
		},
		"max values": {
			input: message.SubscribeMessage{
				SubscribeID:         message.SubscribeID(^uint64(0)),
				TrackPath:           []string{"very", "long", "path"},
				TrackPriority:       message.TrackPriority(^byte(0)),
				GroupOrder:          message.GroupOrder(^byte(0)),
				MinGroupSequence:    message.GroupSequence(^uint64(0)),
				MaxGroupSequence:    message.GroupSequence(^uint64(0)),
				SubscribeParameters: message.Parameters{1: bytes.Repeat([]byte("a"), 1024)},
			},
		},
		"nil parameters": {
			input: message.SubscribeMessage{
				SubscribeID:   1,
				TrackPath:     []string{"path"},
				TrackPriority: 1,
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
			var decoded message.SubscribeMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

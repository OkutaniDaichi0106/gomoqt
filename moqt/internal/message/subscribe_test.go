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
				SubscribeID:      1,
				BroadcastPath:    "path/to/track",
				TrackPriority:    5,
				MinGroupSequence: 10,
				MaxGroupSequence: 20,
			},
		},
		"empty track path": {
			input: message.SubscribeMessage{
				SubscribeID:      1,
				BroadcastPath:    "",
				TrackPriority:    5,
				MinGroupSequence: 10,
				MaxGroupSequence: 20,
			},
		},
		"nil parameters": {
			input: message.SubscribeMessage{
				SubscribeID:   1,
				BroadcastPath: "path",
				TrackPriority: 1,
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
			var decoded message.SubscribeMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.FetchMessage
		wantErr bool
	}{
		"valid message": {
			input: message.FetchMessage{
				SubscribeID:   1,
				TrackPath:     []string{"path", "to", "track"},
				TrackPriority: 5,
				GroupSequence: 10,
				FrameSequence: 15,
			},
		},
		"empty track path": {
			input: message.FetchMessage{
				SubscribeID:   1,
				TrackPath:     []string{},
				TrackPriority: 5,
				GroupSequence: 10,
				FrameSequence: 15,
			},
		},
		"max values": {
			input: message.FetchMessage{
				SubscribeID:   message.SubscribeID(^uint64(0)),
				TrackPath:     []string{"very", "long", "path"},
				TrackPriority: message.TrackPriority(^byte(0)),
				GroupSequence: message.GroupSequence(^uint64(0)),
				FrameSequence: message.FrameSequence(^uint64(0)),
			},
		},
		"nil track path": {
			input: message.FetchMessage{
				SubscribeID:   1,
				TrackPriority: 1,
				GroupSequence: 1,
				FrameSequence: 1,
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
			var decoded message.FetchMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnounceMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.AnnounceMessage
		wantErr bool
	}{
		"valid message": {
			input: message.AnnounceMessage{
				AnnounceStatus: message.AnnounceStatus(1),
				TrackSuffix:    "path/to/track",
			},
		},
		"empty wildcard parameters": {
			input: message.AnnounceMessage{
				AnnounceStatus: message.AnnounceStatus(1),
				TrackSuffix:    "",
			},
		},
		// "max values": {
		// 	input: message.AnnounceMessage{
		// 		AnnounceStatus:  message.AnnounceStatus(^byte(0)),
		// 		TrackPathSuffix: []string{"very", "long", "path"},
		// 		Parameters: map[uint64][]byte{
		// 			^uint64(0): bytes.Repeat([]byte("a"), 1024),
		// 		},
		// 	},
		// },
		// "nil wildcard parameters": {
		// 	input: message.AnnounceMessage{
		// 		AnnounceStatus:     message.AnnounceStatus(1),
		// 		WildcardParameters: nil,
		// 	},
		// },
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
			var decoded message.AnnounceMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

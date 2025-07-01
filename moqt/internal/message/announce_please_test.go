package message_test

import (
	"bytes"
	"testing" // Encode

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnouncePleaseMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.AnnouncePleaseMessage
		wantErr bool
	}{
		"valid message": {
			input: message.AnnouncePleaseMessage{
				TrackPrefix: "part1/part2",
				// AnnounceParameters: message.Parameters{0: []byte("value")},
			},
		},
		"empty track prefix": {
			input: message.AnnouncePleaseMessage{
				TrackPrefix: "",
				// AnnounceParameters: message.Parameters{1: []byte("value")},
			},
		},
		"empty parameters": {
			input: message.AnnouncePleaseMessage{
				TrackPrefix: "path",
				// AnnounceParameters: message.Parameters{},
			},
		},
		"long path": {
			input: message.AnnouncePleaseMessage{
				TrackPrefix: "very/long/path/with/many/segments",
				// AnnounceParameters: message.Parameters{1: []byte("value")},
			},
		},
		// "large parameter": {
		// 	input: message.AnnouncePleaseMessage{
		// 		TrackPathPrefix: []string{"path"},
		// 		Parameters:      message.Parameters{^uint64(0): bytes.Repeat([]byte("a"), 1024)},
		// 	},
		// },
		// "nil values": {
		// 	input: message.AnnouncePleaseMessage{},
		// },
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
			var decoded message.AnnouncePleaseMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

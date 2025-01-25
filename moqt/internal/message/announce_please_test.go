package message_test

import (
	"bytes"
	"testing"

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
				TrackPathPrefix: []string{"part1", "part2"},
				Parameters:      message.Parameters{0: []byte("value")},
			},
		},
		"empty track prefix": {
			input: message.AnnouncePleaseMessage{
				TrackPathPrefix: []string{},
				Parameters:      message.Parameters{1: []byte("value")},
			},
		},
		"empty parameters": {
			input: message.AnnouncePleaseMessage{
				TrackPathPrefix: []string{"path"},
				Parameters:      message.Parameters{},
			},
		},
		"long path": {
			input: message.AnnouncePleaseMessage{
				TrackPathPrefix: []string{"very", "long", "path", "with", "many", "segments"},
				Parameters:      message.Parameters{1: []byte("value")},
			},
		},
		"large parameter": {
			input: message.AnnouncePleaseMessage{
				TrackPathPrefix: []string{"path"},
				Parameters:      message.Parameters{^uint64(0): bytes.Repeat([]byte("a"), 1024)},
			},
		},
		"nil values": {
			input: message.AnnouncePleaseMessage{},
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
			var decoded message.AnnouncePleaseMessage
			dn, err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
			assert.Equal(t, en, dn, "encoded and decoded message should have the same length")
		})
	}
}

package message_test

import (
	"bytes"
	"testing" // Encode

	"github.com/okdaichi/gomoqt/moqt/internal/message"
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

func TestAnnouncePleaseMessage_DecodeErrors(t *testing.T) {
	t.Run("read message length error", func(t *testing.T) {
		var aim message.AnnouncePleaseMessage
		src := bytes.NewReader([]byte{})
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read full error", func(t *testing.T) {
		var aim message.AnnouncePleaseMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 10)
		buf.WriteByte(0x00)
		src := bytes.NewReader(buf.Bytes()[:2])
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read string error", func(t *testing.T) {
		var aim message.AnnouncePleaseMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 1)
		buf.WriteByte(0x00)
		buf.WriteByte(0x80) // invalid string
		src := bytes.NewReader(buf.Bytes())
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra data", func(t *testing.T) {
		var aim message.AnnouncePleaseMessage
		var buf bytes.Buffer
		buf.WriteByte(0x04) // length varint = 4
		buf.WriteByte(0x01) // string length 1
		buf.WriteByte('a')
		buf.WriteByte(0x00) // extra byte 1
		buf.WriteByte(0x00) // extra byte 2 (to fill 4 bytes)
		src := bytes.NewReader(buf.Bytes())
		err := aim.Decode(src)
		assert.Error(t, err)
		assert.Equal(t, message.ErrMessageTooShort, err)
	})
}

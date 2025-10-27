package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
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
		"max values": {
			input: message.AnnounceMessage{
				AnnounceStatus: message.AnnounceStatus(^byte(0)),
				TrackSuffix:    "very/long/path",
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
			var decoded message.AnnounceMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

func TestAnnounceMessage_DecodeErrors(t *testing.T) {
	t.Run("read message length error", func(t *testing.T) {
		var am message.AnnounceMessage
		src := bytes.NewReader([]byte{}) // empty, should cause error
		err := am.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read full error", func(t *testing.T) {
		var am message.AnnounceMessage
		// Write length but not enough data
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 10) // varint for 10
		buf.WriteByte(0x00)
		src := bytes.NewReader(buf.Bytes()[:2]) // only 2 bytes, but length says 10
		err := am.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read varint error", func(t *testing.T) {
		var am message.AnnounceMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 1) // length 1
		buf.WriteByte(0x00)
		buf.WriteByte(0x80) // invalid varint
		src := bytes.NewReader(buf.Bytes())
		err := am.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read string error", func(t *testing.T) {
		var am message.AnnounceMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 2) // length 2
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // status
		buf.WriteByte(0x80) // invalid string varint
		src := bytes.NewReader(buf.Bytes())
		err := am.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra data", func(t *testing.T) {
		var am message.AnnounceMessage
		// Manually construct data with extra bytes after valid data
		var buf bytes.Buffer
		buf.WriteByte(0x04) // length 4
		buf.WriteByte(0x01) // status
		buf.WriteByte(0x01) // string length 1
		buf.WriteByte('a')  // string
		buf.WriteByte(0x00) // extra byte
		src := bytes.NewReader(buf.Bytes())
		err := am.Decode(src)
		assert.Error(t, err)
		assert.Equal(t, message.ErrMessageTooShort, err)
	})
}

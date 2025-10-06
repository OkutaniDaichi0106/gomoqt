package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionUpdateMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SessionUpdateMessage
		wantErr bool
	}{
		"valid bitrate": {
			input: message.SessionUpdateMessage{
				Bitrate: 12345,
			},
		},
		"zero bitrate": {
			input: message.SessionUpdateMessage{
				Bitrate: 0,
			},
		},
		// "max bitrate": {
		// 	input: message.SessionUpdateMessage{
		// 		Bitrate: ^uint64(0),
		// 	},
		// },
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			// Encode
			err := tc.input.Encode(&buf)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Decode
			var decoded message.SessionUpdateMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

func TestSessionUpdateMessage_DecodeErrors(t *testing.T) {
	t.Run("read message length error", func(t *testing.T) {
		var sum message.SessionUpdateMessage
		src := bytes.NewReader([]byte{})
		err := sum.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read full error", func(t *testing.T) {
		var sum message.SessionUpdateMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 10)
		buf.WriteByte(0x00)
		src := bytes.NewReader(buf.Bytes()[:2])
		err := sum.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read varint error for bitrate", func(t *testing.T) {
		var sum message.SessionUpdateMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 1)
		buf.WriteByte(0x00)
		buf.WriteByte(0x80) // invalid varint
		src := bytes.NewReader(buf.Bytes())
		err := sum.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra data", func(t *testing.T) {
		var sum message.SessionUpdateMessage
		var buf bytes.Buffer
		buf.WriteByte(0x03)
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // bitrate
		buf.WriteByte(0x00) // extra
		src := bytes.NewReader(buf.Bytes())
		err := sum.Decode(src)
		assert.Error(t, err)
		assert.Equal(t, message.ErrMessageTooShort, err)
	})
}

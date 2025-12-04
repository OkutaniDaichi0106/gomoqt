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
				TrackPriority: 5,
			},
		},
		"empty parameters": {
			input: message.SubscribeUpdateMessage{
				TrackPriority: 5,
			},
		},
		"max values": {
			input: message.SubscribeUpdateMessage{
				TrackPriority: 255,
			},
		},
		"zero values": {
			input: message.SubscribeUpdateMessage{
				TrackPriority: 0,
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
			var decoded message.SubscribeUpdateMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare all fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

func TestSubscribeUpdateMessage_DecodeErrors(t *testing.T) {
	t.Run("read message length error", func(t *testing.T) {
		var sum message.SubscribeUpdateMessage
		src := bytes.NewReader([]byte{})
		err := sum.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read full error", func(t *testing.T) {
		var sum message.SubscribeUpdateMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 10)
		buf.WriteByte(0x00)
		src := bytes.NewReader(buf.Bytes()[:2])
		err := sum.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read varint error for track priority", func(t *testing.T) {
		var sum message.SubscribeUpdateMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 1)
		buf.WriteByte(0x00)
		buf.WriteByte(0x80) // invalid varint
		src := bytes.NewReader(buf.Bytes())
		err := sum.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra data", func(t *testing.T) {
		var sum message.SubscribeUpdateMessage
		var buf bytes.Buffer
		buf.WriteByte(0x00) // length u16 high byte
		buf.WriteByte(0x02) // length u16 low byte = 2
		buf.WriteByte(0x01) // track priority
		buf.WriteByte(0x00) // extra
		src := bytes.NewReader(buf.Bytes())
		err := sum.Decode(src)
		assert.Error(t, err)
		assert.Equal(t, message.ErrMessageTooShort, err)
	})
}

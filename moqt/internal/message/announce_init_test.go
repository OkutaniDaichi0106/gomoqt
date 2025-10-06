package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnounceInitMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.AnnounceInitMessage
		wantErr bool
	}{
		"valid message": {
			input: message.AnnounceInitMessage{
				Suffixes: []string{"suffix1", "suffix2"},
			},
		},
		"empty suffixes": {
			input: message.AnnounceInitMessage{
				Suffixes: []string{},
			},
		},
		"single suffix": {
			input: message.AnnounceInitMessage{
				Suffixes: []string{"onlyone"},
			},
		},
		"long suffix": {
			input: message.AnnounceInitMessage{
				Suffixes: []string{"very/long/suffix/with/many/segments"},
			},
		},
		// "nil suffixes": {
		// 	input: message.AnnounceInitMessage{},
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
			var decoded message.AnnounceInitMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

func TestAnnounceInitMessage_DecodeErrors(t *testing.T) {
	t.Run("read message length error", func(t *testing.T) {
		var aim message.AnnounceInitMessage
		src := bytes.NewReader([]byte{}) // empty
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read full error", func(t *testing.T) {
		var aim message.AnnounceInitMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 10) // length 10
		buf.WriteByte(0x00)
		src := bytes.NewReader(buf.Bytes()[:2]) // incomplete
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read varint error", func(t *testing.T) {
		var aim message.AnnounceInitMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 1) // length 1
		buf.WriteByte(0x00)
		buf.WriteByte(0x80) // invalid varint
		src := bytes.NewReader(buf.Bytes())
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read string error", func(t *testing.T) {
		var aim message.AnnounceInitMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 3) // length 3
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // count 1
		buf.WriteByte(0x80) // invalid string varint
		src := bytes.NewReader(buf.Bytes())
		err := aim.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra data", func(t *testing.T) {
		var aim message.AnnounceInitMessage
		var buf bytes.Buffer
		buf.WriteByte(0x05) // length 5
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // count 1
		buf.WriteByte(0x01) // string length 1
		buf.WriteByte('a')  // string
		buf.WriteByte(0x00) // extra
		src := bytes.NewReader(buf.Bytes())
		err := aim.Decode(src)
		assert.Error(t, err)
		assert.Equal(t, message.ErrMessageTooShort, err)
	})
}

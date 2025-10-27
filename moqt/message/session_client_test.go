package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionClientMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.SessionClientMessage
		wantErr bool
	}{
		"valid message": {
			input: message.SessionClientMessage{
				SupportedVersions: []internal.Version{0, 1},
				Parameters: message.Parameters{
					1: []byte("value1"),
					2: []byte("value2"),
				},
			},
		},
		"empty parameters": {
			input: message.SessionClientMessage{
				SupportedVersions: []internal.Version{0},
				Parameters:        message.Parameters{},
			},
		},
		// "max values": {
		// 	input: message.SessionClientMessage{
		// 		SupportedVersions: []protocol.Version{protocol.Version(^byte(0))},
		// 		Parameters: message.Parameters{
		// 			^uint64(0): bytes.Repeat([]byte("a"), 1024),
		// 		},
		// 	},
		// },
		"nil values": {
			input: message.SessionClientMessage{
				SupportedVersions: []internal.Version{0},
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
			var decoded message.SessionClientMessage
			err = decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare SupportedVersions
			assert.Equal(t, tc.input.SupportedVersions, decoded.SupportedVersions, "SupportedVersions should match")

			// Compare Parameters (nil and empty map are treated as equivalent)
			if len(tc.input.Parameters) == 0 && len(decoded.Parameters) == 0 {
				// Both are empty, treat as equal
				return
			}
			assert.Equal(t, tc.input.Parameters, decoded.Parameters, "Parameters should match")
		})
	}
}

func TestSessionClientMessage_DecodeErrors(t *testing.T) {
	t.Run("read message length error", func(t *testing.T) {
		var scm message.SessionClientMessage
		src := bytes.NewReader([]byte{})
		err := scm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read full error", func(t *testing.T) {
		var scm message.SessionClientMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 10)
		buf.WriteByte(0x00)
		src := bytes.NewReader(buf.Bytes()[:2])
		err := scm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read varint error for count", func(t *testing.T) {
		var scm message.SessionClientMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 1)
		buf.WriteByte(0x00)
		buf.WriteByte(0x80) // invalid varint
		src := bytes.NewReader(buf.Bytes())
		err := scm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read varint error for version", func(t *testing.T) {
		var scm message.SessionClientMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 3)
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // count 1
		buf.WriteByte(0x80) // invalid varint for version
		src := bytes.NewReader(buf.Bytes())
		err := scm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("read parameters error", func(t *testing.T) {
		var scm message.SessionClientMessage
		var buf bytes.Buffer
		buf.WriteByte(0x80 | 3)
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // count 1
		buf.WriteByte(0x01) // version 1
		buf.WriteByte(0x80) // invalid parameters
		src := bytes.NewReader(buf.Bytes())
		err := scm.Decode(src)
		assert.Error(t, err)
	})

	t.Run("extra data", func(t *testing.T) {
		var scm message.SessionClientMessage
		var buf bytes.Buffer
		buf.WriteByte(0x05)
		buf.WriteByte(0x00)
		buf.WriteByte(0x01) // count 1
		buf.WriteByte(0x01) // version 1
		buf.WriteByte(0x00) // parameters count 0
		buf.WriteByte(0x00) // extra
		src := bytes.NewReader(buf.Bytes())
		err := scm.Decode(src)
		assert.Error(t, err)
		assert.Equal(t, message.ErrMessageTooShort, err)
	})
}

package message_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/qumo-dev/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		token []byte
	}{
		"empty token":  {token: nil},
		"opaque token": {token: []byte("eyJhbGciOiJFZERTQSJ9.e30.sig")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, message.AuthMessage{Token: tt.token}.Encode(&buf))

			var decoded message.AuthMessage
			require.NoError(t, decoded.Decode(&buf))

			assert.Equal(t, tt.token, decoded.Token)
			assert.Zero(t, buf.Len())
		})
	}
}

func TestAuthMessage_Decode_TrailingBytes(t *testing.T) {
	// Length 3: a one-byte token, then a byte the message does not have.
	src := bytes.NewReader([]byte{0x03, 0x01, 'a', 'b'})

	var decoded message.AuthMessage
	err := decoded.Decode(src)

	assert.ErrorIs(t, err, message.ErrMessageTooShort)
}

func TestAuthOkMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input message.AuthOkMessage
	}{
		"grants nothing": {
			input: message.AuthOkMessage{PublishPatterns: []string{}, SubscribePatterns: []string{}},
		},
		"subscribe only, never expires": {
			input: message.AuthOkMessage{
				PublishPatterns:   []string{},
				SubscribePatterns: []string{"/room/1/"},
			},
		},
		"both, with expiry": {
			input: message.AuthOkMessage{
				PublishPatterns:   []string{"/room/1/alice"},
				SubscribePatterns: []string{"/room/1/", "/lobby"},
				Expires:           60_000,
			},
		},
		"largest expiry": {
			input: message.AuthOkMessage{
				PublishPatterns:   []string{},
				SubscribePatterns: []string{},
				Expires:           message.MaxAuthExpires,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, tt.input.Encode(&buf))

			typ, err := message.ReadAuthReplyType(&buf)
			require.NoError(t, err)
			assert.Equal(t, message.AuthReplyOK, typ)

			var decoded message.AuthOkMessage
			require.NoError(t, decoded.Decode(&buf))

			assert.Equal(t, tt.input, decoded)
			assert.Zero(t, buf.Len())
		})
	}
}

func TestAuthOkMessage_Encode_ExpiresTooLate(t *testing.T) {
	var buf bytes.Buffer

	err := message.AuthOkMessage{Expires: message.MaxAuthExpires + 1}.Encode(&buf)

	assert.ErrorIs(t, err, message.ErrInvalidAuthMessage)
	assert.Zero(t, buf.Len())
}

func TestAuthOkMessage_Decode_ExpiresTooLate(t *testing.T) {
	// Body: no publish patterns, no subscribe patterns, Expires 2^53 as an
	// 8-byte varint.
	body := []byte{0x00, 0x00, 0xc0, 0x20, 0, 0, 0, 0, 0, 0}
	src := bytes.NewReader(append([]byte{byte(len(body))}, body...))

	var decoded message.AuthOkMessage
	err := decoded.Decode(src)

	assert.ErrorIs(t, err, message.ErrInvalidAuthMessage)
}

func TestAuthErrorMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input message.AuthErrorMessage
	}{
		"unauthorized, no reason": {input: message.AuthErrorMessage{Code: 0x2, Reason: ""}},
		"not supported":           {input: message.AuthErrorMessage{Code: 0x30, Reason: "grant too large"}},
		"longest reason":          {input: message.AuthErrorMessage{Code: 0x2, Reason: strings.Repeat("a", message.MaxAuthReasonLength)}},
		"largest code":            {input: message.AuthErrorMessage{Code: 1<<32 - 1, Reason: "x"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, tt.input.Encode(&buf))

			typ, err := message.ReadAuthReplyType(&buf)
			require.NoError(t, err)
			assert.Equal(t, message.AuthReplyError, typ)

			var decoded message.AuthErrorMessage
			require.NoError(t, decoded.Decode(&buf))

			assert.Equal(t, tt.input, decoded)
		})
	}
}

func TestAuthErrorMessage_Encode_Invalid(t *testing.T) {
	tests := map[string]struct {
		input message.AuthErrorMessage
	}{
		"code over 32 bits":   {input: message.AuthErrorMessage{Code: 1 << 32}},
		"reason too long":     {input: message.AuthErrorMessage{Reason: strings.Repeat("a", message.MaxAuthReasonLength+1)}},
		"reason is not UTF-8": {input: message.AuthErrorMessage{Reason: "\xff"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer

			err := tt.input.Encode(&buf)

			assert.ErrorIs(t, err, message.ErrInvalidAuthMessage)
		})
	}
}

func TestAuthErrorMessage_Decode_InvalidUTF8(t *testing.T) {
	// Body: code 0x2, then a one-byte reason that is not UTF-8.
	src := bytes.NewReader([]byte{0x03, 0x02, 0x01, 0xff})

	var decoded message.AuthErrorMessage
	err := decoded.Decode(src)

	assert.ErrorIs(t, err, message.ErrInvalidAuthMessage)
}

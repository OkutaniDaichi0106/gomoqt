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

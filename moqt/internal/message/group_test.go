package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		input   message.GroupMessage
		wantErr bool
	}{
		"valid message": {
			input: message.GroupMessage{
				SubscribeID:   1,
				GroupSequence: 1,
			},
		},
		"max values": {
			input: message.GroupMessage{
				SubscribeID:   message.SubscribeID(1<<(64-2) - 1),   // maxVarInt8 (uint62 max)
				GroupSequence: message.GroupSequence(1<<(64-2) - 1), // maxVarInt8 (uint62 max)
			},
		},
		"too large number": {
			input: message.GroupMessage{
				SubscribeID:   message.SubscribeID(1<<64 - 1),   // uint64 max
				GroupSequence: message.GroupSequence(1<<64 - 1), // uint64 max
			},
			wantErr: true,
		},
		"zero values": {
			input: message.GroupMessage{
				SubscribeID:   0,
				GroupSequence: 0,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			// Encode
			func() {
				defer func() {
					if r := recover(); r != nil {
						if tc.wantErr {
							// Expected panic, treat as error
							return
						}
						panic(r) // Re-panic if not expected
					}
				}()
				err := tc.input.Encode(&buf)
				if tc.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}()

			if tc.wantErr {
				return // Skip decode for error cases
			}

			// Decode
			var decoded message.GroupMessage
			err := decoded.Decode(&buf)
			require.NoError(t, err)

			// Compare fields
			assert.Equal(t, tc.input, decoded, "decoded message should match input")
		})
	}
}

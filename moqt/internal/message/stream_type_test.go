package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestStreamTypeMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		streamType message.StreamType
		wantErr    bool
	}{
		"valid message": {
			streamType: 0,
			wantErr:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			streamTypeMessage := &message.StreamTypeMessage{
				StreamType: tc.streamType,
			}
			var buf bytes.Buffer

			err := streamTypeMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedStreamTypeMessage := &message.StreamTypeMessage{}
			err = decodedStreamTypeMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, streamTypeMessage.StreamType, decodedStreamTypeMessage.StreamType)
		})
	}
}

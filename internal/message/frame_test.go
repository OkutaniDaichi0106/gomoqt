package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestFrameMessage(t *testing.T) {
	tests := map[string]struct {
		payload []byte
		want    []byte
		wantErr bool
	}{
		"valid payload":  {payload: []byte{1, 2}, want: []byte{1, 2}, wantErr: false},
		"empty payload":  {payload: []byte{}, want: []byte{}, wantErr: false},
		"string payload": {payload: []byte{0x62, 0x61, 0x72}, want: []byte{0x62, 0x61, 0x72}, wantErr: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			frame := &message.FrameMessage{
				Payload: tc.payload,
			}
			var buf bytes.Buffer

			err := frame.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			err = frame.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, frame.Payload, tc.payload)

		})
	}
}

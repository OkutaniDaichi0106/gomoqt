package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

func TestSessionUpdateMessage(t *testing.T) {
	tests := map[string]struct {
		input   uint64
		want    uint64
		wantErr bool
	}{
		"valid bitrate": {input: 12345, want: 12345, wantErr: false},
		"zero bitrate":  {input: 0, want: 0, wantErr: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sum := message.SessionUpdateMessage{Bitrate: tc.input}

			var buf bytes.Buffer
			err := sum.Encode(&buf)

			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error: %v, got: %v", tc.wantErr, err)
			}

			var deserialized message.SessionUpdateMessage
			err = deserialized.Decode(quicvarint.NewReader(&buf))

			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error: %v, got: %v", tc.wantErr, err)
			}

			if deserialized.Bitrate != tc.want {
				t.Fatalf("expected bitrate: %v, got: %v", tc.want, deserialized.Bitrate)
			}
		})
	}
}

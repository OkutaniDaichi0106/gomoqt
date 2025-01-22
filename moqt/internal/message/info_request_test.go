package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestInfoRequestMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		trackPath []string
		wantErr   bool
	}{
		"valid message": {
			trackPath: []string{"path", "to", "track"},
			wantErr:   false,
		},
		"empty track path": {
			trackPath: []string{},
			wantErr:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			infoRequestMessage := &message.InfoRequestMessage{
				TrackPath: tc.trackPath,
			}
			var buf bytes.Buffer

			err := infoRequestMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedInfoRequestMessage := &message.InfoRequestMessage{}
			err = decodedInfoRequestMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, infoRequestMessage.TrackPath, decodedInfoRequestMessage.TrackPath)
		})
	}
}

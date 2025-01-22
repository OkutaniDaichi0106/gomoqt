package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestAnnouncePleaseEncodeDecode(t *testing.T) {
	tests := map[string]struct {
		trackPrefix []string
		parameters  message.Parameters
		wantErr     bool
	}{
		"valid": {
			trackPrefix: []string{"part1", "part2"},
			parameters:  message.Parameters{0: []byte("value")},
			wantErr:     false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			announce := &message.AnnouncePleaseMessage{
				TrackPathPrefix: tc.trackPrefix,
				Parameters:      tc.parameters,
			}
			var buf bytes.Buffer

			err := announce.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedAnnounce := &message.AnnouncePleaseMessage{}
			err = decodedAnnounce.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, announce.TrackPathPrefix, decodedAnnounce.TrackPathPrefix)
			assert.Equal(t, announce.Parameters, decodedAnnounce.Parameters)
		})
	}
}

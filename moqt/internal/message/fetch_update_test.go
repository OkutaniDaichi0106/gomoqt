package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestFetchUpdateMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		trackPriority message.TrackPriority
		wantErr       bool
	}{
		"valid message": {
			trackPriority: 5,
			wantErr:       false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fetchUpdateMessage := &message.FetchUpdateMessage{
				TrackPriority: tc.trackPriority,
			}
			var buf bytes.Buffer

			err := fetchUpdateMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedFetchUpdateMessage := &message.FetchUpdateMessage{}
			err = decodedFetchUpdateMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, fetchUpdateMessage.TrackPriority, decodedFetchUpdateMessage.TrackPriority)
		})
	}
}

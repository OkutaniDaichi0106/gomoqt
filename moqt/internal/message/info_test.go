package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestInfoMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		trackPriority       message.TrackPriority
		latestGroupSequence message.GroupSequence
		groupOrder          message.GroupOrder
		wantErr             bool
	}{
		"valid message": {
			trackPriority:       1,
			latestGroupSequence: 2,
			groupOrder:          3,
			wantErr:             false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			infoMessage := &message.InfoMessage{
				TrackPriority:       tc.trackPriority,
				LatestGroupSequence: tc.latestGroupSequence,
				GroupOrder:          tc.groupOrder,
			}
			var buf bytes.Buffer

			err := infoMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedInfoMessage := &message.InfoMessage{}
			err = decodedInfoMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, infoMessage.TrackPriority, decodedInfoMessage.TrackPriority)
			assert.Equal(t, infoMessage.LatestGroupSequence, decodedInfoMessage.LatestGroupSequence)
			assert.Equal(t, infoMessage.GroupOrder, decodedInfoMessage.GroupOrder)
		})
	}
}

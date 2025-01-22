package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestFetchMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		subscribeID   message.SubscribeID
		trackPath     []string
		trackPriority message.TrackPriority
		groupSequence message.GroupSequence
		frameSequence message.FrameSequence
		wantErr       bool
	}{
		"valid message": {
			subscribeID:   1,
			trackPath:     []string{"path", "to", "track"},
			trackPriority: 5,
			groupSequence: 10,
			frameSequence: 15,
			wantErr:       false,
		},
		"empty track path": {
			subscribeID:   1,
			trackPath:     []string{},
			trackPriority: 5,
			groupSequence: 10,
			frameSequence: 15,
			wantErr:       false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fetchMessage := &message.FetchMessage{
				SubscribeID:   tc.subscribeID,
				TrackPath:     tc.trackPath,
				TrackPriority: tc.trackPriority,
				GroupSequence: tc.groupSequence,
				FrameSequence: tc.frameSequence,
			}
			var buf bytes.Buffer

			err := fetchMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedFetchMessage := &message.FetchMessage{}
			err = decodedFetchMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, fetchMessage.SubscribeID, decodedFetchMessage.SubscribeID)
			assert.Equal(t, fetchMessage.TrackPath, decodedFetchMessage.TrackPath)
			assert.Equal(t, fetchMessage.TrackPriority, decodedFetchMessage.TrackPriority)
			assert.Equal(t, fetchMessage.GroupSequence, decodedFetchMessage.GroupSequence)
			assert.Equal(t, fetchMessage.FrameSequence, decodedFetchMessage.FrameSequence)
		})
	}
}

package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		subscribeID      message.SubscribeID
		trackPath        []string
		trackPriority    message.TrackPriority
		groupOrder       message.GroupOrder
		minGroupSequence message.GroupSequence
		maxGroupSequence message.GroupSequence
		parameter        message.Parameters
		wantErr          bool
	}{
		"valid message": {
			subscribeID:      1,
			trackPath:        []string{"path", "to", "track"},
			trackPriority:    5,
			groupOrder:       1,
			minGroupSequence: 10,
			maxGroupSequence: 20,
			parameter:        message.Parameters{1: []byte("value")},
			wantErr:          false,
		},
		"empty track path": {
			subscribeID:      1,
			trackPath:        []string{},
			trackPriority:    5,
			groupOrder:       1,
			minGroupSequence: 10,
			maxGroupSequence: 20,
			parameter:        message.Parameters{1: []byte("value")},
			wantErr:          false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			subscribe := &message.SubscribeMessage{
				SubscribeID:         tc.subscribeID,
				TrackPath:           tc.trackPath,
				TrackPriority:       tc.trackPriority,
				GroupOrder:          tc.groupOrder,
				MinGroupSequence:    tc.minGroupSequence,
				MaxGroupSequence:    tc.maxGroupSequence,
				SubscribeParameters: tc.parameter,
			}
			var buf bytes.Buffer

			err := subscribe.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedSubscribe := &message.SubscribeMessage{}
			err = decodedSubscribe.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, subscribe.SubscribeID, decodedSubscribe.SubscribeID)
			assert.Equal(t, subscribe.TrackPath, decodedSubscribe.TrackPath)
			assert.Equal(t, subscribe.TrackPriority, decodedSubscribe.TrackPriority)
			assert.Equal(t, subscribe.GroupOrder, decodedSubscribe.GroupOrder)
			assert.Equal(t, subscribe.MinGroupSequence, decodedSubscribe.MinGroupSequence)
			assert.Equal(t, subscribe.MaxGroupSequence, decodedSubscribe.MaxGroupSequence)
			assert.Equal(t, subscribe.SubscribeParameters, decodedSubscribe.SubscribeParameters)
		})
	}
}

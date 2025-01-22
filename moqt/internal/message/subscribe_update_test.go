package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeUpdateMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		trackPriority    message.TrackPriority
		minGroupSequence message.GroupSequence
		maxGroupSequence message.GroupSequence
		parameters       message.Parameters
		wantErr          bool
	}{
		"valid message": {
			trackPriority:    5,
			minGroupSequence: 10,
			maxGroupSequence: 20,
			parameters: map[uint64][]byte{
				1: []byte("value1"),
				2: []byte("value2"),
			},
			wantErr: false,
		},
		"empty parameters": {
			trackPriority:    5,
			minGroupSequence: 10,
			maxGroupSequence: 20,
			parameters:       map[uint64][]byte{},
			wantErr:          false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			subscribeUpdate := &message.SubscribeUpdateMessage{
				TrackPriority:             message.TrackPriority(tc.trackPriority),
				MinGroupSequence:          tc.minGroupSequence,
				MaxGroupSequence:          tc.maxGroupSequence,
				SubscribeUpdateParameters: tc.parameters,
			}
			var buf bytes.Buffer

			err := subscribeUpdate.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedSubscribeUpdate := &message.SubscribeUpdateMessage{}
			err = decodedSubscribeUpdate.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, subscribeUpdate.TrackPriority, decodedSubscribeUpdate.TrackPriority)
			assert.Equal(t, subscribeUpdate.MinGroupSequence, decodedSubscribeUpdate.MinGroupSequence)
			assert.Equal(t, subscribeUpdate.MaxGroupSequence, decodedSubscribeUpdate.MaxGroupSequence)
			assert.Equal(t, subscribeUpdate.GroupOrder, decodedSubscribeUpdate.GroupOrder)
			assert.Equal(t, subscribeUpdate.SubscribeUpdateParameters, decodedSubscribeUpdate.SubscribeUpdateParameters)
		})
	}
}

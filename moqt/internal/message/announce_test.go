package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestAnnounceMessage_EncodeDecode(t *testing.T) {
	tests := []struct {
		name            string
		announceStatus  uint64
		trackPathSuffix []string
		parameters      map[uint64][]byte
		wantErr         bool
	}{
		{
			name:            "valid message",
			announceStatus:  1,
			trackPathSuffix: []string{"path", "to", "track"},
			parameters: map[uint64][]byte{
				1: []byte("value1"),
				2: []byte("value2"),
			},
			wantErr: false,
		},
		{
			name:            "empty track path suffix",
			announceStatus:  1,
			trackPathSuffix: []string{},
			parameters: map[uint64][]byte{
				1: []byte("value1"),
				2: []byte("value2"),
			},
			wantErr: false,
		},
		{
			name:            "empty parameters",
			announceStatus:  1,
			trackPathSuffix: []string{"path", "to", "track"},
			parameters:      map[uint64][]byte{},
			wantErr:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			announce := &message.AnnounceMessage{
				AnnounceStatus:  message.AnnounceStatus(tc.announceStatus),
				TrackPathSuffix: tc.trackPathSuffix,
				Parameters:      tc.parameters,
			}
			var buf bytes.Buffer

			err := announce.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedAnnounce := &message.AnnounceMessage{}
			err = decodedAnnounce.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, announce.AnnounceStatus, decodedAnnounce.AnnounceStatus)
			assert.Equal(t, announce.TrackPathSuffix, decodedAnnounce.TrackPathSuffix)
			assert.Equal(t, announce.Parameters, decodedAnnounce.Parameters)
		})
	}
}

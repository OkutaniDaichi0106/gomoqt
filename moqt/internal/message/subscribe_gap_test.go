package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeGapMessageEncodeDecode(t *testing.T) {
	tests := map[string]struct {
		minGapSequence message.GroupSequence
		maxGapSequence message.GroupSequence
		groupErrorCode message.GroupErrorCode
		wantErr        bool
	}{
		"valid message": {
			minGapSequence: 100,
			maxGapSequence: 200,
			groupErrorCode: 1,
			wantErr:        false,
		},
		"zero sequences": {
			minGapSequence: 0,
			maxGapSequence: 0,
			groupErrorCode: 0,
			wantErr:        false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			subscribeGap := &message.SubscribeGapMessage{
				MinGapSequence: tc.minGapSequence,
				MaxGapSequence: tc.maxGapSequence,
				GroupErrorCode: tc.groupErrorCode,
			}
			var buf bytes.Buffer

			err := subscribeGap.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedSubscribeGap := &message.SubscribeGapMessage{}
			err = decodedSubscribeGap.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, subscribeGap.MinGapSequence, decodedSubscribeGap.MinGapSequence)
			assert.Equal(t, subscribeGap.MaxGapSequence, decodedSubscribeGap.MaxGapSequence)
			assert.Equal(t, subscribeGap.GroupErrorCode, decodedSubscribeGap.GroupErrorCode)
		})
	}
}

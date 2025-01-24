package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestGroupMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		testcase message.GroupMessage
		want     message.GroupMessage
		wantErr  bool
	}{
		"valid parameter": {
			testcase: message.GroupMessage{
				SubscribeID:   1,
				GroupSequence: 1,
			},
			want: message.GroupMessage{
				SubscribeID:   1,
				GroupSequence: 1,
			},
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			group := tc.testcase
			var buf bytes.Buffer
			err := group.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			err = group.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if err == nil && tc.wantErr {
				t.Fatalf("expected error")
			}

			assert.Equal(t, group, tc.want)

		})
	}
}

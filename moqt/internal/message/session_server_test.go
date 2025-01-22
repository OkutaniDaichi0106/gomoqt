package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
)

func TestSessionServerMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		version    protocol.Version
		parameters message.Parameters
		wantErr    bool
	}{
		"valid message": {
			version: 0,
			parameters: message.Parameters{
				1: []byte("value1"),
				2: []byte("value2"),
			},
			wantErr: false,
		},
		"empty parameters": {
			version:    0,
			parameters: message.Parameters{},
			wantErr:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sessionServerMessage := &message.SessionServerMessage{
				SelectedVersion: tc.version,
				Parameters:      tc.parameters,
			}
			var buf bytes.Buffer

			err := sessionServerMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedSessionServerMessage := &message.SessionServerMessage{}
			err = decodedSessionServerMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, sessionServerMessage.SelectedVersion, decodedSessionServerMessage.SelectedVersion)
			assert.Equal(t, sessionServerMessage.Parameters, decodedSessionServerMessage.Parameters)
		})
	}
}

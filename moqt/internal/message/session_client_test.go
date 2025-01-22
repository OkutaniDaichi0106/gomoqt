package message_test

import (
	"bytes"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/stretchr/testify/assert"
)

func TestSessionClientMessage_EncodeDecode(t *testing.T) {
	tests := map[string]struct {
		versions   []protocol.Version
		parameters message.Parameters
		wantErr    bool
	}{
		"valid message": {
			versions: []protocol.Version{0},
			parameters: message.Parameters{
				1: []byte("value1"),
				2: []byte("value2"),
			},
			wantErr: false,
		},
		"empty parameters": {
			versions:   []protocol.Version{0},
			parameters: message.Parameters{},
			wantErr:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sessionClientMessage := &message.SessionClientMessage{
				SupportedVersions: tc.versions,
				Parameters:        tc.parameters,
			}
			var buf bytes.Buffer

			err := sessionClientMessage.Encode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			decodedSessionClientMessage := &message.SessionClientMessage{}
			err = decodedSessionClientMessage.Decode(&buf)
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && tc.wantErr {
				t.Fatalf("expected error: %v", err)
			}

			assert.Equal(t, sessionClientMessage.SupportedVersions, decodedSessionClientMessage.SupportedVersions)
			assert.Equal(t, sessionClientMessage.Parameters, decodedSessionClientMessage.Parameters)
		})
	}
}

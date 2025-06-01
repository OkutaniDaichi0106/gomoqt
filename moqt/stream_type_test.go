package moqt

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/stretchr/testify/assert"
)

func TestStreamType_Constants(t *testing.T) {
	tests := map[string]struct {
		streamType message.StreamType
		expected   message.StreamType
	}{
		"session constant": {
			streamType: stream_type_session,
			expected:   message.StreamType(0x0),
		},
		"announce constant": {
			streamType: stream_type_announce,
			expected:   message.StreamType(0x1),
		},
		"subscribe constant": {
			streamType: stream_type_subscribe,
			expected:   message.StreamType(0x2),
		},
		"group constant": {
			streamType: stream_type_group,
			expected:   message.StreamType(0x0),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.streamType)
		})
	}
}

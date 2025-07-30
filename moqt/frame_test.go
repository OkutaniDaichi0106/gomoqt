package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFrame(t *testing.T) {
	tests := map[string]struct {
		data     []byte
		expected []byte
	}{
		"normal data": {
			data:     []byte("test frame data"),
			expected: []byte("test frame data"),
		},
		"empty data": {
			data:     []byte{},
			expected: []byte{},
		},
		"binary data": {
			data:     []byte{0x00, 0x01, 0x02, 0xFF},
			expected: []byte{0x00, 0x01, 0x02, 0xFF},
		}, "nil data": {
			data:     nil,
			expected: []byte{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			frame := NewFrame(tt.data)
			assert.NotNil(t, frame)

			copiedBytes := frame.Bytes()
			if tt.expected == nil {
				assert.Empty(t, copiedBytes)
			} else {
				assert.Equal(t, tt.expected, copiedBytes)

				// Verify it's a copy, not the same slice
				if len(copiedBytes) > 0 {
					copiedBytes[0] = 'X'
					originalCopy := frame.Bytes()
					assert.NotEqual(t, copiedBytes[0], originalCopy[0])
				}
			}
		})
	}
}

func TestFrame_CopyBytes(t *testing.T) {
	tests := map[string]struct {
		data []byte
	}{
		"normal data": {
			data: []byte("hello world"),
		},
		"empty data": {
			data: []byte{},
		},
		"binary data": {
			data: []byte{0x00, 0x01, 0x02, 0xFF},
		},
		"nil data": {
			data: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			frame := NewFrame(tt.data)
			copiedBytes := frame.Bytes()

			if tt.data == nil {
				assert.Empty(t, copiedBytes)
			} else {
				assert.Equal(t, tt.data, copiedBytes)
			}
		})
	}
}

func TestFrame_Size(t *testing.T) {
	tests := map[string]struct {
		data []byte
		want int
	}{
		"normal data": {
			data: []byte("hello"),
			want: 5,
		},
		"empty data": {
			data: []byte{},
			want: 0,
		},
		"large data": {
			data: make([]byte, 1024),
			want: 1024,
		},
		"nil data": {
			data: nil,
			want: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			frame := NewFrame(tt.data)
			size := frame.Len()
			assert.Equal(t, tt.want, size)
		})
	}
}

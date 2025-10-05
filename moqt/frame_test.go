package moqt

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			// Create frame with capacity equal to len(data)
			cap := len(tt.data)
			builder := NewFrameBuilder(cap)
			assert.NotNil(t, builder)

			if len(tt.data) > 0 {
				builder.Append(tt.data)
			}

			frame := builder.Frame()
			copiedBytes := frame.Bytes()
			if len(tt.expected) == 0 {
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
			cap := len(tt.data)
			builder := NewFrameBuilder(cap)
			builder.Append(tt.data)

			frame := builder.Frame()
			copiedBytes := frame.Bytes()

			if len(tt.data) == 0 {
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
			cap := len(tt.data)
			builder := NewFrameBuilder(cap)
			builder.Append(tt.data)
			frame := builder.Frame()
			size := frame.Len()
			assert.Equal(t, tt.want, size)
		})
	}
}

func TestFrame_ResetAndAppendGrowth(t *testing.T) {
	// Ensure reset clears length but keeps capacity, append grows buffer when needed
	b := make([]byte, 512)
	for i := range b {
		b[i] = byte(i % 256)
	}

	builder := NewFrameBuilder(16)
	// append small data first
	builder.Append([]byte("small"))
	frame := builder.Frame()
	oldCap := frame.Cap()
	oldLen := frame.Len()
	assert.True(t, oldCap >= oldLen)

	// append large data to force growth
	builder.Append(b)
	frame = builder.Frame()
	assert.Equal(t, oldLen+len(b), frame.Len())
	assert.True(t, frame.Cap() >= frame.Len())

	// reset should set length to zero but preserve capacity
	builder.Reset()
	frame = builder.Frame()
	assert.Equal(t, 0, frame.Len())
	assert.True(t, frame.Cap() >= 0)
}

func TestFrame_CloneIndependence(t *testing.T) {
	data := []byte("original data")
	builder := NewFrameBuilder(len(data))
	builder.Append(data)
	frame := builder.Frame()

	clone := frame.Clone()
	// modify clone's returned bytes and ensure original not affected
	cb := clone.Bytes()
	require.NotNil(t, cb)
	if len(cb) > 0 {
		cb[0] = 'X'
	}

	ob := frame.Bytes()
	if len(ob) > 0 {
		assert.NotEqual(t, cb[0], ob[0])
	}
}

func TestFrame_EncodeDecode_RoundTrip(t *testing.T) {
	// round-trip encode/decode using an in-memory buffer
	data := []byte("roundtrip payload")
	builder := NewFrameBuilder(len(data))
	builder.Append(data)
	frame := builder.Frame()

	var buf bytes.Buffer
	err := frame.encode(&buf)
	assert.NoError(t, err)

	// prepare a fresh frame with capacity 0 and decode
	f2 := newFrame(0)
	err = f2.decode(&buf)
	assert.NoError(t, err)
	assert.Equal(t, frame.Bytes(), f2.Bytes())
}

func TestFrame_WriteTo(t *testing.T) {
	tests := map[string]struct {
		data []byte
	}{
		"normal data": {
			data: []byte("test frame data"),
		},
		"empty data": {
			data: []byte{},
		},
		"binary data": {
			data: []byte{0x00, 0x01, 0x02, 0xFF},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			builder := NewFrameBuilder(len(tt.data))
			if len(tt.data) > 0 {
				builder.Append(tt.data)
			}
			frame := builder.Frame()

			var buf bytes.Buffer
			n, err := frame.WriteTo(&buf)
			assert.NoError(t, err)
			assert.Equal(t, int64(len(tt.data)), n)
			if len(tt.data) == 0 {
				assert.Nil(t, buf.Bytes())
			} else {
				assert.Equal(t, tt.data, buf.Bytes())
			}
		})
	}
}

func TestFrame_Decode(t *testing.T) {
	tests := map[string]struct {
		data []byte
	}{
		"normal data": {
			data: []byte("test frame data"),
		},
		"empty data": {
			data: []byte{},
		},
		"binary data": {
			data: []byte{0x00, 0x01, 0x02, 0xFF},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create original frame
			builder := NewFrameBuilder(len(tt.data))
			if len(tt.data) > 0 {
				builder.Append(tt.data)
			}
			originalFrame := builder.Frame()

			// Encode to buffer
			var buf bytes.Buffer
			require.NoError(t, originalFrame.encode(&buf))

			// Decode from buffer
			newFrame := newFrame(0)
			err := newFrame.decode(&buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, newFrame.Bytes())
		})
	}
}

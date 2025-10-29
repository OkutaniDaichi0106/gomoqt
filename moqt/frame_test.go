package moqt

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFrame(t *testing.T) {
	tests := map[string]struct {
		capacity int
		data     []byte
		expected []byte
	}{
		"normal data": {
			capacity: 100,
			data:     []byte("test frame data"),
			expected: []byte("test frame data"),
		},
		"empty data": {
			capacity: 10,
			data:     []byte{},
			expected: []byte{},
		},
		"binary data": {
			capacity: 50,
			data:     []byte{0x00, 0x01, 0x02, 0xFF},
			expected: []byte{0x00, 0x01, 0x02, 0xFF},
		},
		"zero capacity": {
			capacity: 0,
			data:     nil,
			expected: []byte{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			frame := NewFrame(tt.capacity)
			assert.NotNil(t, frame)

			if len(tt.data) > 0 {
				frame.Write(tt.data)
			}

			assert.Equal(t, tt.expected, frame.Body())
		})
	}
}

func TestFrame_WriteGrowth(t *testing.T) {
	// Test that Write correctly grows the body buffer when needed
	frame := NewFrame(5)
	assert.Equal(t, 5, frame.Cap())
	assert.Equal(t, 0, frame.Len())

	// Write data within capacity
	frame.Write([]byte("abc"))
	assert.Equal(t, 3, frame.Len())
	assert.Equal(t, []byte("abc"), frame.Body())

	// Write more data to trigger growth
	frame.Write([]byte("defghij"))
	assert.Equal(t, 10, frame.Len())
	assert.Equal(t, []byte("abcdefghij"), frame.Body())
	assert.True(t, frame.Cap() >= 10)
}

func TestFrame_Reset(t *testing.T) {
	// Test that Reset clears the body while preserving capacity
	frame := NewFrame(20)
	frame.Write([]byte("some data"))
	assert.Equal(t, 9, frame.Len())
	originalCap := frame.Cap()

	frame.Reset()
	assert.Equal(t, 0, frame.Len())
	assert.Equal(t, originalCap, frame.Cap())
	assert.Len(t, frame.Body(), 0)

	// Verify we can write again after reset
	frame.Write([]byte("new data"))
	assert.Equal(t, 8, frame.Len())
	assert.Equal(t, []byte("new data"), frame.Body())
}

func TestFrame_Clone(t *testing.T) {
	// Test that Clone creates an independent copy
	originalData := []byte("original data")
	frame := NewFrame(len(originalData))
	frame.Write(originalData)

	clone := frame.Clone()
	assert.NotNil(t, clone)
	assert.Equal(t, frame.Body(), clone.Body())

	// Modify original and verify clone is unchanged
	frame.Reset()
	frame.Write([]byte("modified"))
	assert.Equal(t, []byte("modified"), frame.Body())
	assert.Equal(t, originalData, clone.Body())

	// Verify bodies point to different underlying arrays
	assert.NotSame(t, &frame.Body()[0], &clone.Body()[0])
}

func TestFrame_EncodeDecode_RoundTrip(t *testing.T) {
	// Test encode/decode round trip with various payloads
	tests := []struct {
		name    string
		payload []byte
	}{
		{"empty", []byte{}},
		{"small", []byte("test")},
		{"medium", []byte("this is a medium-sized payload for testing")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}},
		{"large", make([]byte, 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and encode frame
			frame := NewFrame(len(tt.payload) + 8)
			frame.Write(tt.payload)

			var buf bytes.Buffer
			err := frame.encode(&buf)
			assert.NoError(t, err)

			// Decode into new frame
			decodedFrame := NewFrame(0)
			err = decodedFrame.decode(&buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.payload, decodedFrame.Body())
		})
	}
}

func TestFrame_WriteTo(t *testing.T) {
	// Test WriteTo writes the payload to a writer
	tests := []struct {
		name    string
		payload []byte
	}{
		{"empty", []byte{}},
		{"text", []byte("hello world")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewFrame(len(tt.payload) + 8)
			frame.Write(tt.payload)

			var buf bytes.Buffer
			n, err := frame.WriteTo(&buf)
			assert.NoError(t, err)
			assert.Equal(t, int64(len(tt.payload)), n)
			// For empty payload, buf.Bytes() returns nil, not empty slice
			if len(tt.payload) == 0 {
				assert.Len(t, buf.Bytes(), 0)
			} else {
				assert.Equal(t, tt.payload, buf.Bytes())
			}
		})
	}
}

func TestFrame_LenAndCap(t *testing.T) {
	// Test Len and Cap methods reflect internal state correctly
	frame := NewFrame(50)
	assert.Equal(t, 0, frame.Len())
	assert.Equal(t, 50, frame.Cap())

	frame.Write([]byte("test"))
	assert.Equal(t, 4, frame.Len())
	assert.Equal(t, 50, frame.Cap())

	// After growth
	largeData := make([]byte, 100)
	frame.Write(largeData)
	assert.Equal(t, 104, frame.Len())
	assert.True(t, frame.Cap() >= 104)
}

func TestFrame_EncodeHeaderLayout(t *testing.T) {
	// Test that encode correctly uses the header buffer for length encoding
	// This verifies the MOQ encoding optimization where header stores length varint
	frame := NewFrame(10)
	frame.Write([]byte("test"))

	var buf bytes.Buffer
	err := frame.encode(&buf)
	assert.NoError(t, err)

	encoded := buf.Bytes()
	assert.NotEmpty(t, encoded)
	// First byte(s) should be length header, followed by payload
	assert.True(t, len(encoded) > len([]byte("test")))
}

func TestFrame_Write(t *testing.T) {
	// Test Write implements io.Writer interface
	tests := []struct {
		name  string
		data  [][]byte
		total int
	}{
		{"single write", [][]byte{[]byte("hello")}, 5},
		{"multiple writes", [][]byte{[]byte("hello"), []byte(" "), []byte("world")}, 11},
		{"empty write", [][]byte{[]byte("")}, 0},
		{"binary data", [][]byte{[]byte{0x00, 0x01, 0x02}}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewFrame(100)
			totalWritten := 0

			for _, data := range tt.data {
				n, err := frame.Write(data)
				assert.NoError(t, err)
				assert.Equal(t, len(data), n)
				totalWritten += n
			}

			assert.Equal(t, tt.total, frame.Len())
			assert.Equal(t, tt.total, totalWritten)
		})
	}
}

func TestFrame_Write_AsIOWriter(t *testing.T) {
	// Test that Frame can be used as io.Writer with io.Copy
	frame := NewFrame(100)

	// Write to frame using io.Copy
	source := bytes.NewReader([]byte("test data from reader"))
	n, err := io.Copy(frame, source)
	assert.NoError(t, err)
	assert.Equal(t, int64(21), n)
	assert.Equal(t, []byte("test data from reader"), frame.Body())
}

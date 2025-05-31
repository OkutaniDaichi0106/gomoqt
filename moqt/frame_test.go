package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestNewFrame(t *testing.T) {
	testData := []byte("test frame data")

	frame := moqt.NewFrame(testData)
	assert.NotNil(t, frame)

	// Test CopyBytes returns a copy of the data
	copiedBytes := frame.CopyBytes()
	assert.Equal(t, testData, copiedBytes)

	// Verify it's a copy, not the same slice
	if len(copiedBytes) > 0 {
		copiedBytes[0] = 'X'
		originalCopy := frame.CopyBytes()
		assert.NotEqual(t, copiedBytes[0], originalCopy[0])
	}
}

func TestFrame_CopyBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "normal data",
			data: []byte("hello world"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF},
		},
		{
			name: "nil data",
			data: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := moqt.NewFrame(tt.data)
			copiedBytes := frame.CopyBytes()

			if tt.data == nil {
				assert.Nil(t, copiedBytes)
			} else {
				assert.Equal(t, tt.data, copiedBytes)
			}
		})
	}
}

func TestFrame_Size(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			name: "normal data",
			data: []byte("hello"),
			want: 5,
		},
		{
			name: "empty data",
			data: []byte{},
			want: 0,
		},
		{
			name: "large data",
			data: make([]byte, 1024),
			want: 1024,
		},
		{
			name: "nil data",
			data: nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := moqt.NewFrame(tt.data)
			size := frame.Size()
			assert.Equal(t, tt.want, size)
		})
	}
}

func TestFrame_Release(t *testing.T) {
	// Test that Release doesn't panic
	frame := moqt.NewFrame([]byte("test data"))
	assert.NotPanics(t, func() {
		frame.Release()
	})

	// Test multiple releases don't panic
	assert.NotPanics(t, func() {
		frame.Release()
	})
}

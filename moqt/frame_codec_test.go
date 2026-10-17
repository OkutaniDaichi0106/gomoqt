package moqt

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestFrame returns a frame at timestamp ts carrying body.
func newTestFrame(tb testing.TB, ts uint64, body string) *Frame {
	tb.Helper()
	f := NewFrame(len(body))
	f.Timestamp = ts
	_, err := f.Write([]byte(body))
	require.NoError(tb, err)
	return f
}

// The bytes are draft-lcurley-moq-lite-05's FRAME messages, worked out by
// hand: a zigzag delta from the previous timestamp (the first from 0), a
// QUIC varint length, the payload.
func TestFrameWriter_WriteFrame_Wire(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFrameWriter(&buf)

	require.NoError(t, fw.WriteFrame(newTestFrame(t, 0, "a")))
	require.NoError(t, fw.WriteFrame(newTestFrame(t, 3000, "bc")))
	require.NoError(t, fw.WriteFrame(newTestFrame(t, 1500, "")))

	assert.Equal(t, []byte{
		0x00, 0x01, 'a', // delta 0
		0x57, 0x70, 0x02, 'b', 'c', // zigzag(3000) = 6000 = 0x1770
		0x4b, 0xb7, 0x00, // zigzag(-1500) = 2999 = 0x0bb7
	}, buf.Bytes())
}

// A group the writer wrote reads back frame for frame, with reordered
// timestamps as B-frames carry them, through a reader that is an
// io.ByteReader and through one that is not.
func TestFrameReader_RoundTrip(t *testing.T) {
	timestamps := []uint64{90000, 96000, 93000, 1 << 50}
	var buf bytes.Buffer
	fw := NewFrameWriter(&buf)
	for i, ts := range timestamps {
		require.NoError(t, fw.WriteFrame(newTestFrame(t, ts, string(rune('a'+i)))))
	}
	tests := map[string]io.Reader{
		"a byte reader":     bytes.NewReader(buf.Bytes()),
		"a plain io.Reader": struct{ io.Reader }{bytes.NewReader(buf.Bytes())},
	}
	for name, r := range tests {
		t.Run(name, func(t *testing.T) {
			fr := NewFrameReader(r)
			got := NewFrame(0)

			for i, ts := range timestamps {
				require.NoError(t, fr.ReadFrame(got))
				assert.Equal(t, ts, got.Timestamp, "frame %d", i)
				assert.Equal(t, []byte{byte('a' + i)}, got.Body(), "frame %d", i)
			}
			assert.Equal(t, io.EOF, fr.ReadFrame(got), "a clean end after the last frame")
		})
	}
}

// Stored bytes cut short anywhere inside a frame are a truncated frame, never
// the clean end of the group.
func TestFrameReader_ReadFrame_Truncated(t *testing.T) {
	tests := map[string][]byte{
		"inside the timestamp delta": {0x40},
		"before the length":          {0x00},
		"inside the length":          {0x00, 0x40},
		"before the payload":         {0x00, 0x02},
		"inside the payload":         {0x00, 0x02, 'a'},
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			for _, r := range []io.Reader{bytes.NewReader(payload), struct{ io.Reader }{bytes.NewReader(payload)}} {
				err := NewFrameReader(r).ReadFrame(NewFrame(0))

				assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
			}
		})
	}
}

func TestFrameReader_ReadFrame_NegativeTimestamp(t *testing.T) {
	// zigzag 1 is -1: a first frame before zero.
	err := NewFrameReader(bytes.NewReader([]byte{0x01, 0x00})).ReadFrame(NewFrame(0))

	assert.ErrorIs(t, err, ErrTimestampOutOfRange)
}

func TestFrameWriter_WriteFrame_Nil(t *testing.T) {
	var buf bytes.Buffer

	err := NewFrameWriter(&buf).WriteFrame(nil)

	assert.NoError(t, err)
	assert.Zero(t, buf.Len())
}

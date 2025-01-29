package benchmarks_test

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

// Frame
func (f *GroupPointer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(quicvarint.Append(nil, uint64(len(*f.payload))))
	if err != nil {
		return int64(n), err
	}

	n2, err := w.Write(*f.payload)
	if err != nil {
		return int64(n + n2), err
	}
	return int64(n + n2), nil
}

// Frame with payload length as a field
type SerializedFrameWithLength struct {
	payload *[]byte
	length  int
}

func NewFrameWithLength(payload []byte) *SerializedFrameWithLength {
	buf := ptrBytesPool.Get().(*[]byte)
	*buf = (*buf)[:0]
	*buf = quicvarint.Append(*buf, uint64(len(payload)))
	copy(*buf, payload)
	return &SerializedFrameWithLength{
		payload: buf,
		length:  len(*buf),
	}
}

func (f *SerializedFrameWithLength) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(quicvarint.Append(nil, uint64(f.length)))
	if err != nil {
		return int64(n), err
	}
	n2, err := w.Write(*f.payload)
	if err != nil {
		return int64(n + n2), err
	}
	return int64(n + n2), nil
}

// // Frame with serialized payload and offset
// type FrameWithOffset struct {
// 	serializedPayload *[]byte
// 	offset            int
// }

// func NewFrameWithOffset(payload []byte) *FrameWithOffset {
// 	buf := ptrBytesPool.Get().(*[]byte)
// 	*buf = (*buf)[:0]
// 	*buf = quicvarint.Append(*buf, uint64(len(payload)))
// 	copy((*buf)[quicvarint.Len(uint64(len(payload))):], payload)
// 	return &FrameWithOffset{
// 		serializedPayload: buf,
// 		offset:            quicvarint.Len(uint64(len(payload))),
// 	}
// }

// func (f *FrameWithOffset) WriteTo(w io.Writer) (int64, error) {
// 	n, err := w.Write(*f.serializedPayload)
// 	return int64(n), err
// }

// // Frame with serialized payload and length
// type FrameSerializedWithLength struct {
// 	serializedPayload *[]byte
// 	length            int
// }

// func NewFrameWithSerializedLength(payload []byte) *FrameSerializedWithLength {
// 	buf := ptrBytesPool.Get().(*[]byte)
// 	*buf = (*buf)[:0]
// 	*buf = quicvarint.Append(*buf, uint64(len(payload)))
// 	copy((*buf)[quicvarint.Len(uint64(len(payload))):], payload)
// 	return &FrameSerializedWithLength{
// 		serializedPayload: buf,
// 		length:            len(payload),
// 	}
// }

// func (f *FrameSerializedWithLength) WriteTo(w io.Writer) (int64, error) {
// 	n, err := w.Write(*f.serializedPayload)
// 	return int64(n), err
// }

// // Frame with serialized payload and both length and payload
// type FrameWithSerializedPayload struct {
// 	serializedPayloadLen *[]byte
// 	payload              *[]byte
// }

// func NewFrameWithSerializedPayload(payload []byte) *FrameWithSerializedPayload {
// 	buf := ptrBytesPool.Get().(*[]byte)
// 	*buf = (*buf)[:0]
// 	*buf = quicvarint.Append(*buf, uint64(len(payload)))
// 	bufp := ptrBytesPool.Get().(*[]byte)
// 	*bufp = (*bufp)[:len(payload)] // 修正: bufp のサイズを payload の長さに設定
// 	copy(*bufp, payload)
// 	return &FrameWithSerializedPayload{
// 		serializedPayloadLen: buf,
// 		payload:              bufp,
// 	}
// }

// func (f *FrameWithSerializedPayload) WriteTo(w io.Writer) (int64, error) {
// 	n, err := w.Write(*f.serializedPayloadLen)
// 	if err != nil {
// 		return int64(n), err
// 	}
// 	n2, err := w.Write(*f.payload)
// 	return int64(n + n2), err
// }

// // Benchmark tests
// func BenchmarkFrameWriting(b *testing.B) {
// 	payload := make([]byte, 1<<10)
// 	buf := new(bytes.Buffer)

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		frame := NewFramePointerCopy(payload)
// 		frame.WriteTo(buf)
// 	}
// }

// func BenchmarkFrameWriting_WithLength(b *testing.B) {
// 	payload := make([]byte, 1024)
// 	buf := new(bytes.Buffer)

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		frame := NewFrameWithLength(payload)
// 		frame.WriteTo(buf)
// 	}
// }

// func BenchmarkFrameWriting_WithOffset(b *testing.B) {
// 	payload := make([]byte, 1<<10)
// 	buf := new(bytes.Buffer)

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		frame := NewFrameWithOffset(payload)
// 		frame.WriteTo(buf)
// 	}
// }

// func BenchmarkFrameWriting_SerializedWithLength(b *testing.B) {
// 	payload := make([]byte, 1<<10)
// 	buf := new(bytes.Buffer)

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		frame := NewFrameWithSerializedLength(payload)
// 		frame.WriteTo(buf)
// 	}
// }

// func BenchmarkFrameWriting_WithSerializedLengthAndPayload_Twice(b *testing.B) {
// 	payload := make([]byte, 1<<10)
// 	buf := new(bytes.Buffer)

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.Reset()
// 		frame := NewFrameWithSerializedPayload(payload)
// 		frame.WriteTo(buf)
// 	}
// }

// func TestFrameWriteToConsistency(t *testing.T) {
// 	payload := make([]byte, 1<<3)

// 	frames := []interface {
// 		WriteTo(io.Writer) (int64, error)
// 	}{
// 		NewFramePointerCopy(payload),
// 		NewFrameWithLength(payload),
// 		NewFrameWithOffset(payload),
// 		NewFrameWithSerializedLength(payload),
// 		NewFrameWithSerializedPayload(payload),
// 	}

// 	var expectedOutput []byte
// 	for i, frame := range frames {
// 		buf := new(bytes.Buffer)
// 		_, err := frame.WriteTo(buf)
// 		assert.NoError(t, err)

// 		if i == 0 {
// 			expectedOutput = buf.Bytes()
// 		} else {
// 			assert.Equal(t, expectedOutput, buf.Bytes(), "Frame %d did not match expected output", i)
// 		}
// 	}
// }

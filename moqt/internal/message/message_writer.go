package message

import (
	"fmt"
	"io"
)

func WriteVarint(w io.Writer, i uint64) error {
	if i <= maxVarInt1 {
		_, err := w.Write([]byte{byte(i)})
		return err
	}
	if i <= maxVarInt2 {
		b := []byte{
			uint8(i>>8) | 0x40,
			byte(i),
		}
		_, err := w.Write(b)
		return err
	}
	if i <= maxVarInt4 {
		b := []byte{
			uint8(i>>24) | 0x80,
			uint8(i >> 16),
			uint8(i >> 8),
			byte(i),
		}
		_, err := w.Write(b)
		return err
	}
	if i <= maxVarInt8 {
		b := []byte{
			uint8(i>>56) | 0xc0,
			uint8(i >> 48),
			uint8(i >> 40),
			uint8(i >> 32),
			uint8(i >> 24),
			uint8(i >> 16),
			uint8(i >> 8),
			byte(i),
		}
		_, err := w.Write(b)
		return err
	}
	panic(fmt.Sprintf("%#x doesn't fit into 62 bits", i))
}

func WriteBytes(w io.Writer, b []byte) error {
	if err := WriteVarint(w, uint64(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func WriteString(w io.Writer, s string) error {
	return WriteBytes(w, []byte(s))
}

func WriteStringArray(w io.Writer, arr []string) error {
	if err := WriteVarint(w, uint64(len(arr))); err != nil {
		return err
	}
	for _, str := range arr {
		if err := WriteString(w, str); err != nil {
			return err
		}
	}
	return nil
}

func WriteParameters(w io.Writer, params Parameters) error {
	if err := WriteVarint(w, uint64(len(params))); err != nil {
		return err
	}
	for key, value := range params {
		if err := WriteVarint(w, key); err != nil {
			return err
		}
		if err := WriteBytes(w, value); err != nil {
			return err
		}
	}
	return nil
}

const (
	maxVarInt1 = 1<<(8-2) - 1
	maxVarInt2 = 1<<(16-2) - 1
	maxVarInt4 = 1<<(32-2) - 1
	maxVarInt8 = 1<<(64-2) - 1
)

// import (
// 	"io"

// 	"github.com/quic-go/quic-go/quicvarint"
// )

// func Encode(w Writer, msg Message) error {
// 	w.WriteVarint(uint64(msg.Len()))
// 	msg.EncodePayload(w)
// 	return w.Flush()
// }

// type Writer interface {
// 	Release()
// 	Flush() error
// 	WriteVarint(num uint64)
// 	WriteString(str string)
// 	WriteBytes(bytes []byte)
// 	WriteStringArray(arr []string)
// 	WriteParameters(params Parameters)
// }

// func NewWriter(w io.Writer) *writer {
// 	return &writer{
// 		buf: getBytes(),
// 		w:   w,
// 	}
// }Append()

// type writer struct {
// 	buf []byte
// 	w   io.Writer
// }

// func (w *writer) Release() {
// 	putBytes(w.buf)
// 	w.buf = nil
// 	w.w = nil
// }

// func (w *writer) Flush() error {
// 	_, err := w.w.Write(w.buf)
// 	w.buf = w.buf[:0]
// 	return err
// }

// // Append a number w.buf the byte slice
// func (w *writer) WriteVarint(num uint64) {
// 	quicvarint.Append(w.buf, num)
// }

// // Append a string w.buf the byte slice
// func (w *writer) WriteString(str string) {
// 	w.WriteBytes([]byte(str))
// }

// // Append a byte slice w.buf the byte slice
// func (w *writer) WriteBytes(bytes []byte) {
// 	w.buf = quicvarint.Append(w.buf, uint64(len(bytes)))
// 	w.buf = append(w.buf, bytes...)
// }

// // Append a string array w.buf the byte slice
// func (w *writer) WriteStringArray(arr []string) {
// 	w.WriteVarint(uint64(len(arr)))
// 	for _, str := range arr {
// 		w.WriteString(str)
// 	}
// }

// // Append parameters w.buf the byte slice
// func (w *writer) WriteParameters(params Parameters) {
// 	w.WriteVarint(uint64(len(params)))
// 	for key, value := range params {
// 		w.WriteVarint(key)
// 		w.WriteBytes(value)
// 	}
// }

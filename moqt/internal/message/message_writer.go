package message

import (
	"fmt"
	"io"
)

func writeVarint(w io.Writer, i uint64) error {
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
	if err := writeVarint(w, uint64(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func WriteString(w io.Writer, s string) error {
	return WriteBytes(w, []byte(s))
}

func WriteStringArray(w io.Writer, arr []string) error {
	if err := writeVarint(w, uint64(len(arr))); err != nil {
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
	if err := writeVarint(w, uint64(len(params))); err != nil {
		return err
	}
	for key, value := range params {
		if err := writeVarint(w, key); err != nil {
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

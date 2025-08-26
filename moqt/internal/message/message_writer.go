package message

import (
	"fmt"
)

func WriteVarint(b []byte, i uint64) int {
	if i <= maxVarInt1 {
		b = append(b, byte(i))
		return 1
	}
	if i <= maxVarInt2 {
		b = append(b,
			uint8(i>>8)|0x40,
			byte(i),
		)
		return 2
	}
	if i <= maxVarInt4 {
		b = append(b,
			uint8(i>>24)|0x80,
			uint8(i>>16),
			uint8(i>>8),
			byte(i),
		)
		return 4
	}
	if i <= maxVarInt8 {
		b = append(b,
			uint8(i>>56)|0xc0,
			uint8(i>>48),
			uint8(i>>40),
			uint8(i>>32),
			uint8(i>>24),
			uint8(i>>16),
			uint8(i>>8),
			byte(i),
		)
		return 8
	}
	panic(fmt.Sprintf("%#x doesn't fit into 62 bits", i))
}

func WriteBytes(dest []byte, b []byte) int {
	n := WriteVarint(dest, uint64(len(b)))
	dest = append(dest, b...)
	return n + len(b)
}

func WriteString(dest []byte, s string) int {
	return WriteBytes(dest, []byte(s))
}

func WriteStringArray(dest []byte, arr []string) int {
	n := WriteVarint(dest, uint64(len(arr)))
	for _, str := range arr {
		n += WriteString(dest[n:], str)
	}
	return n
}

func WriteParameters(dest []byte, params Parameters) int {
	n := WriteVarint(dest, uint64(len(params)))
	for key, value := range params {
		n += WriteVarint(dest[n:], key)
		n += WriteBytes(dest[n:], value)
	}
	return n
}

const (
	maxVarInt1 = 1<<(8-2) - 1
	maxVarInt2 = 1<<(16-2) - 1
	maxVarInt4 = 1<<(32-2) - 1
	maxVarInt8 = 1<<(64-2) - 1
)

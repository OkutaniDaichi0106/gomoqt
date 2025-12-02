package message

import (
	"encoding/binary"
	"io"
	"math"
)

func ReadVarint(b []byte) (uint64, int, error) {
	if len(b) < 1 {
		return 0, 0, io.EOF
	}
	l := 1 << ((b[0] & 0xc0) >> 6)
	if len(b) < l {
		return 0, 0, io.EOF
	}
	var i uint64
	switch l {
	case 1:
		i = uint64(b[0] & (0xff - 0xc0))
	case 2:
		i = uint64(b[0]&0x3f)<<8 | uint64(b[1])
	case 4:
		i = uint64(b[0]&0x3f)<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
	case 8:
		i = uint64(b[0]&0x3f)<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
			uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	}
	return i, l, nil
}

func ReadMessageLength(r io.Reader) (uint16, error) {
	buf := [2]byte{}
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	length := binary.BigEndian.Uint16(buf[:])
	return length, nil
}

func ReadBytes(b []byte) ([]byte, int, error) {
	num, n, err := ReadVarint(b)
	if err != nil {
		return nil, 0, err
	}
	b = b[n:]
	if num > math.MaxInt {
		panic("byte slice too large")
	}

	if uint64(len(b)) < num {
		return b, n + len(b), io.EOF
	}

	return b[:num], n + int(num), nil
}

func ReadString(b []byte) (string, int, error) {
	str, n, err := ReadBytes(b)
	if err != nil {
		return "", 0, err
	}
	return string(str), n, nil
}

func ReadStringArray(b []byte) ([]string, int, error) {
	count, total, err := ReadVarint(b)
	if err != nil {
		return nil, 0, err
	}

	if count > math.MaxInt {
		panic("string array too large")
	}

	b = b[total:]

	arr := make([]string, 0, count)
	for range count {
		str, n, err := ReadString(b)
		if err != nil {
			return nil, 0, err
		}
		arr = append(arr, str)
		b = b[n:]
		total += n
	}

	return arr, total, nil
}

// Read parameters from the reader
func ReadParameters(b []byte) (map[uint64][]byte, int, error) {
	count, total, err := ReadVarint(b)
	if err != nil {
		return nil, 0, err
	}

	if count > math.MaxInt {
		panic("parameters too large")
	}

	b = b[total:]

	params := make(map[uint64][]byte, count)
	for range count {
		key, n, err := ReadVarint(b)
		if err != nil {
			return nil, 0, err
		}
		b = b[n:]
		total += n

		value, n, err := ReadBytes(b)
		if err != nil {
			return nil, 0, err
		}
		b = b[n:]
		total += n

		params[key] = value
	}

	return params, total, nil
}

package message

import (
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
		i = uint64(b[1]) + uint64(b[0]&0x3f)<<8
	case 4:
		i = uint64(b[3]) + uint64(b[2])<<8 + uint64(b[1])<<16 + uint64(b[0]&0x3f)<<24
	case 8:
		i = uint64(b[7]) + uint64(b[6])<<8 + uint64(b[5])<<16 + uint64(b[4])<<24 +
			uint64(b[3])<<32 + uint64(b[2])<<40 + uint64(b[1])<<48 + uint64(b[0]&0x3f)<<56
	}
	return i, l, nil
}

func ReadMessageLength(r io.Reader) (uint64, error) {
	buf := [1]byte{}
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	l := 1 << ((buf[0] & 0xc0) >> 6)
	b1 := buf[0] & (0xff - 0xc0)
	if l == 1 {
		return uint64(b1), nil
	}

	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b2 := buf[0]

	if l == 2 {
		return uint64(b2) + uint64(b1)<<8, nil
	}

	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b3 := buf[0]
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b4 := buf[0]

	if l == 4 {
		return uint64(b4) + uint64(b3)<<8 + uint64(b2)<<16 + uint64(b1)<<24, nil
	}

	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b5 := buf[0]
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b6 := buf[0]
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b7 := buf[0]
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	b8 := buf[0]

	return uint64(b8) + uint64(b7)<<8 + uint64(b6)<<16 + uint64(b5)<<24 + uint64(b4)<<32 + uint64(b3)<<40 + uint64(b2)<<48 + uint64(b1)<<56, nil
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

	var arr []string
	for range count {
		str, n, err := ReadString(b)
		if err != nil {
			if err == io.EOF {
				return arr, total, nil
			}
			return nil, 0, err
		}
		arr = append(arr, str)
		b = b[n:]
		total += n
	}

	return arr, total, nil
}

// Read parameters from the reader
func ReadParameters(b []byte) (Parameters, int, error) {
	count, total, err := ReadVarint(b)
	if err != nil {
		return nil, 0, err
	}

	if count > math.MaxInt {
		panic("parameters too large")
	}

	b = b[total:]

	params := make(Parameters, count)
	for range count {
		key, n, err := ReadVarint(b)
		if err != nil {
			return nil, 0, err
		}
		b = b[n:]
		total += n

		value, n, err := ReadBytes(b)
		if err != nil {
			if err == io.EOF {
				return params, total, nil
			}
			return nil, 0, err
		}
		b = b[n:]
		total += n

		params[key] = value
	}

	return params, total, nil
}

package message

import (
	"io"
)

func ReadVarint(r io.Reader) (uint64, error) {
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

func ReadBytes(r io.Reader) ([]byte, error) {
	num, err := ReadVarint(r)
	if err != nil {
		return nil, err
	}

	if num == 0 {
		return []byte{}, nil
	}

	b := make([]byte, num)
	_, err = r.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func ReadString(r io.Reader) (string, error) {
	b, err := ReadBytes(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ReadStringArray(r io.Reader) ([]string, error) {
	count, err := ReadVarint(r)
	if err != nil {
		return nil, err
	}

	arr := make([]string, count)
	for i := range count {
		str, err := ReadString(r)
		if err != nil {
			return nil, err
		}
		arr[i] = str
	}

	return arr, nil
}

// Read parameters from the reader
func ReadParameters(r io.Reader) (Parameters, error) {
	count, err := ReadVarint(r)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	params := make(Parameters, count)
	for range count {
		key, err := ReadVarint(r)
		if err != nil {
			return nil, err
		}

		value, err := ReadBytes(r)
		if err != nil {
			return nil, err
		}

		params[key] = value
	}

	return params, nil
}

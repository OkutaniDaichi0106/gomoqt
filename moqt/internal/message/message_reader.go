package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

// Reader interface for reading bytes
type reader interface {
	Read([]byte) (int, error)
	ReadByte() (byte, error)
}

// Read a number from the reader
func ReadNumber(r reader) (uint64, int, error) {
	num, err := quicvarint.Read(r)
	return num, quicvarint.Len(num), err
}

// Read a string from the reader
func ReadString(r reader) (string, int, error) {
	b, n, err := ReadBytes(r)
	if err != nil {
		return "", n, err
	}
	return string(b), n, nil
}

// Read a byte slice from the reader
func ReadBytes(r reader) ([]byte, int, error) {
	num, n, err := ReadNumber(r)
	if err != nil {
		return nil, n, err
	}

	b := make([]byte, num)
	n2, err := io.ReadFull(r, b)
	if err != nil {
		return nil, n + n2, err
	}

	return b, n + n2, nil
}

// // Read a string array from the reader
// func ReadStringArray(r reader) ([]string, int, error) {
// 	count, n, err := ReadNumber(r)
// 	if err != nil {
// 		if err == io.EOF {
// 			return nil, n, nil
// 		}
// 		return nil, n, err
// 	}

// 	strs := make([]string, count)
// 	totalBytes := n
// 	for i := uint64(0); i < count; i++ {
// 		str, n, err := ReadString(r)
// 		if err != nil {
// 			return nil, totalBytes + n, err
// 		}
// 		strs[i] = str
// 		totalBytes += n
// 	}

// 	return strs, totalBytes, nil
// }

// Read parameters from the reader
func ReadParameters(r reader) (Parameters, int, error) {
	count, n, err := ReadNumber(r)
	if err != nil {
		if err == io.EOF {
			return nil, n, nil
		}
		return nil, n, err
	}

	params := make(Parameters, count)
	totalBytes := n
	for i := uint64(0); i < count; i++ {
		key, n, err := ReadNumber(r)
		if err != nil {
			return nil, totalBytes + n, err
		}

		value, n, err := ReadBytes(r)
		if err != nil {
			return nil, totalBytes + n, err
		}

		params[key] = value
		totalBytes += n
	}

	return params, totalBytes, nil
}

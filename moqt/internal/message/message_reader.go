package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

// Reader interface for reading bytes
type reader interface {
	Read([]byte) (int, error)
	ReadByte() (byte, error)
}

// Create a new reader from an io.Reader
func newReader(r io.Reader) (reader, error) {
	// Get the length of the message
	num, err := readNumber(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read message length", slog.String("error", err.Error()))
		return nil, err
	}

	// Read the message into a byte slice
	buf := make([]byte, num)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		slog.Error("failed to read message", slog.String("error", err.Error()))
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

// Read a number from the reader
func readNumber(r reader) (uint64, error) {
	return quicvarint.Read(r)
}

// Read a string from the reader
func readString(r reader) (string, error) {
	b, err := readBytes(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Read a byte slice from the reader
func readBytes(r reader) ([]byte, error) {
	num, err := readNumber(r)
	if err != nil {
		return nil, err
	}

	b := make([]byte, num)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Read a string array from the reader
func readStringArray(r reader) ([]string, error) {
	count, err := readNumber(r)
	if err != nil {
		return nil, err
	}

	strs := make([]string, count)
	for i := uint64(0); i < count; i++ {
		str, err := readString(r)
		if err != nil {
			return nil, err
		}
		strs[i] = str
	}

	return strs, nil
}

// Read parameters from the reader
func readParameters(r reader) (Parameters, error) {
	count, err := readNumber(r)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	params := make(Parameters, count)
	for i := uint64(0); i < count; i++ {
		key, err := readNumber(r)
		if err != nil {
			return nil, err
		}

		value, err := readBytes(r)
		if err != nil {
			return nil, err
		}

		params[key] = value
	}

	return params, nil
}

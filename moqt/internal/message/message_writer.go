package message

import (
	"github.com/quic-go/quic-go/quicvarint"
)

// Append a number to the byte slice
func AppendNumber(to []byte, num uint64) []byte {
	return quicvarint.Append(to, num)
}

// Append a string to the byte slice
func AppendString(to []byte, str string) []byte {
	return AppendBytes(to, []byte(str))
}

// Append a byte slice to the byte slice
func AppendBytes(to []byte, bytes []byte) []byte {
	to = quicvarint.Append(to, uint64(len(bytes)))
	to = append(to, bytes...)
	return to
}

// Append a string array to the byte slice
func AppendStringArray(to []byte, arr []string) []byte {
	to = AppendNumber(to, uint64(len(arr)))
	for _, str := range arr {
		to = AppendString(to, str)
	}
	return to
}

// Append parameters to the byte slice
func AppendParameters(p []byte, params Parameters) []byte {
	if params == nil {
		return p
	}

	p = AppendNumber(p, uint64(len(params)))
	for key, value := range params {
		p = AppendNumber(p, key)
		p = AppendBytes(p, value)
	}
	return p
}

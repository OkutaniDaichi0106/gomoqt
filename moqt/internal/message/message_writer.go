package message

import (
	"github.com/quic-go/quic-go/quicvarint"
)

// Append a number to the byte slice
func appendNumber(p []byte, num uint64) []byte {
	return quicvarint.Append(p, num)
}

// Append a string to the byte slice
func appendString(p []byte, str string) []byte {
	return appendBytes(p, []byte(str))
}

// Append a byte slice to the byte slice
func appendBytes(p []byte, b []byte) []byte {
	p = quicvarint.Append(p, uint64(len(b)))
	p = append(p, b...)
	return p
}

// Append a string array to the byte slice
func appendStringArray(p []byte, arr []string) []byte {
	p = appendNumber(p, uint64(len(arr)))
	for _, str := range arr {
		p = appendString(p, str)
	}
	return p
}

// Append parameters to the byte slice
func appendParameters(p []byte, params Parameters) []byte {
	p = appendNumber(p, uint64(len(params)))
	for key, value := range params {
		p = appendNumber(p, key)
		p = appendBytes(p, value)
	}
	return p
}

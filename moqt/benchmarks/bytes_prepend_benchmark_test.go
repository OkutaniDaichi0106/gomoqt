package benchmarks_test

import (
	"encoding/binary"
	"testing"
)

var payload = []byte("example payload")

// Append
func prependWithAppend(payload []byte) []byte {
	length := uint16(len(payload))
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, length)
	return append(lengthBytes, payload...)
}

// Copy
func prependWithCopy(payload []byte) []byte {
	length := uint16(len(payload))
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, length)

	result := make([]byte, len(lengthBytes)+len(payload))
	copy(result, lengthBytes)
	copy(result[len(lengthBytes):], payload)
	return result
}

// Shift
func prependWithShift(payload []byte) []byte {
	length := uint16(len(payload))
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, length)

	result := make([]byte, len(lengthBytes)+len(payload))
	copy(result[:len(lengthBytes)], lengthBytes)
	copy(result[len(lengthBytes):], payload)
	return result
}

func BenchmarkPrependWithAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = prependWithAppend(payload)
	}
}

func BenchmarkPrependWithCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = prependWithCopy(payload)
	}
}

func BenchmarkPrependWithShift(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = prependWithShift(payload)
	}
}

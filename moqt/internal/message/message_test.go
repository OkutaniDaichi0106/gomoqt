package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarintLen(t *testing.T) {
	tests := map[string]struct {
		input    uint64
		expected int
	}{
		"zero":            {0, 1},
		"max varint1":     {maxVarInt1, 1},
		"max varint1 + 1": {maxVarInt1 + 1, 2},
		"max varint2":     {maxVarInt2, 2},
		"max varint2 + 1": {maxVarInt2 + 1, 4},
		"max varint4":     {maxVarInt4, 4},
		"max varint4 + 1": {maxVarInt4 + 1, 8},
		"max varint8":     {maxVarInt8, 8},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := VarintLen(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVarintLenPanic(t *testing.T) {
	assert.Panics(t, func() {
		VarintLen(maxVarInt8 + 1)
	})
}

func TestStringLen(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected int
	}{
		"empty string":     {"", VarintLen(0) + 0},
		"short string":     {"hello", VarintLen(5) + 5},
		"longer string":    {string(make([]byte, 300)), VarintLen(300) + 300},
		"very long string": {string(make([]byte, 70000)), VarintLen(70000) + 70000},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := StringLen(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBytesLen(t *testing.T) {
	tests := map[string]struct {
		input    []byte
		expected int
	}{
		"empty bytes":     {[]byte{}, VarintLen(0) + 0},
		"short bytes":     {[]byte("hello"), VarintLen(5) + 5},
		"longer bytes":    {make([]byte, 300), VarintLen(300) + 300},
		"very long bytes": {make([]byte, 70000), VarintLen(70000) + 70000},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := BytesLen(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringArrayLen(t *testing.T) {
	tests := map[string]struct {
		input    []string
		expected int
	}{
		"empty array":       {[]string{}, VarintLen(0)},
		"single element":    {[]string{"hello"}, VarintLen(1) + StringLen("hello")},
		"multiple elements": {[]string{"hello", "world", ""}, VarintLen(3) + StringLen("hello") + StringLen("world") + StringLen("")},
		"long strings":      {[]string{string(make([]byte, 100)), string(make([]byte, 200))}, VarintLen(2) + StringLen(string(make([]byte, 100))) + StringLen(string(make([]byte, 200)))},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := StringArrayLen(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParametersLen(t *testing.T) {
	tests := map[string]struct {
		input    map[uint64][]byte
		expected int
	}{
		"empty parameters": {map[uint64][]byte{}, VarintLen(0)},
		"single parameter": {map[uint64][]byte{1: []byte("value")}, VarintLen(1) + VarintLen(1) + BytesLen([]byte("value"))},
		"multiple parameters": {map[uint64][]byte{
			1: []byte("value1"),
			2: []byte("value2"),
		}, VarintLen(2) + VarintLen(1) + BytesLen([]byte("value1")) + VarintLen(2) + BytesLen([]byte("value2"))},
		"empty value": {map[uint64][]byte{1: []byte{}}, VarintLen(1) + VarintLen(1) + BytesLen([]byte{})},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ParametersLen(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

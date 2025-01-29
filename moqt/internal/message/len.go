package message

import "github.com/quic-go/quic-go/quicvarint"

func numberLen(num uint64) int {
	return quicvarint.Len(num)
}

func stringLen(s string) int {
	return bytesLen([]byte(s))
}

func bytesLen(b []byte) int {
	return numberLen(uint64(len(b))) + len(b)
}

func stringArrayLen(arr []string) int {
	if arr == nil {
		return 0
	}

	l := numberLen(uint64(len(arr)))
	for _, s := range arr {
		l += stringLen(s)
	}
	return l
}

func parametersLen(p Parameters) int {
	if p == nil {
		return 0
	}

	l := numberLen(uint64(len(p)))
	for k, v := range p {
		l += numberLen(k) + bytesLen(v)
	}
	return l
}

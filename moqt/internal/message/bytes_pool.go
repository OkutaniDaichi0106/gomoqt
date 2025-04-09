package message

import "sync"

var defaultBytesPool = &sync.Pool{
	New: func() any {
		return make([]byte, 0, 1<<8) // 64KB
	},
}

func GetBytes() []byte {
	b := defaultBytesPool.Get().([]byte)
	b = b[:0]
	return b
}

func PutBytes(b []byte) {
	b = b[:0]
	defaultBytesPool.Put(b)
}

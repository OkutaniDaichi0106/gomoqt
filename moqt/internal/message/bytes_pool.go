package message

import "sync"

var defaultBytesPool = &sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1<<8)
		return &b
	},
}

func GetBytes() []byte {
	b := defaultBytesPool.Get().(*[]byte)
	return (*b)[:0]
}

func PutBytes(b []byte) {
	b = b[:0]
	defaultBytesPool.Put(&b)
}

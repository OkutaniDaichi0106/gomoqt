package message

import "sync"

var defaultBytesPool = &sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 1<<8) // 64KB
		return &buf
	},
}

func GetBytes() *[]byte {
	b := defaultBytesPool.Get().(*[]byte)
	*b = (*b)[:0]
	return b
}

func PutBytes(b *[]byte) {
	*b = (*b)[:0]
	defaultBytesPool.Put(b)
}

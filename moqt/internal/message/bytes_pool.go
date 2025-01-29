package message

import "sync"

var defaultBytesPool = &sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 0)
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

package benchmarks_test

import (
	"sync"
	"testing"
)

type GroupDirect struct {
	payload []byte
}

type GroupPointer struct {
	payload *[]byte
}

var defaultPayload = []byte{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
	11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
	31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
	51, 52, 53, 54, 55, 56, 57, 58, 59, 60,
	61, 62, 63, 64, 65, 66, 67, 68, 69, 70,
	71, 72, 73, 74, 75, 76, 77, 78, 79, 80,
	81, 82, 83, 84, 85, 86, 87, 88, 89, 90,
	91, 92, 93, 94, 95, 96, 97, 98, 99, 100,
}

var (
	bytesPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	ptrBytesPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 1024)
			return &b
		},
	}

	groupDirectPool = sync.Pool{
		New: func() interface{} {
			return GroupDirect{
				payload: make([]byte, 1024),
			}
		},
	}

	ptrGroupDirectPool = sync.Pool{
		New: func() interface{} {
			return &GroupDirect{
				payload: make([]byte, 1024),
			}
		},
	}

	groupPointerPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 1024)
			return GroupPointer{
				payload: &b,
			}
		},
	}

	ptrGroupPointerPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 1024)
			return &GroupPointer{
				payload: &b,
			}
		},
	}
)

func BenchmarkPoolGetPut_Bytes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			buf := bytesPool.Get().([]byte)
			bytesPool.Put(buf[:0])
		}
	}
}

func BenchmarkPoolGetPut_PtrBytes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			bufPtr := ptrBytesPool.Get().(*[]byte)
			ptrBytesPool.Put(bufPtr)
		}
	}
}

func BenchmarkPoolGetPut_GroupDirect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			group := groupDirectPool.Get().(GroupDirect)
			groupDirectPool.Put(group)
		}
	}
}

func BenchmarkPoolGetPut_PtrGroupDirect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			group := ptrGroupDirectPool.Get().(*GroupDirect)
			ptrGroupDirectPool.Put(group)
		}
	}
}

func BenchmarkPoolGetPut_GroupPointer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			group := groupPointerPool.Get().(GroupPointer)
			groupPointerPool.Put(group)
		}
	}
}

func BenchmarkPoolGetPut_PtrGroupPointer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			group := ptrGroupPointerPool.Get().(*GroupPointer)
			ptrGroupPointerPool.Put(group)
		}
	}
}

func BenchmarkPoolNew_Bytes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bytesPool.New()
	}
}

func BenchmarkPoolNew_PtrBytes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ptrBytesPool.New()
	}
}

func BenchmarkPoolNew_GroupDirect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		groupDirectPool.New()
	}
}

func BenchmarkPoolNew_PtrGroupDirect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptrGroupDirectPool.New()
	}
}

func BenchmarkPoolNew_GroupPointer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		groupPointerPool.New()
	}
}

func BenchmarkPoolNew_PtrGroupPointer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptrGroupPointerPool.New()
	}
}

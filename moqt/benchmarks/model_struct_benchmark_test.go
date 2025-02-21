package benchmarks_test

import (
	"sync"
	"testing"
)

const testData = "benchmark"

// Updated Frame definition with a []byte field.
type Frame struct {
	Data []byte // Frame now holds a byte slice.
	// ...existing code...
}

type Group struct {
	frames []*Frame
	// ...existing code...
}

func BenchmarkNewFrameValue(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return Frame{}
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Get().(Frame)
	}
}

func BenchmarkNewFramePointer(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return new(Frame)
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Get().(*Frame)
	}
}

func BenchmarkNewSliceOfFramePointer(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return []*Frame{}
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Get().([]*Frame)
	}
}

func BenchmarkNewGroupValue(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return Group{}
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Get().(Group)
	}
}

func BenchmarkNewGroupPointer(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return new(Group)
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Get().(*Group)
	}
}

func BenchmarkFrameRoundTripDataWrite(b *testing.B) {
	cases := []struct {
		name    string
		pool    *sync.Pool
		getFunc func(interface{}) *Frame
	}{
		{
			name: "FrameValue",
			pool: &sync.Pool{
				New: func() interface{} { return Frame{} },
			},
			// For a Frame value, we work on a copy.
			getFunc: func(obj interface{}) *Frame {
				f := obj.(Frame)
				return &f
			},
		},
		{
			name: "FramePointer",
			pool: &sync.Pool{
				New: func() interface{} { return new(Frame) },
			},
			getFunc: func(obj interface{}) *Frame {
				return obj.(*Frame)
			},
		},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				obj := tc.pool.Get()
				frame := tc.getFunc(obj)
				frame.Data = []byte(testData)
				tc.pool.Put(obj)
			}
		})
	}
}

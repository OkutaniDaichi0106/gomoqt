package moqt

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// BenchmarkAnnouncement_NewAnnouncement benchmarks announcement creation
func BenchmarkAnnouncement_NewAnnouncement(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")

	for b.Loop() {
		ann, end := NewAnnouncement(ctx, path)
		_ = ann
		_ = end
	}
}

// BenchmarkAnnouncement_End benchmarks announcement ending
func BenchmarkAnnouncement_End(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")

	for b.Loop() {
		b.StopTimer()
		ann, end := NewAnnouncement(ctx, path)
		b.StartTimer()

		end()
		_ = ann
	}
}

// BenchmarkAnnouncement_AfterFunc benchmarks AfterFunc registration
func BenchmarkAnnouncement_AfterFunc(b *testing.B) {
	sizes := []int{1, 10, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("handlers-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ctx := context.Background()
				path := BroadcastPath("/test/path")
				ann, end := NewAnnouncement(ctx, path)
				b.StartTimer()

				// Register multiple handlers
				for range size {
					ann.AfterFunc(func() {})
				}

				b.StopTimer()
				end()
				b.StartTimer()
			}
		})
	}
}

// BenchmarkAnnouncement_Done benchmarks Done channel operations
func BenchmarkAnnouncement_Done(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")
	ann, end := NewAnnouncement(ctx, path)

	// End the announcement to close the channel
	end()

	for b.Loop() {
		<-ann.Done()
	}
}

// BenchmarkAnnouncement_IsActive benchmarks IsActive checks
func BenchmarkAnnouncement_IsActive(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")
	ann, _ := NewAnnouncement(ctx, path)

	for b.Loop() {
		_ = ann.IsActive()
	}
}

// BenchmarkAnnouncement_BroadcastPath benchmarks BroadcastPath retrieval
func BenchmarkAnnouncement_BroadcastPath(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")
	ann, _ := NewAnnouncement(ctx, path)

	for b.Loop() {
		_ = ann.BroadcastPath()
	}
}

// BenchmarkAnnouncement_ConcurrentAfterFunc benchmarks concurrent AfterFunc calls
func BenchmarkAnnouncement_ConcurrentAfterFunc(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")
	ann, end := NewAnnouncement(ctx, path)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ann.AfterFunc(func() {})
		}
	})

	end()
}

// BenchmarkAnnouncement_ConcurrentIsActive benchmarks concurrent IsActive checks
func BenchmarkAnnouncement_ConcurrentIsActive(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	path := BroadcastPath("/test/path")
	ann, end := NewAnnouncement(ctx, path)
	defer end()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ann.IsActive()
		}
	})
}

// BenchmarkAnnouncement_EndWithHandlers benchmarks ending with multiple handlers
func BenchmarkAnnouncement_EndWithHandlers(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("handlers-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ctx := context.Background()
				path := BroadcastPath("/test/path")
				ann, end := NewAnnouncement(ctx, path)

				// Register handlers
				var wg sync.WaitGroup
				wg.Add(size)
				for range size {
					ann.AfterFunc(func() {
						wg.Done()
					})
				}

				b.StartTimer()
				end()
				wg.Wait()
			}
		})
	}
}

// BenchmarkAnnouncement_StopFunc benchmarks stop function execution
func BenchmarkAnnouncement_StopFunc(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx := context.Background()
		path := BroadcastPath("/test/path")
		ann, end := NewAnnouncement(ctx, path)

		stop := ann.AfterFunc(func() {})
		b.StartTimer()

		stop()

		b.StopTimer()
		end()
		b.StartTimer()
	}
}

// BenchmarkAnnouncement_MemoryAllocation benchmarks overall memory usage
func BenchmarkAnnouncement_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		path := BroadcastPath(fmt.Sprintf("/test/path/%d", i))
		ann, end := NewAnnouncement(ctx, path)

		// Register some handlers
		for range 5 {
			ann.AfterFunc(func() {})
		}

		// Check state
		_ = ann.IsActive()
		_ = ann.BroadcastPath()

		end()
	}
}

// BenchmarkAnnouncement_PathValidation benchmarks path validation overhead
func BenchmarkAnnouncement_PathValidation(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	// Pre-generate valid paths
	paths := make([]BroadcastPath, 100)
	for i := range paths {
		paths[i] = BroadcastPath(fmt.Sprintf("/test/path/%d", i))
	}

	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		ann, end := NewAnnouncement(ctx, path)
		_ = ann
		end()
	}
}

package moqt

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/mock"
)

// BenchmarkTrackMux_NewTrackMux benchmarks TrackMux creation
func BenchmarkTrackMux_NewTrackMux(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mux := NewTrackMux()
		_ = mux
	}
}

// BenchmarkTrackMux_Handle benchmarks handler registration
func BenchmarkTrackMux_Handle(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			mux := NewTrackMux()
			ctx := context.Background()
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Pre-generate paths to avoid string generation overhead during benchmark
			paths := make([]BroadcastPath, size)
			for i := 0; i < size; i++ {
				paths[i] = BroadcastPath(fmt.Sprintf("/path/%d", i))
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Use modulo to cycle through paths for repeated benchmarks
				path := paths[i%size]
				mux.Handle(ctx, path, handler)
			}
		})
	}
}

// BenchmarkTrackMux_Handler benchmarks handler lookup
func BenchmarkTrackMux_Handler(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			mux := NewTrackMux()
			ctx := context.Background()
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Pre-populate with handlers
			paths := make([]BroadcastPath, size)
			for i := 0; i < size; i++ {
				path := BroadcastPath(fmt.Sprintf("/path/%d", i))
				paths[i] = path
				mux.Handle(ctx, path, handler)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				path := paths[i%size]
				_ = mux.Handler(path)
			}
		})
	}
}

// BenchmarkTrackMux_ServeTrack benchmarks track serving
func BenchmarkTrackMux_ServeTrack(b *testing.B) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test/path")

	// Register a simple handler
	mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {
		// Simple no-op handler for benchmarking
	}))

	// Create a test track writer
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	onCloseTrack := func() {}
	trackWriter := newTrackWriter(path, TrackName("test_track"), nil, openUniStreamFunc, onCloseTrack)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mux.ServeTrack(trackWriter)
	}
}

// BenchmarkTrackMux_ServeAnnouncements benchmarks announcement serving
func BenchmarkTrackMux_ServeAnnouncements(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			mux := NewTrackMux()
			ctx := context.Background()
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Pre-populate with handlers under /room/ prefix
			for i := 0; i < size; i++ {
				path := BroadcastPath(fmt.Sprintf("/room/user%d", i))
				mux.Handle(ctx, path, handler)
			}

			// Create mock announcement writer
			mockWriter := &MockAnnouncementWriter{}
			mockWriter.On("SendAnnouncement", mock.Anything).Return(nil)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				mux.ServeAnnouncements(mockWriter, "/room/")
			}
		})
	}
}

// BenchmarkTrackMux_ConcurrentRead benchmarks concurrent handler lookups
func BenchmarkTrackMux_ConcurrentRead(b *testing.B) {
	mux := NewTrackMux()
	ctx := context.Background()
	handler := TrackHandlerFunc(func(tw *TrackWriter) {})

	// Pre-populate with handlers
	const numPaths = 1000
	paths := make([]BroadcastPath, numPaths)
	for i := 0; i < numPaths; i++ {
		path := BroadcastPath(fmt.Sprintf("/path/%d", i))
		paths[i] = path
		mux.Handle(ctx, path, handler)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := paths[i%numPaths]
			_ = mux.Handler(path)
			i++
		}
	})
}

// BenchmarkTrackMux_ConcurrentWrite benchmarks concurrent handler registration
func BenchmarkTrackMux_ConcurrentWrite(b *testing.B) {
	handler := TrackHandlerFunc(func(tw *TrackWriter) {})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mux := NewTrackMux()
			ctx := context.Background()
			path := BroadcastPath(fmt.Sprintf("/path/%d", i))
			mux.Handle(ctx, path, handler)
			i++
		}
	})
}

// BenchmarkTrackMux_MixedWorkload benchmarks mixed read/write operations
func BenchmarkTrackMux_MixedWorkload(b *testing.B) {
	mux := NewTrackMux()
	ctx := context.Background()
	handler := TrackHandlerFunc(func(tw *TrackWriter) {})

	// Pre-populate with some handlers
	const initialPaths = 500
	paths := make([]BroadcastPath, initialPaths)
	for i := 0; i < initialPaths; i++ {
		path := BroadcastPath(fmt.Sprintf("/existing/%d", i))
		paths[i] = path
		mux.Handle(ctx, path, handler)
	}

	var writeCounter int64

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			if localCounter%10 == 0 {
				// 10% writes - new handler registration
				newPath := BroadcastPath(fmt.Sprintf("/new/%d", writeCounter))
				mux.Handle(ctx, newPath, handler)
				writeCounter++
			} else {
				// 90% reads - handler lookup
				path := paths[localCounter%initialPaths]
				_ = mux.Handler(path)
			}
			localCounter++
		}
	})
}

// BenchmarkTrackMux_DeepNestedPaths benchmarks performance with deeply nested paths
func BenchmarkTrackMux_DeepNestedPaths(b *testing.B) {
	depths := []int{5, 10, 20}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			mux := NewTrackMux()
			ctx := context.Background()
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Create deeply nested path
			pathBuilder := "/root"
			for i := 0; i < depth; i++ {
				pathBuilder += fmt.Sprintf("/level%d", i)
			}
			path := BroadcastPath(pathBuilder)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if i%2 == 0 {
					mux.Handle(ctx, path, handler)
				} else {
					_ = mux.Handler(path)
				}
			}
		})
	}
}

// BenchmarkTrackMux_MemoryUsage benchmarks memory usage with different numbers of handlers
func BenchmarkTrackMux_MemoryUsage(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("handlers-%d", size), func(b *testing.B) {
			var m1, m2 runtime.MemStats

			b.ReportAllocs()
			runtime.GC()
			runtime.ReadMemStats(&m1)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				mux := NewTrackMux()
				ctx := context.Background()
				handler := TrackHandlerFunc(func(tw *TrackWriter) {})

				// Register many handlers
				for j := 0; j < size; j++ {
					path := BroadcastPath(fmt.Sprintf("/path/%d/%d", i, j))
					mux.Handle(ctx, path, handler)
				}

				// Perform some operations to measure realistic memory usage
				for j := 0; j < 100; j++ {
					lookupPath := BroadcastPath(fmt.Sprintf("/path/%d/%d", i, j%size))
					_ = mux.Handler(lookupPath)
				}
			}

			b.StopTimer()
			runtime.GC()
			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "allocs/op")
		})
	}
}

// Benchmark helper that measures allocation patterns for specific operations
func BenchmarkTrackMux_AllocationPatterns(b *testing.B) {
	b.Run("handler-registration", func(b *testing.B) {
		mux := NewTrackMux()
		ctx := context.Background()
		handler := TrackHandlerFunc(func(tw *TrackWriter) {})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			path := BroadcastPath(fmt.Sprintf("/alloc/test/%d", i))
			mux.Handle(ctx, path, handler)
		}
	})

	b.Run("handler-lookup", func(b *testing.B) {
		mux := NewTrackMux()
		ctx := context.Background()
		handler := TrackHandlerFunc(func(tw *TrackWriter) {})

		// Pre-register handlers
		paths := make([]BroadcastPath, 1000)
		for i := 0; i < 1000; i++ {
			path := BroadcastPath(fmt.Sprintf("/alloc/lookup/%d", i))
			paths[i] = path
			mux.Handle(ctx, path, handler)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			path := paths[i%1000]
			_ = mux.Handler(path)
		}
	})
}

// BenchmarkTrackMux_StringOperations benchmarks string operations impact
func BenchmarkTrackMux_StringOperations(b *testing.B) {
	b.Run("path-validation", func(b *testing.B) {
		paths := []BroadcastPath{
			BroadcastPath("/valid/path"),
			BroadcastPath("/another/valid/path"),
			BroadcastPath("/deep/nested/valid/path"),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			path := paths[i%len(paths)]
			_ = isValidPath(path)
		}
	})

	b.Run("prefix-validation", func(b *testing.B) {
		prefixes := []string{
			"/valid/",
			"/another/",
			"/deep/nested/",
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			prefix := prefixes[i%len(prefixes)]
			_ = isValidPrefix(prefix)
		}
	})

	b.Run("path-splitting", func(b *testing.B) {
		paths := []BroadcastPath{
			BroadcastPath("/level1/level2/track"),
			BroadcastPath("/room/user123/video"),
			BroadcastPath("/game/match456/audio"),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			path := paths[i%len(paths)]
			_ = strings.Split(string(path), "/")
		}
	})
}

// BenchmarkTrackMux_LockContention benchmarks mutex contention scenarios
func BenchmarkTrackMux_LockContention(b *testing.B) {
	mux := NewTrackMux()
	ctx := context.Background()
	handler := TrackHandlerFunc(func(tw *TrackWriter) {})

	// Pre-populate with handlers
	const numPaths = 1000
	for i := 0; i < numPaths; i++ {
		path := BroadcastPath(fmt.Sprintf("/path/%d", i))
		mux.Handle(ctx, path, handler)
	}

	b.Run("read-heavy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				path := BroadcastPath(fmt.Sprintf("/path/%d", i%numPaths))
				_ = mux.Handler(path)
				i++
			}
		})
	})

	b.Run("write-heavy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				newMux := NewTrackMux()
				path := BroadcastPath(fmt.Sprintf("/new/%d", i))
				newMux.Handle(ctx, path, handler)
				i++
			}
		})
	})

	b.Run("mixed-contention", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%20 == 0 {
					// 5% writes
					path := BroadcastPath(fmt.Sprintf("/contention/%d", i))
					mux.Handle(ctx, path, handler)
				} else {
					// 95% reads
					path := BroadcastPath(fmt.Sprintf("/path/%d", i%numPaths))
					_ = mux.Handler(path)
				}
				i++
			}
		})
	})
}

// BenchmarkTrackMux_AnnouncementTree benchmarks the announcement tree operations
func BenchmarkTrackMux_AnnouncementTree(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("tree-traversal-size-%d", size), func(b *testing.B) {
			mux := NewTrackMux()
			ctx := context.Background()
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Create a deep tree structure
			for i := 0; i < size; i++ {
				path := BroadcastPath(fmt.Sprintf("/level1/level2/level3/track%d", i))
				mux.Handle(ctx, path, handler)
			}

			mockWriter := &MockAnnouncementWriter{}
			mockWriter.On("SendAnnouncement", mock.Anything).Return(nil)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				mux.ServeAnnouncements(mockWriter, "/level1/level2/")
			}
		})
	}
}

// BenchmarkTrackMux_MapOperations benchmarks raw map performance
func BenchmarkTrackMux_MapOperations(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("map-lookup-size-%d", size), func(b *testing.B) {
			// Create a map similar to handlerIndex
			handlerMap := make(map[BroadcastPath]TrackHandler, size)
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Pre-populate the map
			for i := 0; i < size; i++ {
				path := BroadcastPath(fmt.Sprintf("/path/%d", i))
				handlerMap[path] = handler
			}

			// Prepare lookup paths
			lookupPaths := make([]BroadcastPath, 1000)
			for i := 0; i < 1000; i++ {
				lookupPaths[i] = BroadcastPath(fmt.Sprintf("/path/%d", i%size))
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				path := lookupPaths[i%1000]
				_ = handlerMap[path]
			}
		})
	}
}

// BenchmarkTrackMux_GCPressure benchmarks GC impact with different allocation patterns
func BenchmarkTrackMux_GCPressure(b *testing.B) {
	b.Run("frequent-mux-creation", func(b *testing.B) {
		ctx := context.Background()
		handler := TrackHandlerFunc(func(tw *TrackWriter) {})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			mux := NewTrackMux()
			for j := 0; j < 10; j++ {
				path := BroadcastPath(fmt.Sprintf("/temp/%d/%d", i, j))
				mux.Handle(ctx, path, handler)
			}
			// Let mux go out of scope for GC
		}
	})

	b.Run("long-lived-mux", func(b *testing.B) {
		mux := NewTrackMux()
		ctx := context.Background()
		handler := TrackHandlerFunc(func(tw *TrackWriter) {})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			path := BroadcastPath(fmt.Sprintf("/persistent/%d", i))
			mux.Handle(ctx, path, handler)

			// Periodic cleanup to simulate real usage
			if i%1000 == 999 {
				mux.Clear()
			}
		}
	})
}

// BenchmarkTrackMux_RealWorldScenarios benchmarks realistic usage patterns
func BenchmarkTrackMux_RealWorldScenarios(b *testing.B) {
	b.Run("broadcast-server", func(b *testing.B) {
		mux := NewTrackMux()
		ctx := context.Background()
		handler := TrackHandlerFunc(func(tw *TrackWriter) {
			// Simulate some work
			for i := 0; i < 10; i++ {
				_ = fmt.Sprintf("processing-%d", i)
			}
		})

		// Simulate broadcast server with multiple rooms
		const numRooms = 100
		const usersPerRoom = 50

		// Pre-register room handlers
		for room := 0; room < numRooms; room++ {
			for user := 0; user < usersPerRoom; user++ {
				path := BroadcastPath(fmt.Sprintf("/room/%d/user/%d", room, user))
				mux.Handle(ctx, path, handler)
			}
		}

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Simulate mixed operations
				switch i % 100 {
				case 0, 1, 2: // 3% - new user joins
					room := i % numRooms
					newUser := usersPerRoom + (i / 100)
					path := BroadcastPath(fmt.Sprintf("/room/%d/user/%d", room, newUser))
					mux.Handle(ctx, path, handler)
				case 3, 4: // 2% - announcement subscription
					room := i % numRooms
					mockWriter := &MockAnnouncementWriter{}
					mockWriter.On("SendAnnouncement", mock.Anything).Return(nil)
					prefix := fmt.Sprintf("/room/%d/", room)
					mux.ServeAnnouncements(mockWriter, prefix)
				default: // 95% - track serving
					room := i % numRooms
					user := i % usersPerRoom
					path := BroadcastPath(fmt.Sprintf("/room/%d/user/%d", room, user))
					trackWriter := newTrackWriter(path, TrackName(fmt.Sprintf("track-%d", i)), nil, func() (quic.SendStream, error) {
						mockSendStream := &MockQUICSendStream{}
						mockSendStream.On("CancelWrite", mock.Anything).Return()
						mockSendStream.On("StreamID").Return(quic.StreamID(1))
						mockSendStream.On("Close").Return(nil)
						mockSendStream.On("Write", mock.Anything).Return(0, nil)
						return mockSendStream, nil
					}, func() {})
					mux.ServeTrack(trackWriter)
				}
				i++
			}
		})
	})

	b.Run("live-streaming", func(b *testing.B) {
		mux := NewTrackMux()
		ctx := context.Background()
		handler := TrackHandlerFunc(func(tw *TrackWriter) {
			// Simulate stream processing
		})

		// Simulate live streaming with multiple quality levels
		const numStreams = 200
		qualities := []string{"low", "medium", "high", "4k"}

		// Pre-register stream handlers
		for stream := 0; stream < numStreams; stream++ {
			for _, quality := range qualities {
				path := BroadcastPath(fmt.Sprintf("/stream/%d/quality/%s", stream, quality))
				mux.Handle(ctx, path, handler)
			}
		}

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				stream := i % numStreams
				quality := qualities[i%len(qualities)]
				path := BroadcastPath(fmt.Sprintf("/stream/%d/quality/%s", stream, quality))
				trackWriter := newTrackWriter(path, TrackName(fmt.Sprintf("track-%d", i)), nil, func() (quic.SendStream, error) {
					mockSendStream := &MockQUICSendStream{}
					mockSendStream.On("CancelWrite", mock.Anything).Return()
					mockSendStream.On("StreamID").Return(quic.StreamID(1))
					mockSendStream.On("Close").Return(nil)
					mockSendStream.On("Write", mock.Anything).Return(0, nil)
					return mockSendStream, nil
				}, func() {})
				mux.ServeTrack(trackWriter)
				i++
			}
		})
	})
}

// BenchmarkTrackMux_CPUProfileOptimization provides specific scenarios for CPU profiling
func BenchmarkTrackMux_CPUProfileOptimization(b *testing.B) {
	if !testing.Short() {
		b.Run("cpu-hotspots", func(b *testing.B) {
			mux := NewTrackMux()
			ctx := context.Background()
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

			// Create a scenario that will show up clearly in CPU profiles
			const pathDepth = 10
			const pathCount = 1000

			// Register deeply nested paths
			for i := 0; i < pathCount; i++ {
				var pathBuilder strings.Builder
				for depth := 0; depth < pathDepth; depth++ {
					pathBuilder.WriteString(fmt.Sprintf("/level%d", depth))
				}
				pathBuilder.WriteString(fmt.Sprintf("/track%d", i))
				path := BroadcastPath(pathBuilder.String())
				mux.Handle(ctx, path, handler)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Mix of operations that will show up in CPU profile
				pathIndex := i % pathCount
				var pathBuilder strings.Builder
				for depth := 0; depth < pathDepth; depth++ {
					pathBuilder.WriteString(fmt.Sprintf("/level%d", depth))
				}
				pathBuilder.WriteString(fmt.Sprintf("/track%d", pathIndex))
				path := BroadcastPath(pathBuilder.String())

				// Operations that will consume CPU cycles
				_ = mux.Handler(path) // Map lookup
				trackWriter := newTrackWriter(path, TrackName(fmt.Sprintf("track-%d", i)), nil, func() (quic.SendStream, error) {
					mockSendStream := &MockQUICSendStream{}
					mockSendStream.On("CancelWrite", mock.Anything).Return()
					mockSendStream.On("StreamID").Return(quic.StreamID(1))
					mockSendStream.On("Close").Return(nil)
					mockSendStream.On("Write", mock.Anything).Return(0, nil)
					return mockSendStream, nil
				}, func() {})
				mux.ServeTrack(trackWriter)
			}
		})
	}
}

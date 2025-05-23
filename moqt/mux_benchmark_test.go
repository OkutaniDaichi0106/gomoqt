// Package moqt_test provides benchmarks for the moqt package.
package moqt_test

import (
	"context"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

// BenchmarkPathMatching benchmarks the performance of path matching
func BenchmarkPathMatching(b *testing.B) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Register many handlers
	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			path := moqt.BroadcastPath("/section" + string(rune(i+'0')) + "/subsection" + string(rune(j+'0')))
			mux.Handle(ctx, path, &moqt.MockTrackHandler{})
		}
	}

	// Test a deeply nested path
	writer := &moqt.MockTrackWriter{PathValue: "/section5/subsection7"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.ServeTrack(writer, &moqt.SubscribeConfig{})
	}
}

// BenchmarkHandlerRegistration benchmarks the performance of handler registration
func BenchmarkHandlerRegistration(b *testing.B) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()
	handler := &moqt.MockTrackHandler{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := moqt.BroadcastPath("/bench/path" + string(rune(i%10+'0')))
		mux.Handle(ctx, path, handler)
	}
}

// BenchmarkConcurrentPathMatching benchmarks the performance of concurrent path matching
func BenchmarkConcurrentPathMatching(b *testing.B) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Register paths
	for i := 0; i < 100; i++ {
		path := moqt.BroadcastPath("/section" + string(rune(i%10+'0')))
		mux.Handle(ctx, path, &moqt.MockTrackHandler{})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		writer := &moqt.MockTrackWriter{PathValue: "/section5"}
		for pb.Next() {
			mux.ServeTrack(writer, &moqt.SubscribeConfig{})
		}
	})
}

// BenchmarkNodeCleanup benchmarks the performance of node cleanup after context cancellation
func BenchmarkNodeCleanup(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mux := moqt.NewTrackMux()
		contexts := make([]context.CancelFunc, 0, 100)

		// Register many paths
		for j := 0; j < 100; j++ {
			ctx, cancel := context.WithCancel(context.Background())
			contexts = append(contexts, cancel)

			path := moqt.BroadcastPath("/section" + string(rune(j%10+'0')) + "/cleanup")
			mux.Handle(ctx, path, &moqt.MockTrackHandler{})
		}

		// Cancel all contexts to trigger cleanup
		for _, cancel := range contexts {
			cancel()
		}
	}
}

// BenchmarkWildcardMatching benchmarks the performance of wildcard pattern matching
func BenchmarkWildcardMatching(b *testing.B) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Register handlers with wildcard patterns
	mux.Handle(ctx, "/wildcard/single/*", &moqt.MockTrackHandler{})
	mux.Handle(ctx, "/wildcard/double/**", &moqt.MockTrackHandler{})

	// Test paths for matching
	singleWriter := &moqt.MockTrackWriter{PathValue: "/wildcard/single/match"}
	doubleWriter := &moqt.MockTrackWriter{PathValue: "/wildcard/double/multi/level/match"}

	b.Run("SingleWildcard", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mux.ServeTrack(singleWriter, &moqt.SubscribeConfig{})
		}
	})

	b.Run("DoubleWildcard", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mux.ServeTrack(doubleWriter, &moqt.SubscribeConfig{})
		}
	})
}

// BenchmarkOverwriteHandler benchmarks the performance of handler overwriting
func BenchmarkOverwriteHandler(b *testing.B) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()
	path := moqt.BroadcastPath("/overwrite/test")

	// Create a set of handlers to rotate through
	handlers := make([]*moqt.MockTrackHandler, 10)
	for i := 0; i < 10; i++ {
		handlers[i] = &moqt.MockTrackHandler{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Overwrite the same path with different handlers
		mux.Handle(ctx, path, handlers[i%10])
	}
}

// BenchmarkDeepPathTraversal benchmarks the performance of traversing deep path hierarchies
func BenchmarkDeepPathTraversal(b *testing.B) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Create deeply nested path
	deepPath := moqt.BroadcastPath("/level1/level2/level3/level4/level5/level6/level7/level8/level9/level10")
	handler := &moqt.MockTrackHandler{}
	mux.Handle(ctx, deepPath, handler)

	// Create writer for the path
	writer := &moqt.MockTrackWriter{PathValue: deepPath}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.ServeTrack(writer, &moqt.SubscribeConfig{})
	}
}

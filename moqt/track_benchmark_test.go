package moqt

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/okdaichi/gomoqt/quic"
	"github.com/stretchr/testify/mock"
)

// BenchmarkTrackReader_EnqueueDequeue benchmarks group enqueue and dequeue operations
func BenchmarkTrackReader_EnqueueDequeue(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			substr := newSendSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{}, Info{})
			reader := newTrackReader("broadcastPath", "trackName", substr, func() {})

			// Pre-create mock receive streams
			streams := make([]quic.ReceiveStream, size)
			for i := 0; i < size; i++ {
				mockRecvStream := &MockQUICReceiveStream{}
				mockRecvStream.On("Context").Return(context.Background())
				streams[i] = mockRecvStream
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				idx := i % size

				// Enqueue
				reader.enqueueGroup(GroupSequence(idx), streams[idx])

				// Dequeue
				group := reader.dequeueGroup()
				if group != nil {
					reader.removeGroup(group)
				}
			}
		})
	}
}

// BenchmarkTrackReader_AcceptGroup benchmarks accepting groups with queued data
func BenchmarkTrackReader_AcceptGroup(b *testing.B) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()
	mockStream.On("Context").Return(ctx)
	substr := newSendSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{}, Info{})
	reader := newTrackReader("broadcastPath", "trackName", substr, func() {})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Enqueue a group for this iteration
		mockRecvStream := &MockQUICReceiveStream{}
		mockRecvStream.On("Context").Return(ctx)
		reader.enqueueGroup(GroupSequence(i), mockRecvStream)

		// Accept it immediately (non-blocking since queue has data)
		group, err := reader.AcceptGroup(ctx)
		if err == nil && group != nil {
			reader.removeGroup(group)
		}
	}
}

// BenchmarkTrackReader_ConcurrentAccess benchmarks concurrent enqueue/dequeue operations
func BenchmarkTrackReader_ConcurrentAccess(b *testing.B) {
	concurrency := []int{2, 10, 50}

	for _, conc := range concurrency {
		b.Run(fmt.Sprintf("goroutines-%d", conc), func(b *testing.B) {
			mockStream := &MockQUICStream{}
			ctx := context.Background()
			mockStream.On("Context").Return(ctx)
			substr := newSendSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{}, Info{})
			reader := newTrackReader("broadcastPath", "trackName", substr, func() {})

			// Pre-populate queue
			for i := 0; i < 100; i++ {
				mockRecvStream := &MockQUICReceiveStream{}
				mockRecvStream.On("Context").Return(ctx)
				reader.enqueueGroup(GroupSequence(i), mockRecvStream)
			}

			b.ReportAllocs()
			b.ResetTimer()

			var wg sync.WaitGroup
			wg.Add(conc)

			for g := 0; g < conc; g++ {
				go func(id int) {
					defer wg.Done()
					for i := 0; i < b.N/conc; i++ {
						if id%2 == 0 {
							// Enqueue
							mockRecvStream := &MockQUICReceiveStream{}
							mockRecvStream.On("Context").Return(ctx)
							reader.enqueueGroup(GroupSequence(i+id*1000), mockRecvStream)
						} else {
							// Dequeue
							group := reader.dequeueGroup()
							if group != nil {
								reader.removeGroup(group)
							}
						}
					}
				}(g)
			}

			wg.Wait()
		})
	}
}

// BenchmarkTrackWriter_OpenGroup benchmarks opening groups
func BenchmarkTrackWriter_OpenGroup(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("groups-%d", size), func(b *testing.B) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("StreamID").Return(quic.StreamID(1))
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			mockStream.On("Write", mock.Anything).Return(0, nil)
			mockStream.On("Close").Return(nil)
			mockStream.On("Close").Return(nil)

			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

			streamIdx := 0
			var streamMu sync.Mutex
			openUniStreamFunc := func() (quic.SendStream, error) {
				streamMu.Lock()
				defer streamMu.Unlock()

				mockSendStream := &MockQUICSendStream{}
				mockSendStream.On("Context").Return(context.Background())
				mockSendStream.On("CancelWrite", mock.Anything).Return()
				mockSendStream.On("StreamID").Return(quic.StreamID(streamIdx))
				streamIdx++
				mockSendStream.On("Close").Return(nil)
				mockSendStream.WriteFunc = func(p []byte) (int, error) {
					return len(p), nil
				}
				return mockSendStream, nil
			}

			writer := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, func() {})

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				group, err := writer.OpenGroup(GroupSequence(i % size))
				if err == nil && group != nil {
					_ = group.Close()
				}
			}

			b.StopTimer()
			_ = writer.Close()
		})
	}
}

// BenchmarkTrackWriter_ConcurrentOpenGroup benchmarks concurrent group opening
func BenchmarkTrackWriter_ConcurrentOpenGroup(b *testing.B) {
	concurrency := []int{2, 10, 50}

	for _, conc := range concurrency {
		b.Run(fmt.Sprintf("goroutines-%d", conc), func(b *testing.B) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("StreamID").Return(quic.StreamID(1))
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			mockStream.On("Write", mock.Anything).Return(0, nil)
			mockStream.On("Close").Return(nil)
			mockStream.On("Close").Return(nil)

			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

			var streamIdx int64
			var streamMu sync.Mutex
			openUniStreamFunc := func() (quic.SendStream, error) {
				streamMu.Lock()
				defer streamMu.Unlock()

				mockSendStream := &MockQUICSendStream{}
				mockSendStream.On("Context").Return(context.Background())
				mockSendStream.On("CancelWrite", mock.Anything).Return()
				mockSendStream.On("StreamID").Return(quic.StreamID(streamIdx))
				streamIdx++
				mockSendStream.On("Close").Return(nil)
				mockSendStream.WriteFunc = func(p []byte) (int, error) {
					return len(p), nil
				}
				return mockSendStream, nil
			}

			writer := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, func() {})

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					group, err := writer.OpenGroup(GroupSequence(i))
					if err == nil && group != nil {
						_ = group.Close()
					}
					i++
				}
			})

			b.StopTimer()
			_ = writer.Close()
		})
	}
}

// BenchmarkTrackWriter_ActiveGroupManagement benchmarks active group map operations
func BenchmarkTrackWriter_ActiveGroupManagement(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("StreamID").Return(quic.StreamID(1))
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			mockStream.On("Write", mock.Anything).Return(0, nil)
			mockStream.On("Close").Return(nil)

			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

			streamIdx := 0
			openUniStreamFunc := func() (quic.SendStream, error) {
				mockSendStream := &MockQUICSendStream{}
				mockSendStream.WriteFunc = func(p []byte) (int, error) {
					return len(p), nil
				}
				mockSendStream.On("Context").Return(context.Background())
				mockSendStream.On("CancelWrite", mock.Anything).Return()
				mockSendStream.On("StreamID").Return(quic.StreamID(streamIdx))
				streamIdx++
				mockSendStream.On("Close").Return(nil)
				return mockSendStream, nil
			}

			writer := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, func() {})

			// Pre-create groups
			groups := make([]*GroupWriter, size)
			for i := 0; i < size; i++ {
				group, _ := writer.OpenGroup(GroupSequence(i))
				groups[i] = group
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				idx := i % size

				// Close and re-open group
				if groups[idx] != nil {
					_ = groups[idx].Close()
				}

				group, err := writer.OpenGroup(GroupSequence(idx))
				if err == nil {
					groups[idx] = group
				}
			}

			b.StopTimer()
			_ = writer.Close()
		})
	}
}

// BenchmarkTrackWriter_MemoryAllocation benchmarks memory allocation for track writers
func BenchmarkTrackWriter_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mockStream := &MockQUICStream{}
		mockStream.On("Context").Return(context.Background())
		mockStream.On("StreamID").Return(quic.StreamID(1))
		mockStream.On("Read", mock.Anything).Return(0, io.EOF)
		mockStream.On("Write", mock.Anything).Return(0, nil)
		mockStream.On("Close").Return(nil)

		substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

		openUniStreamFunc := func() (quic.SendStream, error) {
			mockSendStream := &MockQUICSendStream{}
			mockSendStream.On("Context").Return(context.Background())
			mockSendStream.On("CancelWrite", mock.Anything).Return()
			mockSendStream.On("StreamID").Return(quic.StreamID(i))
			mockSendStream.On("Close").Return(nil)
			mockSendStream.On("Write", mock.Anything).Return(0, nil)
			return mockSendStream, nil
		}

		writer := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, func() {})

		// Open and close a group
		group, _ := writer.OpenGroup(GroupSequence(1))
		if group != nil {
			_ = group.Close()
		}

		_ = writer.Close()
	}
}

// BenchmarkTrackReader_MemoryAllocation benchmarks memory allocation for track readers
func BenchmarkTrackReader_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mockStream := &MockQUICStream{}
		mockStream.On("Context").Return(context.Background())
		substr := newSendSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{}, Info{})
		reader := newTrackReader("broadcastPath", "trackName", substr, func() {})

		// Enqueue and dequeue a group
		mockRecvStream := &MockQUICReceiveStream{}
		mockRecvStream.On("Context").Return(context.Background())
		reader.enqueueGroup(GroupSequence(1), mockRecvStream)

		group := reader.dequeueGroup()
		if group != nil {
			reader.removeGroup(group)
		}

		_ = reader.Close()
	}
}

// BenchmarkTrackWriter_CloseWithActiveGroups benchmarks closing with many active groups
func BenchmarkTrackWriter_CloseWithActiveGroups(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("groups-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				mockStream := &MockQUICStream{}
				mockStream.On("Context").Return(context.Background())
				mockStream.On("StreamID").Return(quic.StreamID(1))
				mockStream.On("Read", mock.Anything).Return(0, io.EOF)
				mockStream.On("Write", mock.Anything).Return(0, nil)
				mockStream.On("Close").Return(nil)

				substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

				streamIdx := 0
				openUniStreamFunc := func() (quic.SendStream, error) {
					mockSendStream := &MockQUICSendStream{}
					mockSendStream.WriteFunc = func(p []byte) (int, error) {
						return len(p), nil
					}
					mockSendStream.On("Context").Return(context.Background())
					mockSendStream.On("CancelWrite", mock.Anything).Return()
					mockSendStream.On("StreamID").Return(quic.StreamID(streamIdx))
					streamIdx++
					mockSendStream.On("Close").Return(nil)
					return mockSendStream, nil
				}

				writer := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, func() {})

				// Create many active groups
				for j := 0; j < size; j++ {
					_, _ = writer.OpenGroup(GroupSequence(j))
				}

				// Close all at once
				_ = writer.Close()
			}
		})
	}
}

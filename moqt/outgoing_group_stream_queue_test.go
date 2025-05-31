package moqt

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

func TestNewOutgoingGroupStreamQueue(t *testing.T) {
	q := newOutgoingGroupStreamQueue()

	if q == nil {
		t.Fatal("newOutgoingGroupStreamQueue returned nil")
	}

	if q.queue == nil {
		t.Error("queue should not be nil")
	}

	if len(q.queue) != 0 {
		t.Errorf("initial queue length = %v, want 0", len(q.queue))
	}
}

// Helper function to create a real sendGroupStream for testing
func createTestSendGroupStream(t *testing.T) *sendGroupStream {
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		nil, // no logger needed for this test
		nil,
	)
	trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))
	groupCtx := newGroupContext(trackCtx, GroupSequence(1))

	mockStream := &MockQUICSendStream{}
	return newSendGroupStream(mockStream, groupCtx)
}

func TestOutgoingGroupStreamQueueOperations(t *testing.T) {
	q := newOutgoingGroupStreamQueue()

	// Create test streams
	stream1 := createTestSendGroupStream(t)
	stream2 := createTestSendGroupStream(t)

	// Test add
	q.add(stream1)
	q.add(stream2)

	// Verify streams were added
	q.mu.Lock()
	if len(q.queue) != 2 {
		t.Errorf("queue length after adding 2 streams = %v, want 2", len(q.queue))
	}
	_, existsStream1 := q.queue[stream1]
	_, existsStream2 := q.queue[stream2]
	if !existsStream1 || !existsStream2 {
		t.Error("streams not properly added to queue")
	}
	q.mu.Unlock()

	// Test remove
	q.remove(stream1)

	// Verify stream was removed
	q.mu.Lock()
	if len(q.queue) != 1 {
		t.Errorf("queue length after removing 1 stream = %v, want 1", len(q.queue))
	}
	_, existsStream1 = q.queue[stream1]
	if existsStream1 {
		t.Error("stream1 should have been removed")
	}
	q.mu.Unlock()

	// Test clear with error
	testErr := errors.New("test error")
	q.clear(testErr)

	// Verify queue is cleared
	q.mu.Lock()
	if len(q.queue) != 0 {
		t.Errorf("queue length after clear = %v, want 0", len(q.queue))
	}
	q.mu.Unlock()
}

func TestOutgoingGroupStreamQueueConcurrentAccess(t *testing.T) {
	q := newOutgoingGroupStreamQueue()
	var wg sync.WaitGroup

	// Create a number of streams
	numStreams := 100
	streams := make([]*sendGroupStream, numStreams)
	for i := 0; i < numStreams; i++ {
		streams[i] = createTestSendGroupStream(t)
	}

	// Add streams concurrently
	wg.Add(numStreams)
	for i := 0; i < numStreams; i++ {
		go func(i int) {
			defer wg.Done()
			q.add(streams[i])
		}(i)
	}
	wg.Wait()

	// Verify all streams were added
	q.mu.Lock()
	if len(q.queue) != numStreams {
		t.Errorf("queue length = %v, want %v", len(q.queue), numStreams)
	}
	q.mu.Unlock()

	// Remove half the streams concurrently
	wg.Add(numStreams / 2)
	for i := 0; i < numStreams/2; i++ {
		go func(i int) {
			defer wg.Done()
			q.remove(streams[i])
		}(i)
	}
	wg.Wait()

	// Verify streams were removed
	q.mu.Lock()
	if len(q.queue) != numStreams/2 {
		t.Errorf("queue length = %v, want %v", len(q.queue), numStreams/2)
	}
	q.mu.Unlock()

	// Test clear concurrently with adds
	wg.Add(2)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		q.clear(errors.New("test error"))
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			q.add(createTestSendGroupStream(t))
			time.Sleep(1 * time.Millisecond)
		}
	}()
	wg.Wait()

	// Final check - queue should be empty or have a few streams
	q.mu.Lock()
	t.Logf("Final queue length: %d", len(q.queue))
	q.mu.Unlock()
}

func TestOutgoingGroupStreamQueueClearWithNilError(t *testing.T) {
	q := newOutgoingGroupStreamQueue()

	// Add some streams
	for i := 0; i < 5; i++ {
		stream := createTestSendGroupStream(t)
		q.add(stream)
	}

	// Clear with nil error
	q.clear(nil)

	// Verify queue is cleared
	q.mu.Lock()
	if len(q.queue) != 0 {
		t.Errorf("queue length after clear with nil error = %v, want 0", len(q.queue))
	}
	q.mu.Unlock()
}

func TestOutgoingGroupStreamQueueSafeWithNilStream(t *testing.T) {
	q := newOutgoingGroupStreamQueue()

	// These should not panic
	q.add(nil)
	q.remove(nil)

	// Verify queue is still usable
	stream := createTestSendGroupStream(t)
	q.add(stream)

	q.mu.Lock()
	if len(q.queue) != 1 {
		t.Errorf("queue length = %v, want 1", len(q.queue))
	}
	q.mu.Unlock()
}

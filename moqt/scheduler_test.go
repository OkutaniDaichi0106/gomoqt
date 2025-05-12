// Package moqt_test provides tests for the moqt package.
package moqt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduler_NewScheduler tests the creation of a new scheduler
func TestScheduler_NewScheduler(t *testing.T) {
	s := newScheduler()
	assert.Equal(t, 0, s.Len(), "A new scheduler should have length 0")
}

// TestScheduler_Enqueue tests the Enqueue operation
func TestScheduler_Enqueue(t *testing.T) {
	s := newScheduler()

	// Add an item
	s.Enqueue(1, 10)
	assert.Equal(t, 1, s.Len(), "Scheduler should have length 1 after enqueueing one item")

	// Add another item
	s.Enqueue(2, 5)
	assert.Equal(t, 2, s.Len(), "Scheduler should have length 2 after enqueueing two items")
}

// TestScheduler_Dequeue tests the Dequeue operation
func TestScheduler_Dequeue(t *testing.T) {
	s := newScheduler()

	// Add items with different priorities
	s.Enqueue(1, 30)
	s.Enqueue(2, 10)
	s.Enqueue(3, 20)

	// Check dequeue order (lowest priority value first)
	ctx := context.Background()

	// First dequeue should give ID 2 (priority 10)
	id, err := s.Dequeue(ctx)
	require.NoError(t, err)
	assert.Equal(t, SubscribeID(2), id, "First dequeue should return ID 2 with priority 10")

	// Second dequeue should give ID 3 (priority 20)
	id, err = s.Dequeue(ctx)
	require.NoError(t, err)
	assert.Equal(t, SubscribeID(3), id, "Second dequeue should return ID 3 with priority 20")

	// Third dequeue should give ID 1 (priority 30)
	id, err = s.Dequeue(ctx)
	require.NoError(t, err)
	assert.Equal(t, SubscribeID(1), id, "Third dequeue should return ID 1 with priority 30")

	// Length should be 0 now
	assert.Equal(t, 0, s.Len(), "Scheduler should be empty after dequeueing all items")
}

// TestScheduler_DequeueEmptyWithTimeout tests dequeuing from an empty scheduler with a timeout
func TestScheduler_DequeueEmptyWithTimeout(t *testing.T) {
	s := newScheduler()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Try to dequeue from empty scheduler
	_, err := s.Dequeue(ctx)
	assert.Error(t, err, "Dequeue should return error when context times out")
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Error should be context.DeadlineExceeded")
}

// TestScheduler_ConcurrentEnqueueDequeue tests concurrent enqueue and dequeue operations
func TestScheduler_ConcurrentEnqueueDequeue(t *testing.T) {
	s := newScheduler()

	// Create a background context
	ctx := context.Background()

	// Start a goroutine to dequeue
	done := make(chan SubscribeID)
	go func() {
		id, err := s.Dequeue(ctx)
		require.NoError(t, err)
		done <- id
	}()

	// Give the goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Enqueue an item, which should unblock the dequeue operation
	s.Enqueue(42, 1)

	// Wait for the result with a timeout to avoid hanging tests
	select {
	case id := <-done:
		assert.Equal(t, SubscribeID(42), id, "Dequeued ID should be 42")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for dequeue operation to complete")
	}
}

// TestScheduler_MultipleConcurrentDequeues tests multiple concurrent dequeue operations
func TestScheduler_MultipleConcurrentDequeues(t *testing.T) {
	s := newScheduler()
	ctx := context.Background()

	// Start multiple goroutines to dequeue
	const numDequeuers = 3
	done := make(chan SubscribeID, numDequeuers)

	for i := 0; i < numDequeuers; i++ {
		go func() {
			id, err := s.Dequeue(ctx)
			require.NoError(t, err)
			done <- id
		}()
	}

	// Give goroutines time to start
	time.Sleep(10 * time.Millisecond)

	// Enqueue items one by one, checking that each unblocks one dequeuer
	expectedIDs := []SubscribeID{5, 8, 13}
	receivedIDs := make(map[SubscribeID]bool)

	for _, id := range expectedIDs {
		s.Enqueue(id, TrackPriority(id))

		select {
		case receivedID := <-done:
			receivedIDs[receivedID] = true
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for dequeue operation to complete")
		}
	}

	// Verify all expected IDs were received
	for _, id := range expectedIDs {
		assert.True(t, receivedIDs[id], "Expected ID %d to be dequeued", id)
	}
}

// TestScheduler_DequeueContextCancellation tests that dequeue responds to context cancellation
func TestScheduler_DequeueContextCancellation(t *testing.T) {
	s := newScheduler()

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to dequeue
	errCh := make(chan error, 1)
	go func() {
		_, err := s.Dequeue(ctx)
		errCh <- err
	}()

	// Give the goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for the result with a timeout
	select {
	case err := <-errCh:
		assert.Error(t, err, "Dequeue should return error when context is canceled")
		assert.ErrorIs(t, err, context.Canceled, "Error should be context.Canceled")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for dequeue operation to complete")
	}
}

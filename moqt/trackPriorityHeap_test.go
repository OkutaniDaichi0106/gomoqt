// Package moqt_test provides tests for the moqt package.
package moqt

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTrackPriorityHeap_Empty tests the behavior of an empty heap
func TestTrackPriorityHeap_Empty(t *testing.T) {
	// Create a new heap using the exported constructor function in the test package
	h := newTrackPriorityHeap()
	assert.Equal(t, 0, h.Len(), "A new heap should have length 0")
}

// TestTrackPriorityHeap_Push tests pushing elements into the heap
func TestTrackPriorityHeap_Push(t *testing.T) {
	h := newTrackPriorityHeap()

	// Push one element
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{
		id:       1,
		priority: 10,
	})
	assert.Equal(t, 1, h.Len(), "Heap length should be 1 after pushing an element")

	// Push another element with higher priority (lower priority value)
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{
		id:       2,
		priority: 5,
	})
	assert.Equal(t, 2, h.Len(), "Heap length should be 2 after pushing a second element")
}

// TestTrackPriorityHeap_Pop tests popping elements from the heap
func TestTrackPriorityHeap_Pop(t *testing.T) {
	h := newTrackPriorityHeap()

	// Add elements with different priorities
	entries := []struct {
		id       SubscribeID
		priority TrackPriority
	}{
		{id: 1, priority: 30},
		{id: 2, priority: 10},
		{id: 3, priority: 20},
		{id: 4, priority: 5},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	// We should get the entries in order of ascending priority value
	// (Priority 5 is the highest priority, 30 is the lowest)
	expectedOrder := []SubscribeID{4, 2, 3, 1}

	for i, expectedID := range expectedOrder {
		entry := heap.Pop(h).(struct {
			id       SubscribeID
			priority TrackPriority
		})
		assert.Equal(t, expectedID, entry.id,
			"Pop #%d should return ID %d, but got %d", i+1, expectedID, entry.id)
	}

	assert.Equal(t, 0, h.Len(), "Heap should be empty after popping all elements")
}

// TestTrackPriorityHeap_Order tests that the heap maintains proper order
func TestTrackPriorityHeap_Order(t *testing.T) {
	h := newTrackPriorityHeap()

	// Add elements in arbitrary order
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 10, priority: 100})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 5, priority: 50})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 3, priority: 30})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 7, priority: 70})

	// Pop them to verify order
	assert.Equal(t, SubscribeID(3), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "First pop should return ID 3")
	assert.Equal(t, SubscribeID(5), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "Second pop should return ID 5")
	assert.Equal(t, SubscribeID(7), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "Third pop should return ID 7")
	assert.Equal(t, SubscribeID(10), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "Fourth pop should return ID 10")
}

// TestTrackPriorityHeap_SamePriority tests handling of elements with the same priority
func TestTrackPriorityHeap_SamePriority(t *testing.T) {
	h := newTrackPriorityHeap()

	// Add elements with the same priority
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 1, priority: 10})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 2, priority: 10})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 3, priority: 10})

	// When priorities are equal, the order is undefined
	// We just want to make sure they all come out
	ids := make(map[SubscribeID]bool)

	for i := 0; i < 3; i++ {
		entry := heap.Pop(h).(struct {
			id       SubscribeID
			priority TrackPriority
		})
		ids[entry.id] = true
	}

	// Verify all IDs were present
	assert.True(t, ids[1], "ID 1 should be in the heap")
	assert.True(t, ids[2], "ID 2 should be in the heap")
	assert.True(t, ids[3], "ID 3 should be in the heap")
}

// TestTrackPriorityHeap_MixedOperations tests a mix of push and pop operations
func TestTrackPriorityHeap_MixedOperations(t *testing.T) {
	h := newTrackPriorityHeap()

	// Push initial elements
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 1, priority: 30})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 2, priority: 20})

	// Pop highest priority
	entry := heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	})
	assert.Equal(t, SubscribeID(2), entry.id, "First pop should return ID 2")

	// Push more elements
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 3, priority: 10})
	heap.Push(h, struct {
		id       SubscribeID
		priority TrackPriority
	}{id: 4, priority: 40})

	// Verify order of remaining elements
	assert.Equal(t, SubscribeID(3), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "Second pop should return ID 3")
	assert.Equal(t, SubscribeID(1), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "Third pop should return ID 1")
	assert.Equal(t, SubscribeID(4), heap.Pop(h).(struct {
		id       SubscribeID
		priority TrackPriority
	}).id, "Fourth pop should return ID 4")
}

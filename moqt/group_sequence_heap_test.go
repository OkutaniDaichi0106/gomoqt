package moqt

import (
	"container/heap"
	"testing"
)

func TestNewGroupSequenceHeap(t *testing.T) {
	tests := map[string]struct {
		order GroupOrder
	}{
		"default order": {
			order: GroupOrderDefault,
		},
		"ascending order": {
			order: GroupOrderAscending,
		},
		"descending order": {
			order: GroupOrderDescending,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			h := newGroupSequenceHeap(tt.order)

			if h == nil {
				t.Fatal("newGroupSequenceHeap returned nil")
			}

			if h.order != tt.order {
				t.Errorf("order = %v, want %v", h.order, tt.order)
			}

			if h.Len() != 0 {
				t.Errorf("initial length = %v, want 0", h.Len())
			}

			if h.queue == nil {
				t.Error("queue should not be nil")
			}
		})
	}
}

func TestGroupSequenceHeap_PushPop(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test pushing elements
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 0},
		{seq: GroupSequence(2), index: 1},
		{seq: GroupSequence(8), index: 2},
		{seq: GroupSequence(1), index: 3},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	if h.Len() != len(entries) {
		t.Errorf("length after push = %v, want %v", h.Len(), len(entries))
	}

	// Test popping elements (should come out in ascending order)
	expectedOrder := []GroupSequence{1, 2, 5, 8}

	for i, expected := range expectedOrder {
		if h.Len() == 0 {
			t.Fatalf("heap empty at iteration %d", i)
		}

		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatalf("invalid item type at iteration %d", i)
		}

		if entry.seq != expected {
			t.Errorf("iteration %d: seq = %v, want %v", i, entry.seq, expected)
		}
	}

	if h.Len() != 0 {
		t.Errorf("final length = %v, want 0", h.Len())
	}
}

func TestGroupSequenceHeap_Descending(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderDescending)

	// Test pushing elements
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 0},
		{seq: GroupSequence(2), index: 1},
		{seq: GroupSequence(8), index: 2},
		{seq: GroupSequence(1), index: 3},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	// Test popping elements (should come out in descending order)
	expectedOrder := []GroupSequence{8, 5, 2, 1}
	for i, expected := range expectedOrder {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatalf("invalid item type at iteration %d", i)
		}

		if entry.seq != expected {
			t.Errorf("iteration %d: seq = %v, want %v", i, entry.seq, expected)
		}
	}

	if h.Len() != 0 {
		t.Errorf("final length = %v, want 0", h.Len())
	}
}

func TestGroupSequenceHeap_Default(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderDefault)

	// Test pushing elements
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 0},
		{seq: GroupSequence(2), index: 1},
		{seq: GroupSequence(8), index: 2},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	// With default order, Less always returns true, so order is not guaranteed
	// Just test that we can pop all elements
	for i := 0; i < len(entries); i++ {
		item := heap.Pop(h)
		_, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatalf("invalid item type at iteration %d", i)
		}
	}
}

func TestGroupSequenceHeap_Less(t *testing.T) {
	tests := map[string]struct {
		order GroupOrder
		seqi  GroupSequence
		seqj  GroupSequence
		want  bool
	}{
		"ascending: i < j": {
			order: GroupOrderAscending,
			seqi:  GroupSequence(1),
			seqj:  GroupSequence(2),
			want:  true,
		},
		"ascending: i > j": {
			order: GroupOrderAscending,
			seqi:  GroupSequence(2),
			seqj:  GroupSequence(1),
			want:  false,
		},
		"descending: i > j": {
			order: GroupOrderDescending,
			seqi:  GroupSequence(2),
			seqj:  GroupSequence(1),
			want:  true,
		},
		"descending: i < j": {
			order: GroupOrderDescending,
			seqi:  GroupSequence(1),
			seqj:  GroupSequence(2),
			want:  false,
		},
		"default order": {
			order: GroupOrderDefault,
			seqi:  GroupSequence(1),
			seqj:  GroupSequence(2),
			want:  true,
		},
		"invalid order": {
			order: GroupOrder(99),
			seqi:  GroupSequence(1),
			seqj:  GroupSequence(2),
			want:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			h := newGroupSequenceHeap(tt.order)

			// Add two elements to test Less
			h.queue = append(h.queue, struct {
				seq   GroupSequence
				index int
			}{seq: tt.seqi, index: 0})

			h.queue = append(h.queue, struct {
				seq   GroupSequence
				index int
			}{seq: tt.seqj, index: 1})

			result := h.Less(0, 1)
			if result != tt.want {
				t.Errorf("Less() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGroupSequenceHeap_Swap(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Add two elements
	entry1 := struct {
		seq   GroupSequence
		index int
	}{seq: GroupSequence(1), index: 10}

	entry2 := struct {
		seq   GroupSequence
		index int
	}{seq: GroupSequence(2), index: 20}

	h.queue = append(h.queue, entry1, entry2)

	// Swap them
	h.Swap(0, 1)

	// Check that they were swapped
	if h.queue[0] != entry2 {
		t.Errorf("queue[0] = %v, want %v", h.queue[0], entry2)
	}

	if h.queue[1] != entry1 {
		t.Errorf("queue[1] = %v, want %v", h.queue[1], entry1)
	}
}

func TestGroupSequenceHeap_ResetOrder(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	if h.order != GroupOrderAscending {
		t.Errorf("initial order = %v, want %v", h.order, GroupOrderAscending)
	}

	h.resetOrder(GroupOrderDescending)

	if h.order != GroupOrderDescending {
		t.Errorf("order after reset = %v, want %v", h.order, GroupOrderDescending)
	}
}

func TestGroupSequenceHeap_PushInvalidType(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)
	initialLen := h.Len()

	// Try to push invalid type
	h.Push("invalid")

	// Length should not change
	if h.Len() != initialLen {
		t.Errorf("length after invalid push = %v, want %v", h.Len(), initialLen)
	}
}

func TestGroupSequenceHeap_PopEmpty(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test popping from empty heap
	if h.Len() != 0 {
		t.Errorf("initial length = %v, want 0", h.Len())
	}

	// Pop from empty heap should return nil or empty struct
	item := h.Pop()
	expected := struct {
		seq   GroupSequence
		index int
	}{}

	if item != expected {
		t.Errorf("Pop() from empty heap = %v, want %v", item, expected)
	}
}

func TestGroupSequenceHeap_DuplicateSequences(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test with duplicate sequence numbers
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 0},
		{seq: GroupSequence(2), index: 1},
		{seq: GroupSequence(5), index: 2},
		{seq: GroupSequence(2), index: 3},
		{seq: GroupSequence(5), index: 4},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	if h.Len() != len(entries) {
		t.Errorf("length after push = %v, want %v", h.Len(), len(entries))
	}

	// Pop all elements and verify they come out in correct order
	var results []GroupSequence
	for h.Len() > 0 {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatal("invalid item type")
		}
		results = append(results, entry.seq)
	}

	// Should be sorted in ascending order
	for i := 1; i < len(results); i++ {
		if results[i-1] > results[i] {
			t.Errorf("result not in ascending order: %v > %v at position %d", results[i-1], results[i], i)
		}
	}
}

func TestGroupSequenceHeap_IndexPreservation(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test that index values are preserved correctly
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(3), index: 100},
		{seq: GroupSequence(1), index: 200},
		{seq: GroupSequence(2), index: 300},
	}

	indexMap := make(map[GroupSequence]int)
	for _, entry := range entries {
		heap.Push(h, entry)
		indexMap[entry.seq] = entry.index
	}

	// Pop and verify that indices are preserved
	for h.Len() > 0 {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatal("invalid item type")
		}

		expectedIndex := indexMap[entry.seq]
		if entry.index != expectedIndex {
			t.Errorf("seq %v: index = %v, want %v", entry.seq, entry.index, expectedIndex)
		}
	}
}

func TestGroupSequenceHeap_BoundaryValues(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test with boundary values
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(0), index: 0},              // minimum value
		{seq: GroupSequence(^uint64(0)), index: 1},     // maximum value
		{seq: GroupSequence(1), index: 2},              // near minimum
		{seq: GroupSequence(^uint64(0) - 1), index: 3}, // near maximum
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	// Pop and verify ascending order
	var prev GroupSequence
	first := true
	for h.Len() > 0 {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatal("invalid item type")
		}

		if !first && entry.seq < prev {
			t.Errorf("not in ascending order: %v < %v", entry.seq, prev)
		}
		prev = entry.seq
		first = false
	}
}

func TestGroupSequenceHeap_LargeDataset(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test with larger dataset
	const size = 1000
	for i := 0; i < size; i++ {
		entry := struct {
			seq   GroupSequence
			index int
		}{
			seq:   GroupSequence(size - i), // reverse order to test heap sorting
			index: i,
		}
		heap.Push(h, entry)
	}

	if h.Len() != size {
		t.Errorf("length after push = %v, want %v", h.Len(), size)
	}

	// Pop all and verify ascending order
	var prev GroupSequence
	first := true
	count := 0
	for h.Len() > 0 {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatal("invalid item type")
		}

		if !first && entry.seq < prev {
			t.Errorf("not in ascending order: %v < %v", entry.seq, prev)
		}
		prev = entry.seq
		first = false
		count++
	}

	if count != size {
		t.Errorf("popped %v items, want %v", count, size)
	}
}

func TestGroupSequenceHeap_HeapProperty(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Add some elements
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 0},
		{seq: GroupSequence(2), index: 1},
		{seq: GroupSequence(8), index: 2},
		{seq: GroupSequence(1), index: 3},
		{seq: GroupSequence(9), index: 4},
		{seq: GroupSequence(3), index: 5},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
		// Verify heap property after each push
		if !isValidHeap(h) {
			t.Errorf("heap property violated after pushing %v", entry.seq)
		}
	}

	// Pop elements and verify heap property is maintained
	for h.Len() > 0 {
		heap.Pop(h)
		if h.Len() > 0 && !isValidHeap(h) {
			t.Error("heap property violated after pop")
		}
	}
}

// Helper function to verify heap property
func isValidHeap(h *groupSequenceHeap) bool {
	for i := 0; i < h.Len(); i++ {
		left := 2*i + 1
		right := 2*i + 2

		if left < h.Len() && !h.Less(i, left) {
			return false
		}
		if right < h.Len() && !h.Less(i, right) {
			return false
		}
	}
	return true
}

func TestGroupSequenceHeap_OrderChange(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Add elements
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 0},
		{seq: GroupSequence(2), index: 1},
		{seq: GroupSequence(8), index: 2},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	// Change order and re-initialize
	h.resetOrder(GroupOrderDescending)
	heap.Init(h)

	// Pop elements - should now be in descending order
	expectedOrder := []GroupSequence{8, 5, 2}
	for i, expected := range expectedOrder {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatalf("invalid item type at iteration %d", i)
		}

		if entry.seq != expected {
			t.Errorf("iteration %d: seq = %v, want %v", i, entry.seq, expected)
		}
	}
}

func TestGroupSequenceHeap_EqualSequences(t *testing.T) {
	h := newGroupSequenceHeap(GroupOrderAscending)

	// Test behavior with equal sequences
	entries := []struct {
		seq   GroupSequence
		index int
	}{
		{seq: GroupSequence(5), index: 1},
		{seq: GroupSequence(5), index: 2},
		{seq: GroupSequence(5), index: 3},
	}

	for _, entry := range entries {
		heap.Push(h, entry)
	}

	// All should have same sequence number
	seenIndices := make(map[int]bool)
	for h.Len() > 0 {
		item := heap.Pop(h)
		entry, ok := item.(struct {
			seq   GroupSequence
			index int
		})
		if !ok {
			t.Fatal("invalid item type")
		}

		if entry.seq != GroupSequence(5) {
			t.Errorf("seq = %v, want 5", entry.seq)
		}

		if seenIndices[entry.index] {
			t.Errorf("duplicate index %v", entry.index)
		}
		seenIndices[entry.index] = true
	}

	// Should have seen all indices
	for _, entry := range entries {
		if !seenIndices[entry.index] {
			t.Errorf("missing index %v", entry.index)
		}
	}
}

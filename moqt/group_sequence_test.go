package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupSequence_String(t *testing.T) {
	tests := map[string]struct {
		seq  GroupSequence
		want string
	}{
		"not specified": {
			seq:  GroupSequenceNotSpecified,
			want: "0",
		},
		"first sequence": {
			seq:  GroupSequenceFirst,
			want: "1",
		},
		"normal sequence": {
			seq:  GroupSequence(42),
			want: "42",
		}, "max sequence": {
			seq:  MaxGroupSequence,
			want: "4611686018427387903",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.seq.String()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGroupSequence_Next(t *testing.T) {
	tests := map[string]struct {
		seq  GroupSequence
		want GroupSequence
	}{
		"from not specified": {
			seq:  GroupSequenceNotSpecified,
			want: GroupSequence(1),
		},
		"from first": {
			seq:  GroupSequenceFirst,
			want: GroupSequence(2),
		},
		"normal increment": {
			seq:  GroupSequence(42),
			want: GroupSequence(43),
		},
		"from max wraps to 1": {
			seq:  MaxGroupSequence,
			want: GroupSequence(1),
		},
		"near max": {
			seq:  MaxGroupSequence - 1,
			want: MaxGroupSequence,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.seq.Next()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGroupSequence_Constants(t *testing.T) {
	tests := map[string]struct {
		seq  GroupSequence
		want GroupSequence
	}{
		"not specified": {seq: GroupSequenceNotSpecified, want: GroupSequence(0)},
		"latest":        {seq: GroupSequenceLatest, want: GroupSequence(0)},
		"largest":       {seq: GroupSequenceLargest, want: GroupSequence(0)},
		"first":         {seq: GroupSequenceFirst, want: GroupSequence(1)},
		"max":           {seq: MaxGroupSequence, want: GroupSequence(0x3FFFFFFFFFFFFFFF)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.seq)
		})
	}
}

func TestGroupSequence_Type(t *testing.T) {
	// Test that GroupSequence is based on uint64
	var seq GroupSequence = 100

	// Test assignment and comparison
	assert.Equal(t, GroupSequence(100), seq)

	// Test arithmetic operations
	seq++
	assert.Equal(t, GroupSequence(101), seq)

	seq--
	assert.Equal(t, GroupSequence(100), seq)
}

func TestGroupSequence_Comparison(t *testing.T) {
	seq1 := GroupSequence(10)
	seq2 := GroupSequence(20)
	seq3 := GroupSequence(10)

	// Test ordering
	assert.True(t, seq1 < seq2)
	assert.False(t, seq2 < seq1)

	// Test equality
	assert.Equal(t, seq1, seq3)
	assert.NotEqual(t, seq1, seq2)
}

func TestGroupSequence_ZeroValue(t *testing.T) { // Test zero value behavior
	var seq GroupSequence
	assert.Equal(t, GroupSequenceNotSpecified, seq)
	assert.Equal(t, "0", seq.String())
	assert.Equal(t, GroupSequence(1), seq.Next())
}

func TestGroupSequence_MaxBoundary(t *testing.T) {
	// Test behavior at max boundary
	maxSeq := MaxGroupSequence
	assert.Equal(t, GroupSequence(0x3FFFFFFFFFFFFFFF), maxSeq)

	// Test next wraps to 1
	nextSeq := maxSeq.Next()
	assert.Equal(t, GroupSequence(1), nextSeq)

	// Test that we can increment beyond uint32 max since it's uint64
	largeSeq := GroupSequence(0x100000000) // Beyond uint32
	assert.Equal(t, GroupSequence(0x100000001), largeSeq.Next())
}

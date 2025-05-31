package moqt_test

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

func TestGroupSequence_String(t *testing.T) {
	tests := []struct {
		name string
		seq  moqt.GroupSequence
		want string
	}{
		{
			name: "not specified",
			seq:  moqt.GroupSequenceNotSpecified,
			want: "0",
		},
		{
			name: "first sequence",
			seq:  moqt.GroupSequenceFirst,
			want: "1",
		},
		{
			name: "normal sequence",
			seq:  moqt.GroupSequence(42),
			want: "42",
		},
		{
			name: "max sequence",
			seq:  moqt.MaxGroupSequence,
			want: "4294967295",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.seq.String()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGroupSequence_Next(t *testing.T) {
	tests := []struct {
		name string
		seq  moqt.GroupSequence
		want moqt.GroupSequence
	}{
		{
			name: "from not specified",
			seq:  moqt.GroupSequenceNotSpecified,
			want: moqt.GroupSequence(1),
		},
		{
			name: "from first",
			seq:  moqt.GroupSequenceFirst,
			want: moqt.GroupSequence(2),
		},
		{
			name: "normal increment",
			seq:  moqt.GroupSequence(42),
			want: moqt.GroupSequence(43),
		},
		{
			name: "from max wraps to 1",
			seq:  moqt.MaxGroupSequence,
			want: moqt.GroupSequence(1),
		},
		{
			name: "near max",
			seq:  moqt.MaxGroupSequence - 1,
			want: moqt.MaxGroupSequence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.seq.Next()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGroupSequence_Constants(t *testing.T) {
	tests := map[string]struct {
		seq  moqt.GroupSequence
		want moqt.GroupSequence
	}{
		"not specified": {seq: moqt.GroupSequenceNotSpecified, want: moqt.GroupSequence(0)},
		"latest":        {seq: moqt.GroupSequenceLatest, want: moqt.GroupSequence(0)},
		"largest":       {seq: moqt.GroupSequenceLargest, want: moqt.GroupSequence(0)},
		"first":         {seq: moqt.GroupSequenceFirst, want: moqt.GroupSequence(1)},
		"max":           {seq: moqt.MaxGroupSequence, want: moqt.GroupSequence(0xFFFFFFFF)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.seq)
		})
	}
}

func TestGroupSequence_Type(t *testing.T) {
	// Test that GroupSequence is based on uint64
	var seq moqt.GroupSequence = 100

	// Test assignment and comparison
	assert.Equal(t, moqt.GroupSequence(100), seq)

	// Test arithmetic operations
	seq++
	assert.Equal(t, moqt.GroupSequence(101), seq)

	seq--
	assert.Equal(t, moqt.GroupSequence(100), seq)
}

func TestGroupSequence_Comparison(t *testing.T) {
	seq1 := moqt.GroupSequence(10)
	seq2 := moqt.GroupSequence(20)
	seq3 := moqt.GroupSequence(10)

	// Test ordering
	assert.True(t, seq1 < seq2)
	assert.False(t, seq2 < seq1)

	// Test equality
	assert.Equal(t, seq1, seq3)
	assert.NotEqual(t, seq1, seq2)
}

func TestGroupSequence_ZeroValue(t *testing.T) { // Test zero value behavior
	var seq moqt.GroupSequence
	assert.Equal(t, moqt.GroupSequenceNotSpecified, seq)
	assert.Equal(t, "0", seq.String())
	assert.Equal(t, moqt.GroupSequence(1), seq.Next())
}

func TestGroupSequence_MaxBoundary(t *testing.T) {
	// Test behavior at max boundary
	maxSeq := moqt.MaxGroupSequence
	assert.Equal(t, moqt.GroupSequence(0xFFFFFFFF), maxSeq)

	// Test next wraps to 1
	nextSeq := maxSeq.Next()
	assert.Equal(t, moqt.GroupSequence(1), nextSeq)

	// Test that we can increment beyond uint32 max since it's uint64
	largeSeq := moqt.GroupSequence(0x100000000) // Beyond uint32
	assert.Equal(t, moqt.GroupSequence(0x100000001), largeSeq.Next())
}

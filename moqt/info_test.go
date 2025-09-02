package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	tests := map[string]struct {
		groupPeriod GroupPeriod
	}{
		"default values": {
			groupPeriod: GroupPeriod(0),
		},
		"high priority": {
			groupPeriod: GroupPeriod(1),
		},
		"low priority": {
			groupPeriod: GroupPeriod(2),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			info := Info{
				GroupPeriod: tt.groupPeriod,
			}

			assert.Equal(t, tt.groupPeriod, info.GroupPeriod)
		})
	}
}

func TestInfoZeroValue(t *testing.T) {
	var info Info

	assert.Equal(t, GroupPeriod(0), info.GroupPeriod)
}

func TestInfoComparison(t *testing.T) {
	info1 := Info{
		GroupPeriod: GroupPeriod(1),
	}

	info2 := Info{
		GroupPeriod: GroupPeriod(1),
	}

	info3 := Info{
		GroupPeriod: GroupPeriod(2),
	}

	assert.Equal(t, info1, info2, "identical Info structs should be equal")
	assert.NotEqual(t, info1, info3, "different Info structs should not be equal")
}

package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	tests := map[string]struct {
		trackPriority TrackPriority
		groupOrder    GroupOrder
	}{
		"default values": {
			trackPriority: TrackPriority(0),
			groupOrder:    GroupOrder(0),
		},
		"high priority": {
			trackPriority: TrackPriority(255),
			groupOrder:    GroupOrder(1),
		},
		"low priority": {
			trackPriority: TrackPriority(1),
			groupOrder:    GroupOrder(2),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			info := Info{
				TrackPriority: tt.trackPriority,
				GroupOrder:    tt.groupOrder,
			}

			assert.Equal(t, tt.trackPriority, info.TrackPriority)
			assert.Equal(t, tt.groupOrder, info.GroupOrder)
		})
	}
}

func TestInfoZeroValue(t *testing.T) {
	var info Info

	assert.Equal(t, TrackPriority(0), info.TrackPriority)
	assert.Equal(t, GroupOrder(0), info.GroupOrder)
}

func TestInfoComparison(t *testing.T) {
	info1 := Info{
		TrackPriority: TrackPriority(10),
		GroupOrder:    GroupOrder(1),
	}

	info2 := Info{
		TrackPriority: TrackPriority(10),
		GroupOrder:    GroupOrder(1),
	}

	info3 := Info{
		TrackPriority: TrackPriority(20),
		GroupOrder:    GroupOrder(2),
	}

	assert.Equal(t, info1, info2, "identical Info structs should be equal")
	assert.NotEqual(t, info1, info3, "different Info structs should not be equal")
}

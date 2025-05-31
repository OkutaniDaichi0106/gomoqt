package moqt

import (
	"testing"
)

func TestInfo(t *testing.T) {
	tests := []struct {
		name          string
		trackPriority TrackPriority
		groupOrder    GroupOrder
	}{
		{
			name:          "default values",
			trackPriority: TrackPriority(0),
			groupOrder:    GroupOrder(0),
		},
		{
			name:          "high priority",
			trackPriority: TrackPriority(255),
			groupOrder:    GroupOrder(1),
		},
		{
			name:          "low priority",
			trackPriority: TrackPriority(1),
			groupOrder:    GroupOrder(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Info{
				TrackPriority: tt.trackPriority,
				GroupOrder:    tt.groupOrder,
			}

			if info.TrackPriority != tt.trackPriority {
				t.Errorf("TrackPriority = %v, want %v", info.TrackPriority, tt.trackPriority)
			}

			if info.GroupOrder != tt.groupOrder {
				t.Errorf("GroupOrder = %v, want %v", info.GroupOrder, tt.groupOrder)
			}
		})
	}
}

func TestInfoZeroValue(t *testing.T) {
	var info Info

	if info.TrackPriority != 0 {
		t.Errorf("zero value TrackPriority = %v, want 0", info.TrackPriority)
	}

	if info.GroupOrder != 0 {
		t.Errorf("zero value GroupOrder = %v, want 0", info.GroupOrder)
	}
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

	if info1 != info2 {
		t.Error("identical Info structs should be equal")
	}

	if info1 == info3 {
		t.Error("different Info structs should not be equal")
	}
}

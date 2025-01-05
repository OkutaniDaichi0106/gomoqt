package moqt

import (
	"time"
)

type Track struct {
	/*
	 * Required
	 */
	TrackPath string

	/*
	 * Optional
	 */
	TrackPriority TrackPriority
	GroupOrder    GroupOrder
	GroupExpires  time.Duration

	// Parameters
	AuthorizationInfo string

	DeliveryTimeout time.Duration //TODO

	AnnounceParameters Parameters

	/*
	 * Internal
	 */
	latestGroupSequence GroupSequence
}

func (t *Track) Info() Info {
	return Info{
		TrackPriority:       t.TrackPriority,
		LatestGroupSequence: t.latestGroupSequence,
		GroupOrder:          t.GroupOrder,
		GroupExpires:        t.GroupExpires,
	}
}

// func NewTracks(ts []Track) *Tracks {
// 	tracks := makeTracks(len(ts))

// 	for _, t := range ts {
// 		err := tracks.Add(t)
// 		if err != nil {
// 			slog.Error("failed to add a track", slog.String("error", err.Error()))
// 			return nil
// 		}
// 	}

// 	return &tracks
// }

// func makeTracks(len int) Tracks {
// 	return Tracks{
// 		trackMap: make(map[string]Track, len),
// 	}
// }

// // type Tracks struct {
// // 	trackMap map[string]Track
// // 	mu       sync.RWMutex
// // }

// func (t *Tracks) Slice() []Track {
// 	t.mu.RLock()
// 	defer t.mu.RUnlock()

// 	tracks := make([]Track, 0, len(t.trackMap))
// 	for _, track := range t.trackMap {
// 		tracks = append(tracks, track)
// 	}

// 	return tracks
// }

// func (t *Tracks) Map() map[string]Track {
// 	t.mu.RLock()
// 	defer t.mu.RUnlock()

// 	return t.trackMap
// }

// func (t *Tracks) Len() int {
// 	t.mu.RLock()
// 	defer t.mu.RUnlock()

// 	return len(t.trackMap)
// }

// func (t *Tracks) Get(trackPath string) (Track, bool) {
// 	t.mu.RLock()
// 	defer t.mu.RUnlock()

// 	track, ok := t.trackMap[trackPath]
// 	return track, ok
// }

// func (t *Tracks) Add(track Track) error {
// 	if t.trackMap == nil {
// 		newTracks := makeTracks(1)
// 		t = &newTracks
// 	}

// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	_, ok := t.trackMap[track.TrackPath]
// 	if ok {
// 		return ErrDuplicatedTrack
// 	}

// 	t.trackMap[track.TrackPath] = track

// 	return nil
// }

// func (t *Tracks) Delete(trackPath string) {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	delete(t.trackMap, trackPath)
// }

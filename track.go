package moqt

import (
	"log/slog"
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

func NewTracks(ts []Track) Tracks {
	tracks := make(Tracks, len(ts))

	for _, t := range ts {
		err := tracks.Add(t.TrackPath, t)
		if err != nil {
			slog.Error("failed to add a track", slog.String("error", err.Error()))
			return nil
		}
	}

	return tracks
}

type Tracks map[string]Track

func (t Tracks) Get(trackPath string) (Track, bool) {
	track, ok := t[trackPath]
	return track, ok
}

func (t Tracks) Add(trackPath string, track Track) error {
	if t == nil {
		t = make(Tracks)
	}

	_, ok := t[trackPath]
	if ok {
		return ErrDuplicatedTrack
	}

	t[trackPath] = track

	return nil
}

func (t Tracks) Delete(trackPath string) {
	delete(t, trackPath)
}

package moqt

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type Interest struct {
	TrackPrefix string
	Parameters  Parameters
}

type SentInterest struct {
	Interest
	/*
	 * Sent announcements
	 * Track Path -> Announcement
	 */
	active Tracks
	stream transport.Stream
	mu     sync.RWMutex
}

func (interest *SentInterest) NextActiveTracks() (*Tracks, error) {
	interest.mu.Lock()
	defer interest.mu.Unlock()

	if interest.active.Len() == 0 {
		interest.active = makeTracks(1)
	}

	// Read announcements
	for {
		ann, err := readAnnouncement(interest.stream)
		if err != nil {
			slog.Error("failed to read an announcement", slog.String("error", err.Error()))
			return nil, err
		}

		// Get the full track path
		trackPath := interest.TrackPrefix + "/" + ann.TrackPathSuffix

		// Update the active tracks
		if ann.status == ACTIVE {
			_, ok := interest.active.Get(trackPath)
			if ok {
				return nil, ErrInternalError
			}

			interest.active.Add(Track{
				TrackPath:          trackPath,
				AuthorizationInfo:  ann.AuthorizationInfo,
				AnnounceParameters: ann.Parameters,
			})
		}

		// Delete the active tracks if the it is ended
		if ann.status == ENDED {
			_, ok := interest.active.Get(trackPath)
			if !ok {
				return nil, ErrInternalError
			}

			interest.active.Delete(trackPath)
		}

		if ann.status == LIVE {
			break
		}
	}

	return &interest.active, nil
}

type ReceivedInterest struct {
	Interest
	/*
	 * Sent announcements
	 * Track Path -> Announcement
	 */
	activeTracks map[string]Track
	stream       transport.Stream
	mu           sync.RWMutex
}

func (interest *ReceivedInterest) Announce(tracks *Tracks) error {
	interest.mu.Lock()
	defer interest.mu.Unlock()

	// Create a new active tracks
	newActives := make(map[string]Track, tracks.Len())

	// Announce active tracks
	for path, track := range tracks.Map() {
		if _, ok := newActives[path]; ok {
			return ErrDuplicatedTrack
		}

		if _, ok := interest.activeTracks[track.TrackPath]; !ok {
			err := interest.announceActiveTrack(track)
			if err != nil {
				slog.Error("failed to announce active track",
					slog.String("path", track.TrackPath),
					slog.String("error", err.Error()))
				return err
			}
		}

		newActives[track.TrackPath] = track
	}

	// Announce ended tracks
	for path, track := range interest.activeTracks {
		if _, ok := newActives[path]; !ok {
			err := interest.announceEndedTrack(track)
			if err != nil {
				slog.Error("failed to announce ended track",
					slog.String("path", path),
					slog.String("error", err.Error()))
				return err
			}
		}
	}

	// Update
	interest.activeTracks = newActives

	//
	interest.announceLive()

	return nil
}

func (interest *ReceivedInterest) announceActiveTrack(track Track) error {
	// Verify if the track path has the track prefix
	if !strings.HasPrefix(track.TrackPath, interest.TrackPrefix) {
		return ErrInternalError
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(track.TrackPath, interest.TrackPrefix+"/")

	// Create an announcement
	ann := Announcement{
		status:            ACTIVE,
		TrackPathSuffix:   suffix,
		AuthorizationInfo: track.AuthorizationInfo,
		Parameters:        track.AnnounceParameters,
	}

	// Add the Authorization Info
	if track.AuthorizationInfo != "" {
		ann.Parameters.Add(AUTHORIZATION_INFO, track.AuthorizationInfo)
	}

	// Write the announcement
	err := writeAnnouncement(interest.stream, ann)
	if err != nil {
		slog.Error("failed to write an announcement")
		return err
	}

	return nil
}

func (interest *ReceivedInterest) announceEndedTrack(track Track) error {
	// Verify if the track path has the track prefix
	if !strings.HasPrefix(track.TrackPath, interest.TrackPrefix) {
		return ErrInternalError
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(track.TrackPath, interest.TrackPrefix+"/")

	//
	ann := Announcement{
		status:            ENDED,
		TrackPathSuffix:   suffix,
		AuthorizationInfo: track.AuthorizationInfo,
		Parameters:        track.AnnounceParameters,
	}

	//
	if track.AuthorizationInfo != "" {
		ann.Parameters.Add(AUTHORIZATION_INFO, track.AuthorizationInfo)
	}

	err := writeAnnouncement(interest.stream, ann)
	if err != nil {
		slog.Error("failed to write an announcement")
		return err
	}

	return nil
}

func (interest *ReceivedInterest) announceLive() error {
	ann := Announcement{
		status: ACTIVE,
	}

	err := writeAnnouncement(interest.stream, ann)
	if err != nil {
		slog.Error("failed to write an announcement")
		return err
	}

	return nil
}

type receivedInterestQueue struct {
	queue []*ReceivedInterest
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receivedInterestQueue) Len() int {
	return len(q.queue)
}

func (q *receivedInterestQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receivedInterestQueue) Enqueue(interest *ReceivedInterest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, interest)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedInterestQueue) Dequeue() *ReceivedInterest {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	interest := q.queue[0]
	q.queue = q.queue[1:]

	return interest
}

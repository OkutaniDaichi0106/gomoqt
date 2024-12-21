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
	active map[string]Track
	stream transport.Stream
	mu     sync.RWMutex
}

type ReceivedInterest struct {
	Interest
	/*
	 * Sent announcements
	 * Track Path -> Announcement
	 */
	active map[string]Track
	stream transport.Stream
	mu     sync.RWMutex
}

func (interest *ReceivedInterest) Announce(tracks []Track) error {
	//
	newActives := make(map[string]Track, len(tracks))

	for _, track := range tracks {
		_, ok := interest.active[track.TrackPath]
		if !ok {
			err := interest.announce(track, ACTIVE)
			if err != nil {
				slog.Error("failed to announce")
				return err
			}
		}

		newActives[track.TrackPath] = track
	}

	for path, track := range interest.active {
		_, ok := newActives[path]
		if !ok {
			err := interest.announce(track, ENDED)
			if err != nil {
				slog.Error("failed to announce")
				return err
			}
		}
	}

	// Update
	interest.active = newActives

	return nil
}

func (interest *ReceivedInterest) announce(track Track, status AnnounceStatus) error {
	interest.mu.Lock()
	defer interest.mu.Unlock()

	// Verify if the announcement has the track prefix
	if !strings.HasPrefix(track.TrackPath, interest.TrackPrefix) {
		return ErrInternalError
	}

	//
	_, ok := interest.active[track.TrackPath]
	if ok {
		return ErrDuplicatedTrackPath
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(track.TrackPath, interest.TrackPrefix+"/")

	ann := Announcement{
		status:            status,
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

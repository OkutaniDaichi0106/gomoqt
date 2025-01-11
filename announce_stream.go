package moqt

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type Interest struct {
	TrackPrefix string
	Parameters  Parameters
}

type ReceiveAnnounceStream interface {
	ReceiveAnnouncements() ([]Announcement, error)
}

var _ ReceiveAnnounceStream = (*receiveAnnounceStream)(nil)

type receiveAnnounceStream struct {
	interest Interest
	stream   transport.Stream
	mu       sync.RWMutex

	annMap map[string]Announcement
	ch     chan struct{}
	// activeCh chan []Announcement

	// endedCh chan []Announcement
}

func (ras *receiveAnnounceStream) ReceiveAnnouncements() ([]Announcement, error) {
	ras.ch <- struct{}{}

	ras.mu.Lock()
	defer ras.mu.Unlock()

	announcements := make([]Announcement, 0, len(ras.annMap))

	for _, ann := range ras.annMap {
		announcements = append(announcements, ann)
	}

	return announcements, nil
}

// func (ras *receiveAnnounceStream) NextActive() ([]Announcement, error) {

// 	ras.mu.Lock()
// 	defer ras.mu.Unlock()

// 	return <-ras.activeCh, nil
// }

// func (ras *receiveAnnounceStream) NextEnded() ([]Announcement, error) {
// 	ras.mu.Lock()
// 	defer ras.mu.Unlock()

// 	return <-ras.endedCh, nil
// }

type SendAnnounceStream interface {
	SendAnnouncement(announcements []Announcement) error
	Interest() Interest
	Close() error
	CloseWithError(error) error
}

var _ SendAnnounceStream = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	interest Interest
	/*
	 * Sent announcements
	 * Track Path -> Announcement
	 */
	activeTracks map[string]Announcement
	stream       transport.Stream
	mu           sync.RWMutex
}

func (sas *sendAnnounceStream) Interest() Interest {
	return sas.interest
}

func (sas *sendAnnounceStream) SendAnnouncement(announcements []Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	// Create a new active tracks
	newActives := make(map[string]Announcement, len(announcements))

	// Announce active tracks
	for _, ann := range announcements {
		if _, ok := newActives[ann.TrackPath]; ok {
			return ErrDuplicatedTrack
		}

		if _, ok := sas.activeTracks[ann.TrackPath]; !ok {
			err := announceActiveTrack(sas, ann)
			if err != nil {
				slog.Error("failed to announce active track",
					slog.String("path", ann.TrackPath),
					slog.String("error", err.Error()))
				return err
			}
		}

		newActives[ann.TrackPath] = ann
	}

	// Announce ended tracks
	for path, track := range sas.activeTracks {
		if _, ok := newActives[path]; !ok {
			err := announceEndedTrack(sas, track)
			if err != nil {
				slog.Error("failed to announce ended track",
					slog.String("path", path),
					slog.String("error", err.Error()))
				return err
			}
		}
	}

	// Update
	sas.activeTracks = newActives

	//
	announceLive(sas)

	return nil
}

func (sas *sendAnnounceStream) Close() error {
	return sas.stream.Close()
}

func (sas *sendAnnounceStream) CloseWithError(err error) error { // TODO
	if err == nil {
		return sas.stream.Close()
	}

	return nil
}

func announceActiveTrack(sas *sendAnnounceStream, ann Announcement) error {
	// Verify if the track path has the track prefix
	if !strings.HasPrefix(ann.TrackPath, sas.interest.TrackPrefix) {
		return ErrInternalError
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(ann.TrackPath, sas.interest.TrackPrefix+"/")

	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		AnnounceStatus:  message.ACTIVE,
		TrackPathSuffix: suffix,
		Parameters:      message.Parameters(ann.AnnounceParameters),
	}

	// Encode the ANNOUNCE message
	err := am.Encode(sas.stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully announced", slog.Any("announcement", ann))

	return nil
}

func announceEndedTrack(sas *sendAnnounceStream, ann Announcement) error {
	// Verify if the track path has the track prefix
	if !strings.HasPrefix(ann.TrackPath, sas.interest.TrackPrefix) {
		return ErrInternalError
	}

	// Get a suffix part of the Track Path
	suffix := strings.TrimPrefix(ann.TrackPath, sas.interest.TrackPrefix+"/")

	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		AnnounceStatus:  message.ENDED,
		TrackPathSuffix: suffix,
		Parameters:      message.Parameters(ann.AnnounceParameters),
	}

	// Encode the ANNOUNCE message
	err := am.Encode(sas.stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully announced", slog.Any("announcement", ann))

	return nil
}

func announceLive(sas *sendAnnounceStream) error {
	// Initialize an ANNOUNCE message
	am := message.AnnounceMessage{
		AnnounceStatus: message.LIVE,
	}

	// Encode the ANNOUNCE message
	err := am.Encode(sas.stream)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message.", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Successfully announced")

	return nil
}

func newReceivedInterestQueue() *receivedInterestQueue {
	return &receivedInterestQueue{
		queue: make([]*sendAnnounceStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receivedInterestQueue struct {
	queue []*sendAnnounceStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receivedInterestQueue) Len() int {
	return len(q.queue)
}

func (q *receivedInterestQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receivedInterestQueue) Enqueue(interest *sendAnnounceStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, interest)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedInterestQueue) Dequeue() *sendAnnounceStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	interest := q.queue[0]
	q.queue = q.queue[1:]

	return interest
}

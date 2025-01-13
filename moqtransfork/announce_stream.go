package moqtransfork

import (
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

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
	annMap map[string]Announcement
	stream transport.Stream
	mu     sync.RWMutex
}

func (sas *sendAnnounceStream) Interest() Interest {
	return sas.interest
}

func (sas *sendAnnounceStream) SendAnnouncement(announcements []Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	// Announce active tracks
	for _, ann := range announcements {
		oldAnn, ok := sas.annMap[ann.TrackPath]
		if ok && oldAnn.AnnounceStatus == ann.AnnounceStatus {
			slog.Debug("duplicate announcement status")
			return ErrProtocolViolation
		}

		if !ok && ann.AnnounceStatus == ENDED {
			slog.Debug("ended track is not announced")
			return ErrProtocolViolation
		}

		err := writeAnnouncement(sas.stream, sas.interest.TrackPrefix, ann)
		if err != nil {
			slog.Error("failed to announce",
				slog.String("path", ann.TrackPath),
				slog.String("error", err.Error()))
			return err
		}

		sas.annMap[ann.TrackPath] = ann
	}

	// Announce live
	liveAnn := Announcement{
		AnnounceStatus: LIVE,
	}
	err := writeAnnouncement(sas.stream, sas.interest.TrackPrefix, liveAnn)
	if err != nil {
		slog.Error("failed to announce live")
		return err
	}

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

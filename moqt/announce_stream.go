package moqt

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type ReceiveAnnounceStream interface {
	ReceiveAnnouncements() ([]Announcement, error)
}

var _ ReceiveAnnounceStream = (*receiveAnnounceStream)(nil)

type receiveAnnounceStream struct {
	interest AnnounceConfig
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
	AnnounceConfig() AnnounceConfig
	Close() error
	CloseWithError(error) error
}

var _ SendAnnounceStream = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	annConfig AnnounceConfig
	/*
	 * Sent announcements
	 * Track Path -> Announcement
	 */
	annMap map[string]Announcement
	stream transport.Stream
	mu     sync.RWMutex
}

func (sas *sendAnnounceStream) AnnounceConfig() AnnounceConfig {
	return sas.annConfig
}

func (sas *sendAnnounceStream) SendAnnouncement(announcements []Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	// Announce active tracks
	for _, ann := range announcements {
		oldAnn, ok := sas.annMap[strings.Join(ann.TrackPath, "")]
		if ok && oldAnn.AnnounceStatus == ann.AnnounceStatus {
			slog.Debug("duplicate announcement status")
			return ErrProtocolViolation
		}

		if !ok && ann.AnnounceStatus == ENDED {
			slog.Debug("ended track is not announced")
			return ErrProtocolViolation
		}

		err := writeAnnouncement(sas.stream, sas.annConfig.TrackPrefix, ann)
		if err != nil {
			slog.Error("failed to announce",
				slog.Any("path", ann.TrackPath),
				slog.String("error", err.Error()))
			return err
		}

		sas.annMap[strings.Join(ann.TrackPath, "")] = ann
	}

	// Announce live
	liveAnn := Announcement{
		AnnounceStatus: LIVE,
	}
	err := writeAnnouncement(sas.stream, sas.annConfig.TrackPrefix, liveAnn)
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

func newReceivedInterestQueue() *sendAnnounceStreamQueue {
	return &sendAnnounceStreamQueue{
		queue: make([]*sendAnnounceStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type sendAnnounceStreamQueue struct {
	queue []*sendAnnounceStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *sendAnnounceStreamQueue) Len() int {
	return len(q.queue)
}

func (q *sendAnnounceStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *sendAnnounceStreamQueue) Enqueue(interest *sendAnnounceStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, interest)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *sendAnnounceStreamQueue) Dequeue() *sendAnnounceStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	interest := q.queue[0]
	q.queue = q.queue[1:]

	return interest
}

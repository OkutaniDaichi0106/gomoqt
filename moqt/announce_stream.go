package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type ReceiveAnnounceStream interface {
	ReceiveAnnouncements() ([]Announcement, error)
}

var _ ReceiveAnnounceStream = (*receiveAnnounceStream)(nil)

func newReceiveAnnounceStream(stream transport.Stream, config AnnounceConfig) *receiveAnnounceStream {
	return &receiveAnnounceStream{
		config:    config,
		stream:    stream,
		annMap:    make(map[string]Announcement),
		liveAnnCh: make(chan Announcement, 1),
	}
}

type receiveAnnounceStream struct {
	config AnnounceConfig
	stream transport.Stream
	mu     sync.RWMutex

	annMap    map[string]Announcement
	liveAnnCh chan Announcement
	liveAnn   Announcement
	closed    bool
	closeErr  error
}

func (ras *receiveAnnounceStream) LiveAnnouncement() Announcement {
	return ras.liveAnn
}

func (ras *receiveAnnounceStream) ReceiveAnnouncements() ([]Announcement, error) {
	ras.mu.RLock()
	if ras.closed {
		ras.mu.RUnlock()
		return nil, ras.closeErr
	}
	ras.mu.RUnlock()

	//
	select {
	case ann := <-ras.liveAnnCh:
		ras.mu.Lock()
		defer ras.mu.Unlock()

		ras.liveAnn = ann

		// Verify if the Announce Status is LIVE
		if ann.AnnounceStatus != LIVE {
			return nil, ErrProtocolViolation
		}

		announcements := make([]Announcement, 0, len(ras.annMap))
		for _, a := range ras.annMap {
			announcements = append(announcements, a)
		}
		return announcements, nil
	}
}

func (ras *receiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		return nil
	}

	ras.closed = true

	close(ras.liveAnnCh)

	return ras.stream.Close()
}

func (ras *receiveAnnounceStream) isValidateAnnouncement(ann Announcement) bool {
	if ras.annMap == nil {
		ras.annMap = make(map[string]Announcement)
	}

	oldAnn, ok := ras.findAnnouncement(ann.TrackPath)

	if ok && oldAnn.AnnounceStatus == ann.AnnounceStatus {
		slog.Debug("duplicate announcement status")
	}

	if !ok && ann.AnnounceStatus == ENDED {
		slog.Debug("ended track is not announced")
	}

	return true
}

func (ras *receiveAnnounceStream) findAnnouncement(trackPath []string) (Announcement, bool) {
	ann, exists := ras.annMap[TrackPartsString(trackPath)]
	return ann, exists
}

func (ras *receiveAnnounceStream) storeAnnouncement(ann Announcement) {
	ras.annMap[TrackPartsString(ann.TrackPath)] = ann
}

func (ras *receiveAnnounceStream) deleteAnnouncement(trackPath []string) {
	delete(ras.annMap, TrackPartsString(trackPath))
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

	slog.Debug("sending announcements", slog.Any("announcements", announcements))

	if len(announcements) == 0 {
		return errors.New("empty announcements")
	}

	for _, ann := range announcements {
		if !sas.isValidateAnnouncement(ann) {
			return ErrProtocolViolation
		}

		if err := sas.writeAndStoreAnnouncement(ann); err != nil {
			return err
		}
	}

	slog.Debug("sent announcements", slog.Any("announcements", announcements))

	return sas.announceLive()
}

func (sas *sendAnnounceStream) isValidateAnnouncement(ann Announcement) bool {
	if ann.AnnounceStatus != LIVE && ann.AnnounceStatus != ACTIVE && ann.AnnounceStatus != ENDED {
		slog.Debug("invalid announcement status")
		return false
	}

	oldAnn, exists := sas.findAnnouncement(ann.TrackPath)
	if exists && oldAnn.AnnounceStatus == ann.AnnounceStatus {
		slog.Debug("duplicate announcement status")
		return false
	}

	if !exists && ann.AnnounceStatus == ENDED {
		slog.Debug("ended track is not announced")
		return false
	}

	return true
}

func (sas *sendAnnounceStream) writeAndStoreAnnouncement(ann Announcement) error {
	err := writeAnnouncement(sas.stream, sas.annConfig.TrackPrefix, ann)
	if err != nil {
		slog.Error("failed to announce",
			slog.Any("path", ann.TrackPath),
			slog.String("error", err.Error()))
		return err
	}

	sas.storeAnnouncement(ann)
	return nil
}

func (sas *sendAnnounceStream) announceLive() error {
	liveAnn := Announcement{
		AnnounceStatus: LIVE,
		TrackPath:      sas.annConfig.TrackPrefix,
	}

	err := writeAnnouncement(sas.stream, sas.annConfig.TrackPrefix, liveAnn)
	if err != nil {
		slog.Error("failed to announce live")
		return err
	}

	return nil
}

func (sas *sendAnnounceStream) findAnnouncement(trackPath []string) (Announcement, bool) {
	ann, exists := sas.annMap[TrackPartsString(trackPath)]
	return ann, exists
}

func (sas *sendAnnounceStream) storeAnnouncement(ann Announcement) {
	sas.annMap[TrackPartsString(ann.TrackPath)] = ann
}

func (sas *sendAnnounceStream) deleteAnnouncement(trackPath []string) {
	delete(sas.annMap, TrackPartsString(trackPath))
}

func (sas *sendAnnounceStream) Close() error {
	return sas.stream.Close()
}

func (sas *sendAnnounceStream) CloseWithError(err error) error { // TODO
	slog.Debug("closing a send announce stream with an error", slog.String("error", err.Error()))

	if err == nil {
		return sas.stream.Close()
	}

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		annerr, ok := err.(AnnounceError)
		if ok {
			code = transport.StreamErrorCode(annerr.AnnounceErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	sas.stream.CancelRead(code)
	sas.stream.CancelWrite(code)

	slog.Debug("closed a send announce stream with an error", slog.String("error", err.Error()))

	return nil
}

func newSendAnnounceStreamQueue() *sendAnnounceStreamQueue {
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

	if q.Len() <= 0 {
		return nil
	}

	sas := q.queue[0]
	q.queue = q.queue[1:]

	return sas
}

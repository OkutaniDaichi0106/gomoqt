package moqt

import (
	"bytes"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type AnnouncementWriter interface {
	SendAnnouncements(announcements []*Announcement) error
}

// func (w AnnouncementWriter) SendAnnouncements(announcements []*Announcement) error {}

func newSendAnnounceStream(stream quic.Stream, prefix string) *sendAnnounceStream {
	sas := &sendAnnounceStream{
		prefix:   prefix,
		stream:   stream,
		actives:  make(map[string]*Announcement),
		pendings: make(map[string]message.AnnounceMessage),
		sendCh:   make(chan struct{}, 1),
	}

	go func() {
		for range sas.sendCh {
			err := sas.send()
			if err != nil {
				slog.Error("failed to send announcements", "err", err)
			}
		}
	}()

	return sas
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	prefix string

	stream quic.Stream

	mu sync.Mutex

	actives map[string]*Announcement

	pendings  map[string]message.AnnounceMessage
	pendingMu sync.Mutex

	closed   bool
	closeErr error

	sendCh chan struct{}
}

func (sas *sendAnnounceStream) SendAnnouncements(announcements []*Announcement) error {
	sas.pendingMu.Lock()
	defer sas.pendingMu.Unlock()

	// Set active announcement
	for _, ann := range announcements {
		suffix, ok := ann.BroadcastPath().GetSuffix(sas.prefix)
		if !ok {
			continue
		}

		if active, ok := sas.actives[suffix]; ok {
			active.cancel()
			delete(sas.actives, suffix)
		}

		sas.actives[suffix] = ann

		err := sas.set(suffix, true)
		if err != nil {
			return err
		}

		go func(ann *Announcement) {
			<-ann.AwaitEnd()

			err := sas.set(suffix, false)
			if err != nil {
				slog.Error("failed to set an ended announcement",
					"suffix", suffix,
					"error", err,
				)
				return
			}

			delete(sas.actives, suffix)

			sas.sendCh <- struct{}{}
		}(ann)
	}

	return nil
}

func (sas *sendAnnounceStream) set(suffix string, active bool) error {
	sas.pendingMu.Lock()
	defer sas.pendingMu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return errors.New("stream already closed")
	}

	_, ok := sas.pendings[suffix]

	if active {
		sas.pendings[suffix] = message.AnnounceMessage{
			AnnounceStatus: message.ACTIVE,
			TrackSuffix:    suffix,
		}

		return nil
	} else {
		if ok {
			delete(sas.pendings, suffix)
		} else {
			sas.pendings[suffix] = message.AnnounceMessage{
				AnnounceStatus: message.ENDED,
				TrackSuffix:    suffix,
			}
		}
	}

	return nil
}

func (sas *sendAnnounceStream) send() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return errors.New("stream already closed")
	}

	if len(sas.pendings) == 0 {
		return nil
	}

	var len int
	for _, am := range sas.pendings {
		len += am.Len()
	}

	buf := bytes.NewBuffer(make([]byte, 0, len))

	for _, am := range sas.pendings {
		_, err := am.Encode(buf)
		if err != nil {
			return err
		}
	}

	_, err := sas.stream.Write(buf.Bytes())
	if err != nil {
		return err
	}

	sas.pendings = make(map[string]message.AnnounceMessage)

	return nil
}

func (sas *sendAnnounceStream) Close() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return nil
	}

	sas.closed = true

	return nil
}

func (sas *sendAnnounceStream) CloseWithError(err error) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return nil
	}

	sas.closed = true
	sas.closeErr = err

	return nil
}

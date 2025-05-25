package moqt

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type AnnouncementWriter interface {
	SendAnnouncements(announcements []*Announcement) error
	// Close() error
	// CloseWithError(error) error
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

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

type sendAnnounceStream struct {
	stream quic.Stream
	prefix string
	mu     sync.Mutex

	actives map[string]*Announcement

	pendings map[string]message.AnnounceMessage

	closed   bool
	closeErr error

	sendCh chan struct{}
}

func (sas *sendAnnounceStream) SendAnnouncements(announcements []*Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	// Set active announcement
	for _, ann := range announcements {
		if !ann.IsActive() {
			// Ignore inactive announcement
			slog.Warn("Ignore inactive announcement",
				"track_path", ann.BroadcastPath(),
			)
			continue
		}

		suffix, ok := ann.BroadcastPath().GetSuffix(sas.prefix)
		if !ok {
			return fmt.Errorf("failed to get suffix from broadcast path: %s", ann.BroadcastPath())
		}

		if active, ok := sas.actives[suffix]; ok {
			if active == ann {
				continue
			}

			active.End()
			delete(sas.actives, suffix)
		}

		new := ann.Clone()
		sas.actives[suffix] = new

		am := message.AnnounceMessage{
			AnnounceStatus: message.ACTIVE,
			TrackSuffix:    suffix,
		}

		sas.pendings[suffix] = am

		go func(ann *Announcement) {
			<-ann.AwaitEnd()

			sas.mu.Lock()
			defer sas.mu.Unlock()

			am := message.AnnounceMessage{
				AnnounceStatus: message.ENDED,
				TrackSuffix:    suffix,
			}

			if _, ok := sas.pendings[suffix]; ok {
				delete(sas.pendings, suffix)
			} else {
				sas.pendings[suffix] = am
			}

			delete(sas.actives, suffix)

			sas.sendCh <- struct{}{}
		}(new)
	}

	sas.sendCh <- struct{}{}

	return nil
}

func (sas *sendAnnounceStream) set(path BroadcastPath) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return nil
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

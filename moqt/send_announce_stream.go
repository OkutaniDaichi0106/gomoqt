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
	Close() error
	CloseWithError(error) error
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

func newSendAnnounceStream(stream quic.Stream, config AnnounceConfig) *sendAnnounceStream {
	sas := &sendAnnounceStream{
		config:    config,
		stream:    stream,
		announced: make(map[string]message.AnnounceMessage),
		pending:   make(map[string]message.AnnounceMessage),
		sendCh:    make(chan struct{}, 1),
	}

	go func() {
		for range sas.sendCh {
			err := sas.sendAnnouncements()
			if err != nil {
				slog.Error("failed to send announcements", "err", err)
			}
		}
	}()

	return sas
}

type sendAnnounceStream struct {
	stream quic.Stream
	config AnnounceConfig
	mu     sync.RWMutex

	pending map[string]message.AnnounceMessage

	announced map[string]message.AnnounceMessage

	closed   bool
	closeErr error

	sendCh chan struct{}
}

func (sas *sendAnnounceStream) SendAnnouncements(announcements []*Announcement) error {
	var err error
	var path string

	// Set active announcement
	for _, ann := range announcements {
		if !ann.TrackPath().Match(sas.TrackPattern()) {
			// Ignore mismatched announcement
			slog.Warn("Ignore mismatched announcement",
				"track_path", ann.TrackPath(),
				"pattern", sas.TrackPattern(),
			)
			continue
		}

		if !ann.IsActive() {
			// Ignore inactive announcement
			slog.Warn("Ignore inactive announcement",
				"track_path", ann.TrackPath(),
			)
			continue
		}

		path = string(ann.TrackPath())

		err = sas.setActiveAnnouncement(path)
		if err != nil {
			return err
		}
	}

	select {
	case sas.sendCh <- struct{}{}:
	default:
	}

	// Send ended announcements
	for _, ann := range announcements {
		go func(ann *Announcement) {
			ann.AwaitEnd()

			err := sas.setEndedAnnouncement(ann.TrackPath().String())
			if err != nil {
				slog.Error("failed to set ended announcement",
					"track_path", ann.TrackPath(),
					"err", err)
				return
			}

			select {
			case sas.sendCh <- struct{}{}:
			default:
			}
		}(ann)
	}

	return nil
}

func (s *sendAnnounceStream) TrackPattern() string {
	return s.config.TrackPattern
}

func (s *sendAnnounceStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		if s.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", s.closeErr)
		}
		return errors.New("stream already closed")
	}

	s.closed = true

	slog.Debug("closed a send announce stream gracefully",
		slog.Any("stream_id", s.stream.StreamID()),
	)

	return s.stream.Close()
}

func (sas *sendAnnounceStream) CloseWithError(err error) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	var annerr AnnounceError
	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := quic.StreamErrorCode(annerr.AnnounceErrorCode())
	sas.stream.CancelRead(code)
	sas.stream.CancelWrite(code)

	slog.Debug("closed a send announce stream with an error",
		slog.Any("stream_id", sas.stream.StreamID()),
		slog.String("reason", err.Error()),
	)

	return nil
}

func (sas *sendAnnounceStream) setActiveAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	if am, ok := sas.announced[suffix]; ok {
		if am.AnnounceStatus == message.ACTIVE {
			return fmt.Errorf("duplicated announcement: %s", suffix)
		}
	}

	//
	if am, ok := sas.pending[suffix]; ok {
		if am.AnnounceStatus == message.ACTIVE {
			// Skip
			slog.Warn("active announcement already set", "track_suffix", suffix)
			return nil
		}
	}

	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}

	sas.pending[suffix] = am

	slog.Debug("set active announcement", slog.String("track_suffix", suffix))

	return nil
}

func (sas *sendAnnounceStream) setEndedAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	// Check if the same track has announced already
	if old, ok := sas.announced[suffix]; ok {
		if old.AnnounceStatus == message.ENDED {
			return fmt.Errorf("ended announcement already set: %s", suffix)
		}
	}

	// Check if the same track is to announce
	if old, ok := sas.pending[suffix]; ok {
		if old.AnnounceStatus == message.ENDED {
			// Skip
			slog.Warn("ended announcement already set", "track_suffix", suffix)
			return nil
		}
	}

	am := message.AnnounceMessage{
		AnnounceStatus: message.ENDED,
		TrackSuffix:    suffix,
	}

	sas.pending[suffix] = am

	slog.Debug("set ended announcement", slog.String("track_suffix", suffix))

	return nil
}

func (sas *sendAnnounceStream) sendAnnouncements() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	if len(sas.announced) == 0 {
		return nil
	}

	// Calculate the total length of the ANNOUNCE messages
	var totalLen int
	for _, am := range sas.announced {
		totalLen += am.Len()
	}
	live := message.AnnounceMessage{
		AnnounceStatus: message.LIVE,
	}
	totalLen += live.Len()

	// Create a buffer
	buf := bytes.NewBuffer(make([]byte, 0, totalLen))

	// Encode the ANNOUNCE messages
	for _, am := range sas.announced {
		// Encode the ANNOUNCE message
		_, err := am.Encode(buf)
		if err != nil {
			slog.Error("failed to encode an ANNOUNCE message", "error", err)
			return err
		}
	}
	// Encode the LIVE message
	_, err := live.Encode(buf)
	if err != nil {
		slog.Error("failed to encode a LIVE message", "error", err)
		return err
	}

	// Send the ANNOUNCE messages
	_, err = sas.stream.Write(buf.Bytes())
	if err != nil {
		slog.Error("failed to send ANNOUNCE messages", "error", err)
		return err
	}

	slog.Debug("sent announcement messages successfully", slog.Any("stream_id", sas.stream.StreamID()))

	return nil
}

package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
)

type AnnouncementWriter interface {
	SendAnnouncements(announcements []*Announcement) error
	AnnounceConfig() AnnounceConfig
	Close() error
	CloseWithError(error) error
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

func newSendAnnounceStream(internalStream *internal.SendAnnounceStream) *sendAnnounceStream {
	sas := &sendAnnounceStream{
		internalStream: internalStream,
		sendCh:         make(chan struct{}, 1),
	}

	go func() {
		for range sas.sendCh {
			err := sas.internalStream.SendAnnouncements()
			if err != nil {
				slog.Error("failed to send announcements", "err", err)
			}
		}
	}()

	return sas
}

type sendAnnounceStream struct {
	internalStream *internal.SendAnnounceStream

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

		err = sas.internalStream.SetActiveAnnouncement(path)
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

			err := sas.internalStream.SetEndedAnnouncement(ann.TrackPath().String())
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

func (s *sendAnnounceStream) AnnounceConfig() AnnounceConfig {
	return AnnounceConfig{
		TrackPattern: s.internalStream.AnnouncePleaseMessage.TrackPattern,
	}
}

func (s *sendAnnounceStream) TrackPattern() string {
	return s.internalStream.AnnouncePleaseMessage.TrackPattern
}

func (s *sendAnnounceStream) Close() error {
	return s.internalStream.Close()
}

func (s *sendAnnounceStream) CloseWithError(err error) error {
	return s.internalStream.CloseWithError(err)
}

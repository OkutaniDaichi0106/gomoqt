package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
)

type AnnouncementWriter interface {
	SendAnnouncement(announcements []*Announcement) error
	AnnounceConfig() AnnounceConfig
	Close() error
	CloseWithError(error) error
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	internalStream *internal.SendAnnounceStream
}

func (sas *sendAnnounceStream) SendAnnouncement(announcements []*Announcement) error {
	var err error
	var suffix string
	for _, ann := range announcements {
		if !ann.TrackPath.HasPrefix(sas.TrackPrefix()) {
			continue
		}
		// Get the suffix of the track path from the prefix
		suffix = ann.TrackPath.GetSuffix(sas.TrackPrefix())

		if ann.active {
			err = sas.internalStream.SetActiveAnnouncement(suffix)
			if err != nil {
				return err
			}
		} else {
			err = sas.internalStream.SetEndedAnnouncement(suffix)
			if err != nil {
				return err
			}
		}
	}

	err = sas.internalStream.SendAnnouncements()
	if err != nil {
		return err
	}

	return nil
}

func (s *sendAnnounceStream) AnnounceConfig() AnnounceConfig {
	return AnnounceConfig{
		TrackPrefix: s.internalStream.AnnouncePleaseMessage.TrackPrefix,
	}
}

func (s *sendAnnounceStream) TrackPrefix() string {
	return s.internalStream.AnnouncePleaseMessage.TrackPrefix
}

func (s *sendAnnounceStream) Close() error {
	return s.internalStream.Close()
}

func (s *sendAnnounceStream) CloseWithError(err error) error {
	return s.internalStream.CloseWithError(err)
}

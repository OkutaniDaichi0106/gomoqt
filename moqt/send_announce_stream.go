package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type SendAnnounceStream interface {
	SendAnnouncement(announcements []Announcement) error
	AnnounceConfig() AnnounceConfig
	Close() error
	CloseWithError(error) error
}

var _ SendAnnounceStream = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	internalStream *internal.SendAnnounceStream
}

func (s *sendAnnounceStream) SendAnnouncement(announcements []Announcement) error {
	var err error
	for _, a := range announcements {
		switch a.AnnounceStatus {
		case ACTIVE:
			err = s.internalStream.SendActiveAnnouncement(a.TrackPath, message.Parameters(a.AnnounceParameters.paramMap))
		case LIVE:
			err = s.internalStream.SendLiveAnnouncement(message.Parameters(a.AnnounceParameters.paramMap))
		case ENDED:
			err = s.internalStream.SendEndedAnnouncement(a.TrackPath, message.Parameters(a.AnnounceParameters.paramMap))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *sendAnnounceStream) AnnounceConfig() AnnounceConfig {
	return AnnounceConfig{
		TrackPrefix: s.internalStream.AnnouncePleaseMessage.TrackPrefix,
		Parameters:  Parameters{s.internalStream.AnnouncePleaseMessage.AnnounceParameters},
	}
}

func (s *sendAnnounceStream) Close() error {
	return s.internalStream.Close()
}

func (s *sendAnnounceStream) CloseWithError(err error) error {
	return s.internalStream.CloseWithError(err)
}

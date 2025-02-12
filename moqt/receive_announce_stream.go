package moqt

import (
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
)

type AnnouncementReader interface {
	ReceiveAnnouncements() ([]Announcement, error)
}

var _ AnnouncementReader = (*receiveAnnounceStream)(nil)

type receiveAnnounceStream struct {
	internalStream *internal.ReceiveAnnounceStream
	closed         bool
	closeErr       error
	mu             sync.RWMutex
}

func (ras *receiveAnnounceStream) ReceiveAnnouncements() ([]Announcement, error) {
	ras.mu.RLock()
	defer ras.mu.RUnlock()

	if ras.closed {
		return nil, ras.closeErr
	}

	ams, err := ras.internalStream.ReceiveAnnouncements()
	if err != nil {
		return nil, err
	}

	anns := make([]Announcement, 0, len(ams)+1)
	for _, am := range ams {
		anns = append(anns, Announcement{
			TrackPath:          am.TrackSuffix,
			AnnounceStatus:     AnnounceStatus(am.AnnounceStatus),
			AnnounceParameters: Parameters{am.AnnounceParameters},
		})
	}

	return anns, nil
}

package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Subscriber struct {
	Handler SubscriberHandler

	RemoteTrack [][]string

	announcementCh <-chan struct {
		Announcement
		AnnounceResponceWriter
	}

	infoCh <-chan Info

	dataCh <-chan struct {
		Group
		Stream
	}
	//
	// interestW InterestWriter

	//
	// announceR  AnnounceReader
	// announceRW AnnounceResponceWriter

	// subscribeW SubscribeWriter
}

type SubscriberHandler interface {
	AnnounceHandler
	InfoHandler
	DataHander

	InterestWriter
	SubscribeWriter
	InfoRequestWriter
}

func (s Subscriber) init() {
	s.announcementCh = make(<-chan struct {
		Announcement
		AnnounceResponceWriter
	}, 1<<2)

	s.infoCh = make(<-chan Info, 1<<2)

	s.dataCh = make(<-chan struct {
		Group
		Stream
	}, 1<<2)
}

func (s Subscriber) listen() {
	for {
		select {
		case v := <-s.announcementCh:
			s.Handler.HandleAnnounce(v.Announcement, v.AnnounceResponceWriter)
		case v := <-s.infoCh:
			s.Handler.HandleInfo(v)
		case v := <-s.dataCh:
			s.Handler.HandleData(v.Group, v.Stream)
		}
	}
}

func getAnnouncement(r quicvarint.Reader) (Announcement, error) {
	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err := am.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Return an Announcement
	announcement := Announcement{
		TrackNamespace: am.TrackNamespace,
		Parameters:     am.Parameters,
	}

	authInfo, ok := getAuthorizationInfo(am.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

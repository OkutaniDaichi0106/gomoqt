package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Subscriber struct {
	Handler SubscriberHandler
}

type SubscriberHandler interface {
	AnnounceHandler
	InfoHandler
	DataHander

	InterestWriter
	SubscribeWriter
	InfoRequestWriter
}

func getAnnouncement(r quicvarint.Reader) (Announcement, error) {
	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err := am.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Initialize an Announcement
	announcement := Announcement{
		TrackNamespace: am.TrackNamespace,
		Parameters:     am.Parameters,
	}
	//
	authInfo, ok := getAuthorizationInfo(am.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

func getGroup(r quicvarint.Reader) (Group, error) {
	// Read a GROUP message
	var gm message.GroupMessage
	err := gm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a GROUP message", slog.String("error", err.Error()))
		return Group{}, err
	}

	//
	return Group(gm), nil
}

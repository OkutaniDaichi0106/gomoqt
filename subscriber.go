package moqt

import (
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

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
		TrackNamespace: strings.Join(am.TrackNamespace, "/"),
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
	return Group{
		subscribeID:       SubscribeID(gm.SubscribeID),
		groupSequence:     GroupSequence(gm.GroupSequence),
		PublisherPriority: PublisherPriority(gm.PublisherPriority),
	}, nil
}

func getInfo(r quicvarint.Reader) (Info, error) {
	//
	var im message.InfoMessage
	err := im.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	return Info(im), nil
}

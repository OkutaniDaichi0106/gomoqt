package moqt

import (
	"io"
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

func readAnnouncement(r moq.ReceiveStream) (Announcement, error) {
	// Get a new message reader
	mr, err := message.NewReader(r)
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Read an ANNOUNCE message
	var am message.AnnounceMessage
	err = am.Decode(mr)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE message", slog.String("error", err.Error()))
		return Announcement{}, err
	}

	// Initialize an Announcement
	announcement := Announcement{
		TrackNamespace: strings.Join(am.TrackNamespace, "/"),
		Parameters:     Parameters(am.Parameters),
	}

	//
	authInfo, ok := getAuthorizationInfo(announcement.Parameters)
	if ok {
		announcement.AuthorizationInfo = authInfo
	}

	return announcement, nil
}

func readGroup(r io.Reader) (Group, error) {
	// Get a message reader
	mr, err := message.NewReader(r)
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return Group{}, err
	}

	// Read a GROUP message
	var gm message.GroupMessage
	err = gm.Decode(mr)
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

func readInfo(r io.Reader) (Info, error) {
	// Get a message reader
	mr, err := message.NewReader(r)
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return Info{}, err
	}

	// Read an INFO message
	var im message.InfoMessage
	err = im.Decode(mr)
	if err != nil {
		slog.Error("failed to read a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	return Info(im), nil
}

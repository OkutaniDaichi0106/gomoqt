package moqt

import (
	"io"
	"log/slog"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type RequestHandler interface {
	InterestHandler
	SubscribeHandler
	FetchHandler
	InfoRequestHandler
}

func readInterest(str io.Reader) (Interest, error) {
	r, err := message.NewReader(str)
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return Interest{}, err
	}
	//
	var aim message.AnnounceInterestMessage
	err = aim.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return Interest{}, err
	}

	return Interest{
		TrackPrefix: strings.Join(aim.TrackPrefix, "/"),
		Parameters:  Parameters(aim.Parameters),
	}, nil
}

func readFetchRequest(str moq.Stream) (FetchRequest, error) {
	r, err := message.NewReader(str)
	if err != nil {
		slog.Error("failed to get a new message reader")
	}

	var frm message.FetchMessage
	err = frm.Decode(r)
	if err != nil {
		slog.Error("failed to read a FETCH message", slog.String("error", err.Error()))
		return FetchRequest{}, err
	}

	return FetchRequest(frm), nil
}

func readInfoRequest(str moq.Stream) (InfoRequest, error) {
	r, err := message.NewReader(str)
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
	}

	var irm message.InfoRequestMessage
	err = irm.Decode(r)
	if err != nil {
		slog.Error("failed to read an INFO_REQUEST message", slog.String("error", err.Error()))
		return InfoRequest{}, err
	}

	return InfoRequest{
		TrackNamespace: strings.Join(irm.TrackNamespace, "/"),
		TrackName:      irm.TrackName,
	}, nil
}

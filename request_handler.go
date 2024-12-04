package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type RequestHandler interface {
	InterestHandler
	SubscribeHandler
	FetchHandler
	InfoRequestHandler
}

func readInterest(r io.Reader) (Interest, error) {
	//
	var aim message.AnnounceInterestMessage
	err := aim.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return Interest{}, err
	}

	return Interest{
		TrackPrefix: aim.TrackPrefix,
		Parameters:  Parameters(aim.Parameters),
	}, nil
}

func readFetchRequest(r io.Reader) (FetchRequest, error) {
	var frm message.FetchMessage
	err := frm.Decode(r)
	if err != nil {
		slog.Error("failed to read a FETCH message", slog.String("error", err.Error()))
		return FetchRequest{}, err
	}

	req := FetchRequest{
		TrackNamespace:     frm.TrackNamespace,
		TrackName:          frm.TrackName,
		SubscriberPriority: SubscriberPriority(frm.SubscriberPriority),
		GroupSequence:      GroupSequence(frm.GroupSequence),
		FrameSequence:      FrameSequence(frm.FrameSequence),
	}

	return req, nil
}

func readInfoRequest(r io.Reader) (InfoRequest, error) {

	var irm message.InfoRequestMessage
	err := irm.Decode(r)
	if err != nil {
		slog.Error("failed to read an INFO_REQUEST message", slog.String("error", err.Error()))
		return InfoRequest{}, err
	}

	return InfoRequest{
		TrackNamespace: irm.TrackNamespace,
		TrackName:      irm.TrackName,
	}, nil
}

package moqtransfork

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

// type RequestHandler interface {
// 	InterestHandler
// 	SubscribeHandler
// 	FetchHandler
// 	InfoRequestHandler
// }

func readInterest(r io.Reader) (Interest, error) {
	//
	var aim message.AnnounceInterestMessage
	err := aim.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return Interest{}, err
	}

	return Interest{
		TrackPrefix: aim.TrackPathPrefix,
		Parameters:  Parameters(aim.Parameters),
	}, nil
}

func writeInterest(w io.Writer, interest Interest) error {
	aim := message.AnnounceInterestMessage{
		TrackPathPrefix: interest.TrackPrefix,
		Parameters:      message.Parameters(interest.Parameters),
	}

	err := aim.Encode(w)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return err
	}
	return nil
}

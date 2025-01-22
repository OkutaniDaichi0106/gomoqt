package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type AnnounceConfig struct {
	TrackPrefix []string
	Parameters  Parameters
}

func readInterest(r io.Reader) (AnnounceConfig, error) {
	//
	var aim message.AnnouncePleaseMessage
	err := aim.Decode(r)
	if err != nil {
		slog.Error("failed to read an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return AnnounceConfig{}, err
	}

	return AnnounceConfig{
		TrackPrefix: aim.TrackPathPrefix,
		Parameters:  Parameters{aim.Parameters},
	}, nil
}

func writeInterest(w io.Writer, interest AnnounceConfig) error {
	aim := message.AnnouncePleaseMessage{
		TrackPathPrefix: interest.TrackPrefix,
		Parameters:      message.Parameters(interest.Parameters.paramMap),
	}

	err := aim.Encode(w)
	if err != nil {
		slog.Error("failed to send an ANNOUNCE_INTEREST message", slog.String("error", err.Error()))
		return err
	}
	return nil
}

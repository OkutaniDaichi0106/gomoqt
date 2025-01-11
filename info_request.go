package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

func readInfo(r io.Reader) (Info, error) {
	// Read an INFO message
	var im message.InfoMessage
	err := im.Decode(r)
	if err != nil {
		slog.Error("failed to read a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}

	return info, nil
}

func writeInfoRequest(w io.Writer, req InfoRequest) error {
	// Send an INFO_REQUEST message
	im := message.InfoRequestMessage{
		TrackPath: req.TrackPath,
	}
	err := im.Encode(w)
	if err != nil {
		slog.Error("failed to send an INFO_REQUEST message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

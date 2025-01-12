package moqtransfork

import (
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type Info struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
	GroupExpires        time.Duration
}

func readInfo(r io.Reader) (Info, error) {
	slog.Debug("reading an info")

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

	slog.Debug("read an info")

	return info, nil
}

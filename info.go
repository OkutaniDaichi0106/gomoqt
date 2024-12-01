package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

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

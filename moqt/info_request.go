package moqt

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type InfoRequest struct {
	TrackPath []string
}

func (ir InfoRequest) String() string {
	return fmt.Sprintf("InfoRequest: { TrackPath: %s }", TrackPartsString(ir.TrackPath))
}

func readInfoRequest(r io.Reader) (InfoRequest, error) {
	var irm message.InfoRequestMessage
	err := irm.Decode(r)
	if err != nil {
		slog.Error("failed to read an INFO_REQUEST message", slog.String("error", err.Error()))
		return InfoRequest{}, err
	}

	req := InfoRequest{
		TrackPath: irm.TrackPath,
	}

	return req, nil
}

func writeInfoRequest(w io.Writer, req InfoRequest) error {
	// Send an INFO_REQUEST message
	im := message.InfoRequestMessage{
		TrackPath: req.TrackPath,
	}
	err := im.Encode(w)
	if err != nil {
		slog.Error("failed to encode an INFO_REQUEST message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

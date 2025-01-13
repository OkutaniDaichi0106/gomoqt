package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type SubscribeGap struct {
	StartGroupSequence GroupSequence
	Count              uint64
	GroupErrorCode     GroupErrorCode
}

func readSubscribeGap(r io.Reader) (SubscribeGap, error) {
	var sgm message.SubscribeGapMessage
	err := sgm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return SubscribeGap{}, err
	}

	return SubscribeGap{
		StartGroupSequence: GroupSequence(sgm.GroupStartSequence),
		Count:              sgm.Count,
		GroupErrorCode:     GroupErrorCode(sgm.GroupErrorCode),
	}, nil
}

func writeSubscribeGap(w io.Writer, gap SubscribeGap) error {
	sgm := message.SubscribeGapMessage{
		GroupStartSequence: message.GroupSequence(gap.StartGroupSequence),
		Count:              gap.Count,
		GroupErrorCode:     message.GroupErrorCode(gap.GroupErrorCode),
	}

	err := sgm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return err
	}
	return nil
}

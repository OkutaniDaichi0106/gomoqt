package moqt

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type SubscribeGap struct {
	MinGapSequence GroupSequence
	MaxGapSequence GroupSequence
	GroupErrorCode GroupErrorCode
}

func (sg SubscribeGap) String() string {
	return fmt.Sprintf("SubscribeGap: { MinGapSequence: %d, MaxGapSequence: %d, GroupErrorCode: %d }", sg.MinGapSequence, sg.MaxGapSequence, sg.GroupErrorCode)
}

func readSubscribeGap(r io.Reader) (SubscribeGap, error) {
	var sgm message.SubscribeGapMessage
	_, err := sgm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return SubscribeGap{}, err
	}

	return SubscribeGap{
		MinGapSequence: GroupSequence(sgm.MinGapSequence),
		MaxGapSequence: GroupSequence(sgm.MaxGapSequence),
		GroupErrorCode: GroupErrorCode(sgm.GroupErrorCode),
	}, nil
}

func writeSubscribeGap(w io.Writer, gap SubscribeGap) error {
	sgm := message.SubscribeGapMessage{
		MinGapSequence: message.GroupSequence(gap.MinGapSequence),
		MaxGapSequence: message.GroupSequence(gap.MaxGapSequence),
		GroupErrorCode: message.GroupErrorCode(gap.GroupErrorCode),
	}

	_, err := sgm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return err
	}
	return nil
}

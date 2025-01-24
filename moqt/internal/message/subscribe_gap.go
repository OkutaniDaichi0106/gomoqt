package message

import (
	"io"
	"log/slog"
)

type GroupErrorCode uint64

/*
 * SUBSCRIBE_GAP Message {
 *   Min Gap Sequence (varint),
 *   Max Gap Sequence (varint),
 *   Group Error Code (varint),
 * }
 */
type SubscribeGapMessage struct {
	MinGapSequence GroupSequence
	MaxGapSequence GroupSequence
	GroupErrorCode GroupErrorCode
}

func (sgm SubscribeGapMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SUBSCRIBE_GAP message")

	// Serialize the payload
	p := make([]byte, 0, 1<<5)
	p = appendNumber(p, uint64(sgm.MinGapSequence))
	p = appendNumber(p, uint64(sgm.MaxGapSequence))
	p = appendNumber(p, uint64(sgm.GroupErrorCode))

	// Prepare the final message with length prefix
	b := make([]byte, 0, len(p)+8)
	b = appendNumber(b, uint64(len(p)))
	b = append(b, p...)

	// Write the message
	if _, err := w.Write(b); err != nil {
		slog.Error("failed to write a SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a SUBSCRIBE_GAP message")
	return nil
}

func (sgm *SubscribeGapMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SUBSCRIBE_GAP message")

	// Create a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Deserialize the payload
	num, err := readNumber(mr)
	if err != nil {
		return err
	}
	sgm.MinGapSequence = GroupSequence(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	sgm.MaxGapSequence = GroupSequence(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	sgm.GroupErrorCode = GroupErrorCode(num)

	slog.Debug("decoded a SUBSCRIBE_GAP message")
	return nil
}

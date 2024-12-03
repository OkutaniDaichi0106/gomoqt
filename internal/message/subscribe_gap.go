package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupErrorCode uint64

type SubscribeGapMessage struct {
	GroupStartSequence GroupSequence
	Count              uint64
	GroupErrorCode     GroupErrorCode
}

func (sgm SubscribeGapMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a GROUP_DROP message")

	/*
	 * Serialize the payload in the following format
	 *
	 * GROUP_DROP Message Payload {
	 *   Group Start Sequence (varint),
	 *   Count (varint),
	 *   Group Error Code (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<5)

	// Append the Group Start Sequence
	p = quicvarint.Append(p, uint64(sgm.GroupStartSequence))

	// Append the Count
	p = quicvarint.Append(p, sgm.Count)

	// Append the Group Error Code
	p = quicvarint.Append(p, uint64(sgm.GroupErrorCode))

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (sgm *SubscribeGapMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a GROUP_DROP message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Group Start Sequence
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sgm.GroupStartSequence = GroupSequence(num)

	// Get a Count
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sgm.Count = num

	// Get a Group Error Code
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sgm.GroupErrorCode = GroupErrorCode(num)

	return nil
}

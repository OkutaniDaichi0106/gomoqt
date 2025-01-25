package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
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

func (sgm SubscribeGapMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := make([]byte, 0, 1<<5)
	p = appendNumber(p, uint64(sgm.MinGapSequence))
	p = appendNumber(p, uint64(sgm.MaxGapSequence))
	p = appendNumber(p, uint64(sgm.GroupErrorCode))

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	return w.Write(b)
}

func (sgm *SubscribeGapMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	num, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	sgm.MinGapSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	sgm.MaxGapSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	sgm.GroupErrorCode = GroupErrorCode(num)

	return n, nil
}

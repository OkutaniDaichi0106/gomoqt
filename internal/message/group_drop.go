package message

import (
	"io"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupErrorCode uint64

type GroupDrop struct {
	GroupStartSequence GroupSequence
	Count              uint64
	GroupErrorCode     GroupErrorCode
}

func (gd GroupDrop) Encode(w io.Writer) error {
	/*
	 * Serialize the payload in the following format
	 *
	 *
	 *
	 */
	p := make([]byte, 0, 1<<5)

	// Append the Group Start Sequence
	p = quicvarint.Append(p, uint64(gd.GroupStartSequence))

	// Append the Count
	p = quicvarint.Append(p, gd.Count)

	// Append the Group Error Code
	p = quicvarint.Append(p, uint64(gd.GroupErrorCode))

	log.Print("GROUP_DROP payload", p)

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

func (gd *GroupDrop) Decode(r Reader) error {
	// Get a Group Start Sequence
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	gd.GroupStartSequence = GroupSequence(num)

	// Get a Count
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	gd.Count = num

	// Get a Group Error Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	gd.GroupErrorCode = GroupErrorCode(num)

	return nil
}

package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchMessage struct {
	TrackNamespace     TrackNamespace
	TrackName          string
	SubscriberPriority SubscriberPriority
	GroupSequence      GroupSequence
	GroupOffset        uint64
}

func (fm FetchMessage) Encode(w io.Writer) error {
	/*
	 * Serialize the message in the following format
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = appendTrackNamespace(p, fm.TrackNamespace)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(fm.TrackName)))
	p = append(p, []byte(fm.TrackName)...)

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(fm.SubscriberPriority))

	// Append the Group Sequence
	p = quicvarint.Append(p, uint64(fm.GroupSequence))

	// Append the Group Offset
	p = quicvarint.Append(p, fm.GroupOffset)

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

func (fm *FetchMessage) Decode(r io.Reader) error {
	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace
	fm.TrackNamespace, err = readTrackNamespace(mr)
	if err != nil {
		return err
	}

	// Get a Track Name
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}

	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	fm.TrackName = string(buf)

	// Get a Subscriber Priority
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.SubscriberPriority = SubscriberPriority(num)

	// Get a Group Sequence
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.GroupSequence = GroupSequence(num)

	// Get a Group Offset
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.GroupOffset = num

	return nil
}

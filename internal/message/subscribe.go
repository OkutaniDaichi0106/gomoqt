package message

import (
	"io"
	"log/slog"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64

type TrackAlias uint64

type TrackPriority byte

type GroupOrder byte

type SubscribeMessage struct {
	SubscribeID SubscribeID

	TrackPath string

	TrackPriority TrackPriority

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder GroupOrder

	GroupExpires time.Duration

	/***/
	MinGroupSequence GroupSequence

	/***/
	MaxGroupSequence GroupSequence

	/*
	 * Subscribe Parameters
	 * Parameters should contain Authorization Information
	 */
	Parameters Parameters
}

func (s SubscribeMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SUBSCRIBE message")

	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE Message payload {
	 *   Track Path (string),
	 *   Subscriber Priority (8),
	 *   Group Order (8),
	 *   Group Expires (varint),
	 *   Min Group Sequence (varint),
	 *   Max Group Sequence (varint),
	 *   Subscribe Parameters (Parameters),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<6)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(s.SubscribeID))

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(s.TrackPath)))
	p = append(p, []byte(s.TrackPath)...)

	// Append the Subscriber Priority
	p = append(p, []byte{byte(s.TrackPriority)}...)

	// Append the Group Order
	p = append(p, []byte{byte(s.GroupOrder)}...)

	// Append the Group Expires
	p = append(p, []byte{byte(s.GroupOrder)}...)

	// Append the Min Group Sequence
	p = quicvarint.Append(p, uint64(s.MinGroupSequence))

	// Append the Max Group Sequence
	p = quicvarint.Append(p, uint64(s.MinGroupSequence))

	// Append the Subscribe Update Priority
	p = appendParameters(p, s.Parameters)

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a SUBSCRIBE message")

	return err
}

func (s *SubscribeMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SUBSCRIBE message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get Subscribe ID
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	s.SubscribeID = SubscribeID(num)

	// Get Track Name
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	s.TrackPath = string(buf)

	// Get Subscriber Priority
	bnum, err := mr.ReadByte()
	if err != nil {
		return err
	}
	s.TrackPriority = TrackPriority(bnum)

	// Get Group Order
	bnum, err = mr.ReadByte()
	if err != nil {
		return err
	}
	s.GroupOrder = GroupOrder(bnum)

	// Get Group Expires
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	s.GroupExpires = time.Duration(num)

	// Get Min Group Sequence
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	s.MinGroupSequence = GroupSequence(num)

	// Get Max Group Sequence
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	s.MaxGroupSequence = GroupSequence(num)

	// Get Subscribe Update Parameters
	s.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a SUBSCRIBE message")

	return nil
}

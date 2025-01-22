package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64

type TrackPriority byte

type GroupOrder byte

type SubscribeMessage struct {
	SubscribeID SubscribeID

	TrackPath []string

	TrackPriority TrackPriority

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder GroupOrder

	/***/
	MinGroupSequence GroupSequence

	/***/
	MaxGroupSequence GroupSequence

	/*
	 * Subscribe SubscribeParameters
	 */
	SubscribeParameters Parameters
}

func (s SubscribeMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SUBSCRIBE message")

	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE Message payload {
	 *   Subscribe ID (varint),
	 *   Track Path ([]string),
	 *   Track Priority (varint),
	 *   Group Order (varint),
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

	// Append the Track Path's length
	p = quicvarint.Append(p, uint64(len(s.TrackPath)))

	// Append the Track Path
	for _, part := range s.TrackPath {
		// Append the Track Namespace Prefix Part
		p = quicvarint.Append(p, uint64(len(part)))
		p = append(p, []byte(part)...)
	}

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(s.TrackPriority))

	// Append the Group Order
	p = quicvarint.Append(p, uint64(s.GroupOrder))

	// Append the Min Group Sequence
	p = quicvarint.Append(p, uint64(s.MinGroupSequence))

	// Append the Max Group Sequence
	p = quicvarint.Append(p, uint64(s.MaxGroupSequence))

	// Append the Subscribe Parameters
	p = appendParameters(p, s.SubscribeParameters)

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

	// Get Track Path
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	count := num
	s.TrackPath = make([]string, count)

	// Get Track Path Parts
	for i := uint64(0); i < count; i++ {
		num, err = quicvarint.Read(mr)
		if err != nil {
			return err
		}
		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return err
		}
		s.TrackPath[i] = string(buf)
	}

	// Get Track Priority
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	s.TrackPriority = TrackPriority(num)

	// Get Group Order
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	s.GroupOrder = GroupOrder(num)

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

	// Get Subscribe Parameters
	s.SubscribeParameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a SUBSCRIBE message")

	return nil
}

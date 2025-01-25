package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupSequence uint64

type GroupMessage struct {
	SubscribeID   SubscribeID
	GroupSequence GroupSequence
	TrackPriority TrackPriority
}

func (g GroupMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a GROUP message")

	// Serialize the payload
	p := make([]byte, 0, 1<<4)
	p = appendNumber(p, uint64(g.SubscribeID))
	p = appendNumber(p, uint64(g.GroupSequence))
	p = appendNumber(p, uint64(g.TrackPriority))

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	// Write
	return w.Write(b)
}

func (g *GroupMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a GROUP message")

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
	g.SubscribeID = SubscribeID(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	g.GroupSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	g.TrackPriority = TrackPriority(num)

	slog.Debug("decoded a GROUP message")
	return n, nil
}

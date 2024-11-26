package message

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type SessionUpdateMessage struct {
	/*
	 * Versions selected by the server
	 */
	Bitrate uint64
}

func (sum SessionUpdateMessage) SerializePayload() []byte {
	/*
	 * Serialize the message in the following format
	 *
	 * SESSION_UPDATE Message Payload {
	 *   Bitrate (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<4)

	p = quicvarint.Append(p, sum.Bitrate)

	return p
}

func (sum *SessionUpdateMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get a bitrate
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sum.Bitrate = num

	return nil
}

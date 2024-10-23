package message

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type SessionServerMessage struct {
	/*
	 * Versions selected by the server
	 */
	SelectedVersion uint64

	/*
	 * Setup Parameters
	 * Keys of the maps should not be duplicated
	 */
	Parameters Parameters
}

func (ssm SessionServerMessage) SerializePayload() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SERVER_SETUP Message Payload {
	 *   Selected Version (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */

	p := make([]byte, 0, 1<<8)

	// Append the selected version
	p = quicvarint.Append(p, uint64(ssm.SelectedVersion))

	// Append the parameters
	p = ssm.Parameters.Append(p)

	return p
}

func (ssm *SessionServerMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ssm.SelectedVersion = num

	err = ssm.Parameters.Deserialize(r)

	if err != nil {
		return err
	}
	return nil
}

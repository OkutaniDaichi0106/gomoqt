package moqtmessage

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type ServerSetupMessage struct {
	/*
	 * Versions selected by the server
	 */
	SelectedVersion protocol.Version

	/*
	 * Setup Parameters
	 * Keys of the maps should not be duplicated
	 */
	Parameters Parameters
}

func (ss ServerSetupMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SERVER_SETUP Message {
	 *   Type (varint) = 0x41,
	 *   Length (varint),
	 *   Selected Version (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */

	p := make([]byte, 0, 1<<8)

	// Append the selected version
	p = quicvarint.Append(p, uint64(ss.SelectedVersion))

	// Append the parameters
	p = ss.Parameters.append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(SERVER_SETUP))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (ss *ServerSetupMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ss.SelectedVersion = protocol.Version(num)

	err = ss.Parameters.Deserialize(r)

	if err != nil {
		return err
	}
	return nil
}

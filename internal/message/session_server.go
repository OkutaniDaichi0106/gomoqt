package message

import (
	"io"
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SessionServerMessage struct {
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

func (ssm SessionServerMessage) Encode(w io.Writer) error {
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
	p = appendParameters(p, ssm.Parameters)

	log.Print("SESSION_SERVER payload", p)

	// Get a whole serialized message
	b := make([]byte, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (ssm *SessionServerMessage) Decode(r quicvarint.Reader) error {
	// Get a Version
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	ssm.SelectedVersion = protocol.Version(num)

	// Get Parameters
	ssm.Parameters, err = readParameters(r)
	if err != nil {
		return err
	}

	return nil
}

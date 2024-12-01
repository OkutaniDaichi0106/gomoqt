package message

import (
	"io"

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

	p := make([]byte, 0, 1<<4)

	// Append the selected version
	p = quicvarint.Append(p, uint64(ssm.SelectedVersion))

	// Append the parameters
	p = appendParameters(p, ssm.Parameters)

	// Get a whole serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (ssm *SessionServerMessage) Decode(r io.Reader) error {
	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Version
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	ssm.SelectedVersion = protocol.Version(num)

	// Get Parameters
	ssm.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	return nil
}

package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type ServerSetupMessage struct {
	/*
	 * Versions selected by the server
	 */
	SelectedVersion Version

	/*
	 * Setup Parameters
	 * Keys of the maps should not be duplicated
	 */
	Parameters
}

func (ss ServerSetupMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SERVER_SETUP Payload {
	 *   Selected Version (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */
	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<8) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SERVER_SETUP))
	// Append the selected version
	b = quicvarint.Append(b, uint64(ss.SelectedVersion))

	// Serialize the parameters and append it
	/*
	 * Setup Parameters {
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = ss.Parameters.append(b)

	return b
}

func (ss *ServerSetupMessage) deserialize(r quicvarint.Reader) error {
	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != SERVER_SETUP {
		return errors.New("unexpected message")
	}

	return ss.deserializeBody(r)
}

func (ss *ServerSetupMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ss.SelectedVersion = Version(num)

	err = ss.Parameters.parse(r)
	if err != nil {
		return err
	}
	return nil

	// PATH parameter must not be included in the parameters
	// Keys of the parameters must not be duplicate
}

package moqtransport

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type ClientSetupMessage struct {
	/*
	 * Versions supported by the client
	 */
	Versions []Version

	/*
	 * Setup Parameters
	 */
	Parameters
}

func (cs ClientSetupMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * CLIENT_SETUP Payload {
	 *   Number of Supported Versions (varint),
	 *   Supported Versions (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */

	// Check the condition of the parameters
	// 1. At least one version is required
	if len(cs.Versions) == 0 {
		panic("no version is specifyed")
	}
	// 2. Parameters should conclude role parameter
	ok, _ := cs.Parameters.Contain(ROLE)
	if !ok {
		panic("no role is specifyed")
	}

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(CLIENT_SETUP))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(cs.Versions)))
	var version Version
	for _, version = range cs.Versions {
		b = quicvarint.Append(b, uint64(version))
	}

	// Serialize the parameters and append it
	/*
	 * Setup Parameters {
	 *   Role Parameter (varint),
	 *   [Path Parameter (stirng),]
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = cs.Parameters.append(b)

	return b
}

// func (cs *ClientSetupMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != CLIENT_SETUP {
// 		return errors.New("unexpected message")
// 	}

// 	return cs.deserializeBody(r)
// }

func (cs *ClientSetupMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get number of supported versions
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	count := num
	// Get supported versions
	for i := uint64(0); i < count; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		cs.Versions = append(cs.Versions, Version(num))
	}

	err = cs.Parameters.parse(r)
	if err != nil {
		return err
	}

	return nil
}

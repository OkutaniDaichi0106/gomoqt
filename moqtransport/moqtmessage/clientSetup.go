package moqtmessage

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtversion"

	"github.com/quic-go/quic-go/quicvarint"
)

type ClientSetupMessage struct {
	/*
	 * Versions supported by the client
	 */
	Versions []moqtversion.Version

	/*
	 * Setup Parameters
	 */
	Parameters Parameters
}

func (cs ClientSetupMessage) Serialize() []byte {
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
	_, ok := cs.Parameters.Role()
	if !ok {
		panic("no role is specifyed")
	}

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(CLIENT_SETUP))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(cs.Versions)))
	var version moqtversion.Version
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

func (cs *ClientSetupMessage) DeserializeBody(r quicvarint.Reader) error {
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
		cs.Versions = append(cs.Versions, moqtversion.Version(num))
	}

	err = cs.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}

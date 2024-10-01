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
	 * Serialize the message in the following formatt
	 *
	 * CLIENT_SETUP Payload {
	 *   Number of Supported Versions (varint),
	 *   Supported Versions (varint),
	 *   Number of Parameters (varint),
	 *   Setup Parameters (..),
	 * }
	 */

	// Verify if at least one version is required
	if len(cs.Versions) == 0 {
		panic("no version is specifyed")
	}

	// Verify if the Parameters conclude some role parameter
	_, ok := cs.Parameters.Role()
	if !ok {
		panic("no role is specifyed")
	}

	p := make([]byte, 0, 1<<8)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(cs.Versions)))
	var version moqtversion.Version
	for _, version = range cs.Versions {
		p = quicvarint.Append(p, uint64(version))
	}

	// Append the parameters
	p = cs.Parameters.append(p)

	/*
	 * Serialize the whole data
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(CLIENT_SETUP))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (cs *ClientSetupMessage) DeserializePayload(r quicvarint.Reader) error {
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

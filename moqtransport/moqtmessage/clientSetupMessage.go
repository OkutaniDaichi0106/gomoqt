package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type ClientSetupMessage struct {
	/*
	 * SupportedVersions supported by the client
	 */
	SupportedVersions []Version

	/*
	 * Setup Parameters
	 */
	Parameters Parameters
}

func (cs ClientSetupMessage) Serialize() []byte {
	// Verify if at least one version is specified
	if len(cs.SupportedVersions) == 0 {
		panic("no version is specified")
	}

	// Verify if the Parameters conclude some role parameter
	_, ok := cs.Parameters.Role()
	if !ok {
		panic("no role is specifyed")
	}

	p := make([]byte, 0, 1<<8)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(cs.SupportedVersions)))
	var version Version
	for _, version = range cs.SupportedVersions {
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
	// Get number of supported versions
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get supported versions
	count := num
	for i := uint64(0); i < count; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		cs.SupportedVersions = append(cs.SupportedVersions, Version(num))
	}

	// Get Parameters
	err = cs.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}

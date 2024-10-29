package message

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SessionClientMessage struct {
	/*
	 * SupportedVersions supported by the client
	 */
	SupportedVersions []protocol.Version

	/*
	 * Setup Parameters
	 */
	Parameters Parameters
}

func (scm SessionClientMessage) SerializePayload() []byte {
	// // Verify if at least one version is specified
	// if len(cs.SupportedVersions) == 0 {
	// 	panic("no version is specified")
	// }

	// // Verify if the Parameters conclude some role parameter
	// _, ok := cs.Parameters.Role()
	// if !ok {
	// 	panic("no role is specifyed")
	// }

	/*
	 * Serialize the payload in the following format
	 *
	 * CLIENT_SETUP Message Payload {
	 *   Supported Versions {
	 *     Count (varint),
	 *     Versions (varint...),
	 *   },
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */
	p := make([]byte, 0, 1<<8)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		p = quicvarint.Append(p, uint64(version))
	}

	// Append the parameters
	p = scm.Parameters.Append(p)

	return p
}

func (scm *SessionClientMessage) DeserializePayload(r quicvarint.Reader) error {
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
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(num))
	}

	// Get Parameters
	err = scm.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}

package message

import (
	"io"
	"log"

	"github.com/OkutaniDaichi0106/gomoqt/internal/protocol"
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

func (scm SessionClientMessage) Encode(w io.Writer) error {
	/*
	 * Serialize the payload in the following format
	 *
	 * CLIENT_SETUP Message Payload {
	 *   Supported Versions {
	 *     Count (varint),
	 *     Versions (varint...),
	 *   },
	 *   Announce Parameters (Parameters),
	 * }
	 */
	p := make([]byte, 0, 1<<8)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		p = quicvarint.Append(p, uint64(version))
	}

	// Append the parameters
	p = appendParameters(p, scm.Parameters)

	log.Print("SESSION_CLIENT payload", p)

	// Get a serialized message
	b := make([]byte, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (scm *SessionClientMessage) Decode(r Reader) error {
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
	scm.Parameters, err = readParameters(r)
	if err != nil {
		return err
	}

	return nil
}

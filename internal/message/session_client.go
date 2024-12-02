package message

import (
	"io"
	"log/slog"

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
	slog.Debug("encoding a SESSION_CLIENT message")
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
	p := make([]byte, 0, 1<<6)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		p = quicvarint.Append(p, uint64(version))
	}

	// Append the parameters
	p = appendParameters(p, scm.Parameters)

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (scm *SessionClientMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SESSION_CLIENT message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get number of supported versions
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}

	// Get supported versions
	count := num
	for i := uint64(0); i < count; i++ {
		num, err = quicvarint.Read(mr)
		if err != nil {
			return err
		}
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(num))
	}

	// Get Parameters
	scm.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	return nil
}

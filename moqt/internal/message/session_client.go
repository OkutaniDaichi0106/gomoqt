package message

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * SESSION_CLIENT Message {
 *   Supported Versions {
 *     Count (varint),
 *     Versions (varint...),
 *   },
 *   Session Client Parameters (Parameters),
 * }
 */

type SessionClientMessage struct {
	SupportedVersions []protocol.Version
	Parameters        Parameters
}

func (scm SessionClientMessage) Len() int {
	length := numberLen(uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		length += numberLen(uint64(version))
	}
	length += parametersLen(scm.Parameters)
	return length
}

func (scm SessionClientMessage) Encode(w io.Writer) error {
	// Serialize the payload
	p := getBytes()
	defer putBytes(p)

	// Append the supported versions
	p = AppendNumber(p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		p = AppendNumber(p, uint64(version))
	}

	// Append the parameters
	p = AppendParameters(p, scm.Parameters)

	_, err := w.Write(p)
	return err
}

func (scm *SessionClientMessage) Decode(r io.Reader) error {
	// Read version count
	num, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	// Read versions
	for i := uint64(0); i < num; i++ {
		version, _, err := ReadNumber(quicvarint.NewReader(r))
		if err != nil {
			return err
		}
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(version))
	}

	// Read parameters
	scm.Parameters, _, err = ReadParameters(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	return nil
}

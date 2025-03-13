package message

import (
	"bytes"
	"io"
	"log/slog"

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

func (scm SessionClientMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(scm.Len()))

	// Append the supported versions
	*p = AppendNumber(*p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		*p = AppendNumber(*p, uint64(version))
	}

	// Append the parameters
	*p = AppendParameters(*p, scm.Parameters)

	return w.Write(*p)
}

func (scm *SessionClientMessage) Decode(r io.Reader) (int, error) {

	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for SESSION_CLIENT message", "error", err)
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	// Read version count
	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read supported version count", "error", err)
		return n, err
	}

	// Read versions
	for i := uint64(0); i < num; i++ {
		version, _, err := ReadNumber(mr)
		if err != nil {
			slog.Error("failed to read a supported version", "error", err)
			return n, err
		}
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(version))
	}

	// Read parameters
	scm.Parameters, _, err = ReadParameters(mr)
	if err != nil {
		slog.Error("failed to read parameters", "error", err)
		return n, err
	}

	slog.Debug("decoded a SESSION_CLIENT message")

	return n, nil
}

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

func (scm SessionClientMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a SESSION_CLIENT message")

	// Serialize the payload
	p := make([]byte, 0, 1<<6)

	// Append the supported versions
	p = appendNumber(p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		p = appendNumber(p, uint64(version))
	}

	// Append the parameters
	p = appendParameters(p, scm.Parameters)

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	// Write
	n, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a SESSION_CLIENT message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("encoded a SESSION_CLIENT message")

	return n, nil
}

func (scm *SessionClientMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a SESSION_CLIENT message")

	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	// Read version count
	num, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}

	// Read versions
	for i := uint64(0); i < num; i++ {
		version, _, err := readNumber(mr)
		if err != nil {
			return n, err
		}
		scm.SupportedVersions = append(scm.SupportedVersions, protocol.Version(version))
	}

	// Read parameters
	scm.Parameters, _, err = readParameters(mr)
	if err != nil {
		return n, err
	}

	slog.Debug("decoded a SESSION_CLIENT message")

	return n, nil
}

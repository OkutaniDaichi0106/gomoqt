package message

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
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

func (scm SessionClientMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SESSION_CLIENT message")

	p := make([]byte, 0, 1<<6)

	// Append the supported versions
	p = appendNumber(p, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		p = appendNumber(p, uint64(version))
	}

	// Append the parameters
	p = appendParameters(p, scm.Parameters)

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload and the payload
	b = appendBytes(b, p)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a SESSION_CLIENT message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a SESSION_CLIENT message")

	return nil
}

func (scm *SessionClientMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SESSION_CLIENT message")

	// Get a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get number of supported versions
	num, err := readNumber(mr)
	if err != nil {
		return err
	}

	// Get supported versions
	count := num
	for i := uint64(0); i < count; i++ {
		num, err = readNumber(mr)
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

	slog.Debug("decoded a SESSION_CLIENT message")

	return nil
}

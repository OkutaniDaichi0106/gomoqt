package message

import (
	"io"

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

func (scm SessionClientMessage) Len() int {
	var l int

	l += VarintLen(uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		l += VarintLen(uint64(version))
	}
	l += ParametersLen(scm.Parameters)

	return l
}

func (scm SessionClientMessage) Encode(w io.Writer) error {
	msgLen := scm.Len()
	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
	defer pool.Put(b)

	b, _ = WriteVarint(b, uint64(msgLen))
	b, _ = WriteVarint(b, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		b, _ = WriteVarint(b, uint64(version))
	}

	// Append parameters
	b, _ = WriteVarint(b, uint64(len(scm.Parameters)))
	for key, value := range scm.Parameters {
		b, _ = WriteVarint(b, key)
		b, _ = WriteBytes(b, value)
	}

	_, err := w.Write(b)
	return err
}

func (scm *SessionClientMessage) Decode(src io.Reader) error {
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	count, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	b = b[n:]

	scm.SupportedVersions = make([]protocol.Version, count)
	for i := range count {
		num, n, err := ReadVarint(b)
		if err != nil {
			return err
		}
		scm.SupportedVersions[i] = protocol.Version(num)
		b = b[n:]
	}

	scm.Parameters, n, err = ReadParameters(b)
	if err != nil {
		return err
	}
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}

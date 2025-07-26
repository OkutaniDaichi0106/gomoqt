package message

import (
	"bytes"
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
	b := pool.Get(msgLen)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(len(scm.SupportedVersions)))
	for _, version := range scm.SupportedVersions {
		b = quicvarint.Append(b, uint64(version))
	}

	// Append parameters
	b = quicvarint.Append(b, uint64(len(scm.Parameters)))
	for key, value := range scm.Parameters {
		b = quicvarint.Append(b, key)
		b = quicvarint.Append(b, uint64(len(value)))
		b = append(b, value...)
	}

	_, err := w.Write(b)
	pool.Put(b)
	return err
}

func (scm *SessionClientMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	_, err = io.ReadFull(src, b)
	if err != nil {
		pool.Put(b)
		return err
	}

	r := bytes.NewReader(b)

	count, err := ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}

	scm.SupportedVersions = make([]protocol.Version, count)
	for i := range count {
		num, err = ReadVarint(r)
		if err != nil {
			pool.Put(b)
			return err
		}
		scm.SupportedVersions[i] = protocol.Version(num)
	}

	count, err = ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}

	scm.Parameters = make(Parameters, count)
	for range count {
		key, err := ReadVarint(r)
		if err != nil {
			pool.Put(b)
			return err
		}

		value, err := ReadBytes(r)
		if err != nil {
			pool.Put(b)
			return err
		}

		scm.Parameters[key] = value
	}

	pool.Put(b)

	return nil
}

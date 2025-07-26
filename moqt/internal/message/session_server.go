package message

import (
	"bytes"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SessionServerMessage struct {
	/*
	 * Versions selected by the server
	 */
	SelectedVersion protocol.Version

	/*
	 * Setup Parameters
	 * Keys of the maps should not be duplicated
	 */
	Parameters Parameters
}

func (ssm SessionServerMessage) Len() int {
	var l int

	l += VarintLen(uint64(ssm.SelectedVersion))
	l += ParametersLen(ssm.Parameters)

	return l
}

func (ssm SessionServerMessage) Encode(w io.Writer) error {
	msgLen := ssm.Len()
	b := pool.Get(msgLen)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(ssm.SelectedVersion))

	// Append parameters
	b = quicvarint.Append(b, uint64(len(ssm.Parameters)))
	for key, value := range ssm.Parameters {
		b = quicvarint.Append(b, key)
		b = quicvarint.Append(b, uint64(len(value)))
		b = append(b, value...)
	}

	_, err := w.Write(b)
	pool.Put(b)
	return err
}

func (ssm *SessionServerMessage) Decode(src io.Reader) error {
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

	num, err = ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}

	ssm.SelectedVersion = protocol.Version(num)

	count, err := ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}

	ssm.Parameters = make(Parameters, count)
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

		ssm.Parameters[key] = value
	}

	pool.Put(b)

	return nil
}

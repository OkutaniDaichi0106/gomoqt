package message

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
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

	// Allocate buffer for whole message
	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
	defer pool.Put(b)

	b, _ = WriteVarint(b, uint64(msgLen))
	b, _ = WriteVarint(b, uint64(ssm.SelectedVersion))

	// Append parameters
	b, _ = WriteVarint(b, uint64(len(ssm.Parameters)))
	for key, value := range ssm.Parameters {
		b, _ = WriteVarint(b, key)
		b, _ = WriteBytes(b, value)
	}

	_, err := w.Write(b)
	return err
}

func (ssm *SessionServerMessage) Decode(src io.Reader) error {
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

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	ssm.SelectedVersion = protocol.Version(num)
	b = b[n:]

	ssm.Parameters, n, err = ReadParameters(b)
	if err != nil {
		return err
	}
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}

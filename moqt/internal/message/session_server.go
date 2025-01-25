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

func (ssm SessionServerMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := make([]byte, 0, 1<<4)

	p = appendNumber(p, uint64(ssm.SelectedVersion))
	p = appendParameters(p, ssm.Parameters)

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	return w.Write(b)
}

func (ssm *SessionServerMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	// Read selected version
	version, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	ssm.SelectedVersion = protocol.Version(version)

	// Read parameters
	ssm.Parameters, _, err = readParameters(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}

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
	l := 0
	l += numberLen(uint64(ssm.SelectedVersion))
	l += parametersLen(ssm.Parameters)
	return l
}

func (ssm SessionServerMessage) Encode(w io.Writer) (int, error) {

	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(ssm.Len()))

	p = AppendNumber(p, uint64(ssm.SelectedVersion))
	p = AppendParameters(p, ssm.Parameters)

	return w.Write(p)
}

func (ssm *SessionServerMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	version, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	ssm.SelectedVersion = protocol.Version(version)

	ssm.Parameters, _, err = ReadParameters(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}

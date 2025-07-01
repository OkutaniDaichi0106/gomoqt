package message

import (
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

func (ssm SessionServerMessage) Encode(w io.Writer) error {

	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(ssm.SelectedVersion))
	p = AppendParameters(p, ssm.Parameters)

	_, err := w.Write(p)
	return err
}

func (ssm *SessionServerMessage) Decode(r io.Reader) error {
	version, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	ssm.SelectedVersion = protocol.Version(version)

	ssm.Parameters, _, err = ReadParameters(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	return nil
}

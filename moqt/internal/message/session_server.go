package message

import (
	"bytes"
	"io"
	"log/slog"

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
	slog.Debug("encoding a SESSION_SERVER message")

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(ssm.Len()))

	p = AppendNumber(p, uint64(ssm.SelectedVersion))
	p = AppendParameters(p, ssm.Parameters)

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write a SESSION_SERVER message", "error", err)
		return n, err
	}

	return n, nil
}

func (ssm *SessionServerMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for SESSION_SERVER message", "error", err)
		return n, err
	}

	mr := bytes.NewReader(buf)

	version, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read a selected version", "error", err)
		return n, err
	}
	ssm.SelectedVersion = protocol.Version(version)

	ssm.Parameters, _, err = ReadParameters(mr)
	if err != nil {
		slog.Error("failed to read parameters", "error", err)
		return n, err
	}

	return n, nil
}

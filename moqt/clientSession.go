package moqt

import (
	"context"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type ClientSession struct {
	*session
}

func (sess *ClientSession) OpenDataStream(g Group) (SendStream, error) {
	return sess.openDataStream(g)
}

func (sess ClientSession) AcceptDataStream(ctx context.Context) (Group, ReceiveStream, error) {
	stream, err := sess.conn.AcceptUniStream(ctx)
	if err != nil {
		slog.Error("failed to accept an unidirectional stream", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	group, err := getGroup(quicvarint.NewReader(stream))
	if err != nil {
		slog.Error("failed to get a Group", slog.String("error", err.Error()))
		return Group{}, nil, err
	}

	return group, stream, nil
}

func (sess *ClientSession) SendDatagram(g Group, data []byte) error {
	return sess.sendDatagram(g, data)
}

func (sess ClientSession) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return sess.conn.ReceiveDatagram(ctx)
}

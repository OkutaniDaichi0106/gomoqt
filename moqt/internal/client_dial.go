package internal

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

func SetupWebTransport(ctx context.Context, sess *webtransport.Session, params message.Parameters) (*Session, message.SessionServerMessage, error) {
	slog.Debug("dialing to the server with webtransport")

	// Get a moqt.Connection
	conn := transport.NewMOWTConnection(sess)

	csm := message.SessionClientMessage{
		SupportedVersions: DefaultClientVersions,
		Parameters:        params,
	}

	// Open a session stream
	sessstr, ssm, err := OpenSessionStream(conn, csm)
	if err != nil {
		slog.Error("failed to open a session stream", slog.String("error", err.Error()))
		return nil, message.SessionServerMessage{}, err
	}

	return NewSession(conn, sessstr), ssm, nil

}

func SetupQUIC(ctx context.Context, qconn quic.Connection, params message.Parameters) (*Session, message.SessionServerMessage, error) {
	slog.Debug("dialing to the server with quic")

	// Get a moqt.Connection
	conn := transport.NewMORQConnection(qconn)

	csm := message.SessionClientMessage{
		SupportedVersions: DefaultClientVersions,
		Parameters:        params,
	}

	// Open a session stream
	sessstr, ssm, err := OpenSessionStream(conn, csm)
	if err != nil {
		slog.Error("failed to open a session stream", slog.String("error", err.Error()))
		return nil, message.SessionServerMessage{}, err
	}

	return NewSession(conn, sessstr), ssm, nil
}

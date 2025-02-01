package internal

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type SessionStream struct {
	Stream transport.Stream
}

func (ss *SessionStream) UpdateSession(bitrate uint64) error {
	sum := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}

	_, err := sum.Encode(ss.Stream)
	if err != nil {
		return err
	}

	return nil
}

func openSessionStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("opening a session stream")

	/***/
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_session,
	}

	_, err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Debug("opened a session stream")

	return stream, nil
}

func acceptSessionStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("accepting a session stream")

	// Accept a Bidirectional Stream, which must be a Sesson Stream
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		slog.Error("failed to accept a stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Get a Stream Type message
	var stm message.StreamTypeMessage
	_, err = stm.Decode(stream)
	if err != nil {
		slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
		return nil, err
	}

	// Verify if the Stream is the Session Stream
	if stm.StreamType != stream_type_session {
		slog.Error("unexpected Stream Type ID", slog.Any("ID", stm.StreamType))
		return nil, err
	}

	slog.Debug("accepted a session stream")

	return stream, nil
}

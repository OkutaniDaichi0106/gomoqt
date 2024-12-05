package moqt

import (
	"context"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type ServerSession struct {
	//relayManager *RelayManager
	*session
}

func (svrsess *ServerSession) init(conn moq.Connection) error {
	sess := session{
		conn:               conn,
		subscribeSenders:   make(map[SubscribeID]*SubscribeSender),
		subscribeReceivers: make(map[SubscribeID]*SubscribeReceiver),
	}

	/*
	 * Accept a Session Stream
	 */
	slog.Debug("session stream was opened")

	// Accept a bidirectional Stream for the Sesson Stream
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		slog.Error("failed to open a stream", slog.String("error", err.Error()))
		return err
	}

	// Read the first byte and get Stream Type
	var stm message.StreamTypeMessage
	err = stm.Decode(stream)
	if err != nil {
		slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
		return err
	}

	// Verify if the Stream is the Session Stream
	if stm.StreamType != stream_type_session {
		slog.Error("unexpected Stream Type ID", slog.Any("ID", stm.StreamType))
		return err
	}

	sess.stream = stream

	/*
	 *
	 */
	*svrsess = ServerSession{
		session: &sess,
	}

	return nil
}

func (sess *ServerSession) GoAway(uri string, timeout time.Duration) {
	gam := message.GoAwayMessage{
		NewSessionURI: uri,
	}

	err := gam.Encode(sess.stream)
	if err != nil {
		slog.Error("failed to send GO_AWAY message", slog.String("error", err.Error()))
		return
	}

	// Wait during the given time
	time.Sleep(timeout)

	//
	if len(sess.subscribeReceivers) != 0 {
		slog.Info("subscription is still on the Session")

		// Terminate the Session with an GO_AWAY_TIMEOUT error
		sess.Terminate(ErrGoAwayTimeout)

		return
	}
}

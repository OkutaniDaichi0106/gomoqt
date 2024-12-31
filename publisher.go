package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type Publisher interface {
	AcceptInterest(context.Context) (*SendAnnounceStream, error)

	AcceptSubscription(context.Context) (*ReceivedSubscribeStream, error)

	AcceptFetch(context.Context) (*ReceivedFetch, error)

	AcceptInfoRequest(context.Context) (*SendInfoStream, error)
}

func openGroupStream(conn transport.Connection) (transport.SendStream, error) {
	slog.Debug("opening an Group Stream")

	stream, err := conn.OpenUniStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_group,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func sendDatagram(conn transport.Connection, g sentGroup, payload []byte) error {
	if g.groupSequence == 0 {
		return errors.New("0 sequence number")
	}

	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(g.subscribeID),
		GroupSequence: message.GroupSequence(g.groupSequence),
		GroupPriority: message.GroupPriority(g.groupPriority),
	}

	var buf bytes.Buffer

	// Encode the GROUP message
	err := gm.Encode(&buf)
	if err != nil {
		slog.Error("failed to encode a GROUP message", slog.String("error", err.Error()))
		return err
	}

	// Encode the payload
	_, err = buf.Write(payload)
	if err != nil {
		slog.Error("failed to encode a payload", slog.String("error", err.Error()))
		return err
	}

	// Send the data with the GROUP message
	err = conn.SendDatagram(buf.Bytes())
	if err != nil {
		slog.Error("failed to send a datagram", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func newReceivedInterest(stream transport.Stream) (*SendAnnounceStream, error) {
	// Get an Interest
	interest, err := readInterest(stream)
	if err != nil {
		slog.Error("failed to get an Interest", slog.String("error", err.Error()))
		return nil, err
	}

	return &SendAnnounceStream{
		Interest:     interest,
		activeTracks: make(map[string]Track),
		stream:       stream,
	}, nil
}

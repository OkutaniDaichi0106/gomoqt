package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type publisher interface {
	AcceptInterest(context.Context) (*ReceivedInterest, error)

	AcceptSubscription(context.Context) (*ReceivedSubscription, error)

	AcceptFetch(context.Context) (*ReceivedFetch, error)

	AcceptInfoRequest(context.Context) (*ReceivedInfoRequest, error)
}

var _ publisher = (*Publisher)(nil)

type Publisher struct {
	sess *session

	/*
	 *
	 */
	*publisherManager
}

func (p *Publisher) AcceptInterest(ctx context.Context) (*ReceivedInterest, error) {
	for {
		if p.receivedSubscriptionQueue.Len() != 0 {
			return p.receivedInterestQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.receivedInterestQueue.Chan():
		}
	}
}

func (p *Publisher) AcceptSubscription(ctx context.Context) (*ReceivedSubscription, error) {
	for {
		if p.receivedSubscriptionQueue.Len() != 0 {
			stream := p.receivedSubscriptionQueue.Dequeue()

			// Set the Connection
			stream.conn = p.sess.conn

			return stream, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.receivedInterestQueue.Chan():
		}
	}
}

func (p *Publisher) AcceptFetch(ctx context.Context) (*ReceivedFetch, error) {
	for {
		if p.receivedFetchQueue.Len() != 0 {
			return p.receivedFetchQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.receivedFetchQueue.Chan():
		}
	}
}

func (p *Publisher) AcceptInfoRequest(ctx context.Context) (*ReceivedInfoRequest, error) {
	for {
		if p.receivedInfoRequestQueue.Len() != 0 {
			return p.receivedInfoRequestQueue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-p.receivedInfoRequestQueue.Chan():
		}
	}
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

func newReceivedInterest(stream transport.Stream) (*ReceivedInterest, error) {
	// Get an Interest
	interest, err := readInterest(stream)
	if err != nil {
		slog.Error("failed to get an Interest", slog.String("error", err.Error()))
		return nil, err
	}

	return &ReceivedInterest{
		Interest:     interest,
		activeTracks: make(map[string]Track),
		stream:       stream,
	}, nil
}

package moqt

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type publisher interface {
	StartTrack(Track) error
	EndTrack(Track) error

	AcceptInterest(context.Context) (*ReceivedInterest, error)

	AcceptSubscription(context.Context) (*ReceivedSubscription, error)

	AcceptFetch(context.Context) (*ReceivedFetch, error)

	OpenDataStream(SubscribeID, GroupSequence, GroupPriority) (DataSendStream, error)
}

var _ publisher = (*Publisher)(nil)

type Publisher struct {
	sess *session

	/*
	 *
	 */
	*publisherManager
}

func (p *Publisher) StartTrack(t Track) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.tracks[t.TrackPath]
	if !ok {
		return ErrDuplicatedTrackPath
	}

	p.tracks[t.TrackPath] = t

	return nil
}

func (p *Publisher) EndTrack(t Track) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.tracks[t.TrackPath]
	if !ok {
		return ErrTrackDoesNotExist
	}

	delete(p.tracks, t.TrackPath)

	return nil
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
			return p.receivedSubscriptionQueue.Dequeue(), nil
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

func (p *Publisher) OpenDataStream(id SubscribeID, sequence GroupSequence, priority GroupPriority) (DataSendStream, error) {
	// Verify
	if sequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	// Open
	stream, err := p.openGroupStream()
	if err != nil {
		slog.Error("failed to open a group stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Send the GROUP message
	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(id),
		GroupSequence: message.GroupSequence(sequence),
		GroupPriority: message.GroupPriority(priority),
	}
	err = gm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return dataSendStream{
			SendStream: stream,
			SentGroup: SentGroup{
				subscribeID:   id,
				groupSequence: sequence,
				groupPriority: priority,
				sentAt:        time.Now(),
			},
		},
		nil
}

func (p *Publisher) openGroupStream() (transport.SendStream, error) {
	slog.Debug("opening an Group Stream")

	stream, err := p.sess.conn.OpenUniStream()
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

func (p *Publisher) sendDatagram(g SentGroup, payload []byte) error {
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
	err = p.sess.conn.SendDatagram(buf.Bytes())
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
		Interest: interest,
		active:   make(map[string]Track),
		stream:   stream,
	}, nil
}

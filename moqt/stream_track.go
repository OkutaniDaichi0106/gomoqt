package moqt

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

// type StreamTrackWriter interface {
// 	Close() error
// 	CloseWithError(error) error
// 	SubscribeConfig() SubscribeConfig
// 	// AcceptFetchStream() ReceiveFetchStream
// 	OpenGroupStream(GroupSequence, GroupPriority) (SendGroupStream, error)
// }

type TrackSender interface {
	ReceiveSubscribeStream
	OpenGroupSender(GroupSequence, GroupPriority) (SendGroupStream, error)
}

var _ TrackSender = (*streamTrackSender)(nil)

type streamTrackSender struct {
	ReceiveSubscribeStream
	conn transport.Connection
}

func (tw streamTrackSender) OpenGroupSender(sequence GroupSequence, priority GroupPriority) (SendGroupStream, error) {
	// Verify
	if sequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	// Open
	stream, err := openGroupStream(tw.conn)
	if err != nil {
		slog.Error("failed to open a group stream", slog.String("error", err.Error()))
		return nil, err
	}

	group := group{
		groupSequence: sequence,
		groupPriority: priority,
	}

	// Send the GROUP message
	err = writeGroup(stream, tw.SubscribeID(), group)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return sendGroupStream{
			stream:    stream,
			Group:     group,
			startTime: time.Now(),
		},
		nil
}

type TrackReceiver interface {
	SendSubscribeStream
	AcceptGroupReceiver(context.Context) (ReceiveGroupStream, error)
}

var _ TrackReceiver = (*streamTrackReceiver)(nil)

type streamTrackReceiver struct {
	SendSubscribeStream
	queue *groupReceiverQueue
}

func (tw streamTrackReceiver) AcceptGroupReceiver(ctx context.Context) (ReceiveGroupStream, error) {
	slog.Debug("accepting a data stream")

	for {
		if tw.queue.Len() > 0 {
			return tw.queue.Dequeue(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tw.queue.Chan():
		default:
		}
	}
}

package moqt

import (
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeStream Stream
type SubscribeID uint64
type GroupOrder byte

type Subscription struct {
	SubscribeID        SubscribeID
	TrackNamespace     []string
	TrackName          string
	Parameters         Parameters
	SubscriberPriority SubscriberPriority
	GroupOrder         GroupOrder
	GroupExpires       time.Duration
	MinGroupSequence   uint64
	MaxGroupSequence   uint64
}

/*
 *
 */
type SubscribeWriter interface {
	Subscribe(Subscription) error
}

/*
 *
 */

var _ SubscribeWriter = (*defaultSubscribeWriter)(nil)

type defaultSubscribeWriter struct {
	subscribeID SubscribeID
	stream      SubscribeStream
}

func (w defaultSubscribeWriter) Subscribe(subscription Subscription) error {
	w.subscribeID++
	sm := message.SubscribeMessage{
		SubscribeID:        message.SubscribeID(w.subscribeID),
		TrackNamespace:     subscription.TrackNamespace,
		TrackName:          subscription.TrackName,
		SubscriberPriority: message.SubscriberPriority(subscription.SubscriberPriority),
		GroupOrder:         message.GroupOrder(subscription.GroupOrder),
		MinGroupSequence:   subscription.MinGroupSequence,
		MaxGroupSequence:   subscription.MaxGroupSequence,
		Parameters:         message.Parameters(subscription.Parameters),
	}

	subAttr := slog.Group("subscription",
		slog.Uint64("subscribe ID", uint64(sm.SubscribeID)),
		slog.Any("track namespace", sm.TrackNamespace),
		slog.String("track name", sm.TrackName),
		slog.Uint64("track namespace", uint64(sm.SubscriberPriority)),
		slog.Uint64("group order", uint64(sm.GroupOrder)),
		slog.Uint64("min group sequence", sm.MinGroupSequence),
		slog.Uint64("max group sequence", sm.MaxGroupSequence),
		slog.Any("parameters", sm.Parameters),
	)

	_, err := w.stream.Write(sm.SerializePayload())
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()), subAttr)
		return err
	}

	slog.Info("subscribed", subAttr)

	return nil
}

type SubscribeResponceWriter interface {
	Accept()
	Reject(SubscribeError)
}

type SubscribeHandler interface {
	HandleSubscribe(Subscription, SubscribeResponceWriter)
}

var _ SubscribeResponceWriter = (*defaultSubscribeResponceWriter)(nil)

type defaultSubscribeResponceWriter struct {
	errCh  chan error
	stream Stream
}

func (w defaultSubscribeResponceWriter) Accept() {
	slog.Info("accepted a subscription")
	w.errCh <- nil
}

func (w defaultSubscribeResponceWriter) Reject(err SubscribeError) {
	slog.Info("rejected a subscription")

	// Cancel
	w.stream.CancelRead(StreamErrorCode(err.SubscribeErrorCode()))
	w.stream.CancelWrite(StreamErrorCode(err.SubscribeErrorCode()))

	w.errCh <- err
}

func getSubscription(r quicvarint.Reader) (Subscription, error) {
	var sm message.SubscribeMessage
	err := sm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	return Subscription{
		SubscribeID:        SubscribeID(sm.SubscribeID),
		TrackNamespace:     sm.TrackNamespace,
		TrackName:          sm.TrackName,
		SubscriberPriority: SubscriberPriority(sm.SubscriberPriority),
		GroupOrder:         GroupOrder(sm.GroupOrder),
		MinGroupSequence:   sm.MinGroupSequence,
		MaxGroupSequence:   sm.MaxGroupSequence,
		Parameters:         Parameters(sm.Parameters),
	}, nil
}

func getSubscribeUpdate(old Subscription, r quicvarint.Reader) (Subscription, error) {
	var sum message.SubscribeUpdateMessage
	err := sum.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	new := Subscription{
		SubscribeID:        old.SubscribeID,
		TrackNamespace:     old.TrackNamespace,
		TrackName:          old.TrackName,
		Parameters:         Parameters(sum.Parameters),
		SubscriberPriority: SubscriberPriority(sum.SubscriberPriority),
		GroupOrder:         GroupOrder(sum.GroupOrder),
		GroupExpires:       sum.GroupExpires,
	}

	if sum.MinGroupSequence != 0 {
		if old.MinGroupSequence > sum.MinGroupSequence {
			return Subscription{}, errors.New("minimum group sequence smaller than the prior one was specified")
		}

		new.MinGroupSequence = sum.MinGroupSequence
	}

	if sum.MaxGroupSequence != 0 {
		if old.MaxGroupSequence < sum.MaxGroupSequence {
			return Subscription{}, errors.New("maximum group sequence larger than the prior one was specified")
		}

		new.MaxGroupSequence = sum.MaxGroupSequence
	}

	slog.Info("a subscription was updated", slog.Any("from", old), slog.Any("to", new))

	return new, nil
}

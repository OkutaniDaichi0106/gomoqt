package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeStream Stream
type SubscribeID uint64
type GroupOrder byte

type Subscription struct {
	SubscribeID        SubscribeID
	Announcement       Announcement
	TrackName          string
	Parameters         message.Parameters
	SubscriberPriority SubscriberPriority
	GroupOrder         GroupOrder
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
type SubscribeResponceWriter interface {
	Accept()
	Reject(SubscribeError)
}

type SubscribeHandler interface {
	HandleSubscribe(Subscription, SubscribeResponceWriter)
}

var _ SubscribeWriter = (*defaultSubscribeWriter)(nil)

type defaultSubscribeWriter struct {
	subscribeID SubscribeID
	stream      SubscribeStream
}

func (w defaultSubscribeWriter) Subscribe(subscription Subscription) error {
	w.subscribeID++
	sm := message.SubscribeMessage{
		SubscribeID:        message.SubscribeID(w.subscribeID),
		TrackNamespace:     subscription.Announcement.TrackNamespace,
		TrackName:          subscription.TrackName,
		SubscriberPriority: message.SubscriberPriority(subscription.SubscriberPriority),
		GroupOrder:         message.GroupOrder(subscription.GroupOrder),
		MinGroupSequence:   subscription.MinGroupSequence,
		MaxGroupSequence:   subscription.MaxGroupSequence,
		Parameters:         subscription.Parameters,
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

var _ SubscribeResponceWriter = (*defaultSubscribeResponceWriter)(nil)

type defaultSubscribeResponceWriter struct {
	stream Stream
}

func (defaultSubscribeResponceWriter) Accept() {

}

func (srw defaultSubscribeResponceWriter) Reject(err SubscribeError) {
	srw.stream.CancelRead(StreamErrorCode(err.SubscribeErrorCode()))
	srw.stream.CancelWrite(StreamErrorCode(err.SubscribeErrorCode()))
}

func getSubscription(r quicvarint.Reader) (Subscription, error) {
	var sm message.SubscribeMessage
	err := sm.DeserializePayload(r)
	if err != nil {
		slog.Error("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	return Subscription{
		SubscribeID: SubscribeID(sm.SubscribeID),
		Announcement: Announcement{
			TrackNamespace: sm.TrackNamespace,
		},
		TrackName:          sm.TrackName,
		SubscriberPriority: SubscriberPriority(sm.SubscriberPriority),
		GroupOrder:         GroupOrder(sm.GroupOrder),
		MinGroupSequence:   sm.MinGroupSequence,
		MaxGroupSequence:   sm.MaxGroupSequence,
		Parameters:         sm.Parameters,
	}, nil
}

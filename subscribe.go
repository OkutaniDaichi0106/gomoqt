package moqt

import (
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64

type GroupOrder byte

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

type Subscription struct {
	subscribeID        SubscribeID
	TrackNamespace     string
	TrackName          string
	Parameters         Parameters
	SubscriberPriority SubscriberPriority
	GroupOrder         GroupOrder
	GroupExpires       time.Duration
	MinGroupSequence   GroupSequence
	MaxGroupSequence   GroupSequence
}

func (s Subscription) FirstGrouopSequence() GroupSequence {
	switch s.GroupOrder {
	case ASCENDING, DEFAULT:
		return s.MinGroupSequence
	case DESCENDING:
		return s.MaxGroupSequence
	default:
		return 0
	}
}

func (s Subscription) GetGroup(seq GroupSequence, priority PublisherPriority) Group {
	return Group{
		subscribeID:       s.subscribeID,
		groupSequence:     seq,
		PublisherPriority: priority,
	}
}

type SubscribeWriter struct {
	reader       quicvarint.Reader
	stream       Stream
	subscription Subscription
	mu           sync.RWMutex
}

func (s *SubscribeWriter) Update(subscription Subscription) (Info, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.subscription
	slog.Debug("trying to update", slog.Any("from", old), slog.Any("to", subscription))

	// Verify if the new group range is valid
	if subscription.MinGroupSequence > subscription.MaxGroupSequence {
		slog.Debug("MinGroupSequence is larger than MaxGroupSequence")
		return Info{}, ErrInvalidRange
	}
	if old.MinGroupSequence > subscription.MinGroupSequence {
		slog.Debug("the new MinGroupSequence is smaller than the old MinGroupSequence")
		return Info{}, ErrInvalidRange
	}
	if old.MaxGroupSequence < subscription.MaxGroupSequence {
		slog.Debug("the new MaxGroupSequence is larger than the old MaxGroupSequence")
		return Info{}, ErrInvalidRange
	}

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	sum := message.SubscribeUpdateMessage{
		SubscribeID:        message.SubscribeID(s.subscription.subscribeID),
		SubscriberPriority: message.SubscriberPriority(subscription.SubscriberPriority),
		GroupOrder:         message.GroupOrder(subscription.GroupOrder),
		GroupExpires:       subscription.GroupExpires,
		MinGroupSequence:   message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence:   message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:         message.Parameters(subscription.Parameters),
	}

	_, err := s.stream.Write(sum.SerializePayload())
	if err != nil {
		slog.Debug("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return Info{}, err
	}

	/*
	 * Receive an INFO message
	 */
	if s.reader == nil {
		s.reader = quicvarint.NewReader(s.stream)
	}

	info, err := getInfo(s.reader)
	if err != nil {
		slog.Debug("failed to get an Info")
		return Info{}, err
	}

	return info, nil
}

func (s *SubscribeWriter) Unsubscribe(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil {
		err := s.stream.Close()
		if err != nil {
			slog.Error("failed to close a Subscribe Stream", slog.String("error", err.Error()))
		}
		return
	}

	suberr, ok := err.(SubscribeError)
	if !ok {
		suberr = ErrInternalError
	}

	s.stream.CancelWrite(StreamErrorCode(suberr.SubscribeErrorCode()))
	s.stream.CancelRead(StreamErrorCode(suberr.SubscribeErrorCode()))
}

func (s *SubscribeWriter) Subscription() Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.subscription
}

/*
 *
 */

type SubscribeHandler interface {
	HandleSubscribe(Subscription, *Info, SubscribeResponceWriter)
}

type SubscribeResponceWriter struct {
	doneCh chan struct{}
	stream Stream
}

func (w SubscribeResponceWriter) Accept(i Info) {
	slog.Info("accepted a subscription")

	w.doneCh <- struct{}{}

	close(w.doneCh)
}

func (w SubscribeResponceWriter) Reject(err error) {
	if err == nil {
		err := w.stream.Close()
		if err != nil {
			slog.Debug("failed to close a Subscribe Stream gracefully", slog.String("error", err.Error()))
		}

		return
	}

	var code StreamErrorCode

	var strerr StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		suberr, ok := err.(SubscribeError)
		if ok {
			code = StreamErrorCode(suberr.SubscribeErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	w.stream.CancelRead(code)
	w.stream.CancelWrite(code)

	w.doneCh <- struct{}{}

	close(w.doneCh)

	slog.Debug("rejected a subscription", slog.String("error", err.Error()))
}

func getSubscription(r quicvarint.Reader) (Subscription, error) {
	var sm message.SubscribeMessage
	err := sm.DeserializePayload(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	return Subscription{
		subscribeID:        SubscribeID(sm.SubscribeID),
		TrackNamespace:     strings.Join(sm.TrackNamespace, "/"),
		TrackName:          sm.TrackName,
		SubscriberPriority: SubscriberPriority(sm.SubscriberPriority),
		GroupOrder:         GroupOrder(sm.GroupOrder),
		MinGroupSequence:   GroupSequence(sm.MinGroupSequence),
		MaxGroupSequence:   GroupSequence(sm.MaxGroupSequence),
		Parameters:         Parameters(sm.Parameters),
	}, nil
}

func getSubscribeUpdate(old Subscription, r quicvarint.Reader) (Subscription, error) {
	var sum message.SubscribeUpdateMessage
	err := sum.DeserializePayload(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	new := Subscription{
		subscribeID:        old.subscribeID,
		TrackNamespace:     old.TrackNamespace,
		TrackName:          old.TrackName,
		Parameters:         Parameters(sum.Parameters),
		SubscriberPriority: SubscriberPriority(sum.SubscriberPriority),
		GroupOrder:         GroupOrder(sum.GroupOrder),
		GroupExpires:       sum.GroupExpires,
	}

	return new, nil
}

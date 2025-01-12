package moqtransfork

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SendSubscribeStream interface {
	// Get the SubscribeID
	SubscribeID() SubscribeID

	// Get the subscription
	Subscription() Subscription

	// Update the subscription
	UpdateSubscribe(SubscribeUpdate) error

	// Close the stream
	Close() error

	// Close the stream with an error
	CloseWithError(error) error
}

var _ SendSubscribeStream = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	subscribeID  SubscribeID
	subscription Subscription
	stream       transport.Stream
	mu           sync.Mutex
}

func (ss *sendSubscribeStream) SubscribeID() SubscribeID {
	return ss.subscribeID
}

func (ss *sendSubscribeStream) Subscription() Subscription {
	return ss.subscription
}

func (sss *sendSubscribeStream) UpdateSubscribe(update SubscribeUpdate) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	subscription, err := updateSubscription(sss.subscription, update)
	if err != nil {
		slog.Error("failed to update a subscription", slog.String("error", err.Error()))
		return err
	}

	err = writeSubscribeUpdate(sss.stream, update)
	if err != nil {
		slog.Error("failed to write a subscribe update message", slog.String("error", err.Error()))
		return err
	}

	sss.subscription = subscription

	slog.Debug("updated a subscription", slog.Any("subscription", sss.subscription))

	return nil
}

func (ss *sendSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe send stream", slog.Any("subscription", ss.subscription))

	err := ss.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe send stream", slog.Any("subscription", ss.subscription))

	return nil
}

func (sss *sendSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe send stream", slog.Any("subscription", sss.subscription))

	if err == nil {
		return sss.Close()
	}
	// TODO:

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	sss.stream.CancelRead(code)
	sss.stream.CancelWrite(code)

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", sss.Subscription()))

	return nil
}

type ReceiveSubscribeStream interface {
	SubscribeID() SubscribeID
	Subscription() Subscription
	CountDataGap(GroupSequence, uint64, uint64) error
	CloseWithError(error) error
	Close() error
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	subscribeID  SubscribeID
	subscription Subscription
	stream       transport.Stream
	mu           sync.Mutex
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.subscribeID
}

func (rss *receiveSubscribeStream) Subscription() Subscription {
	return rss.subscription
}

func (rs *receiveSubscribeStream) CountDataGap(start GroupSequence, count uint64, code uint64) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// TODO: Implement
	sgm := message.SubscribeGapMessage{
		GroupStartSequence: message.GroupSequence(start),
		Count:              count,
		GroupErrorCode:     message.GroupErrorCode(code),
	}
	err := sgm.Encode(rs.stream)
	if err != nil {
		slog.Error("failed to encode SUBSCRIBE_GAP message")
		return err
	}

	return nil
}

func (srs *receiveSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	if err == nil {
		return srs.Close()
	}

	// TODO:

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	srs.stream.CancelRead(code)
	srs.stream.CancelWrite(code)

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	return nil
}

func (srs *receiveSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	err := srs.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	return nil
}

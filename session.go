package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type SessionStream moq.Stream //TODO:

type session struct {
	conn   moq.Connection
	stream SessionStream

	publisherManager *publisherManager

	subscriberManager *subscribeManager
}

func (sess *session) Publisher() Publisher {
	return &publisher{
		sess:    sess,
		manager: sess.publisherManager,
	}
}

func (sess *session) Subscriber() Subscriber {
	return &subscriber{
		sess:             sess,
		subscribeManager: sess.subscriberManager,
	}
}

func (sess *session) Terminate(err error) {
	slog.Info("Terminating a session", slog.String("reason", err.Error()))

	var tererr TerminateError

	if err == nil {
		tererr = NoErrTerminate
	} else {
		var ok bool
		tererr, ok = err.(TerminateError)
		if !ok {
			tererr = ErrInternalError
		}
	}

	err = sess.conn.CloseWithError(moq.SessionErrorCode(tererr.TerminateErrorCode()), err.Error())
	if err != nil {
		slog.Error("failed to close the Connection", slog.String("error", err.Error()))
		return
	}

	slog.Info("Terminated a session")
}

// func (sess *session) acceptNewSubscription(sr *SubscribeReceiver) {
// 	sess.mu.Lock()
// 	defer sess.mu.Unlock()

// 	// Verify if the subscription is duplicated or not
// 	_, ok := sess.subscribeReceivers[sr.subscription.subscribeID]
// 	if ok {
// 		slog.Debug("duplicated subscription", slog.Any("Subscribe ID", sr.subscription.subscribeID))
// 		return
// 	}

// 	// Register the subscription
// 	sess.subscribeReceivers[sr.subscription.subscribeID] = sr

// 	slog.Info("Accepted a new subscription", slog.Any("subscription", sr.subscription))
// }

// func (sess *session) updateSubscription(subscription SubscribeUpdate) {
// 	sess.rsMu.Lock()
// 	defer sess.rsMu.Unlock()

// 	old, ok := sess.receivedSubscriptions[subscription.TrackPath]
// 	if !ok {
// 		slog.Debug("no subscription", slog.Any("Subscribe ID", subscription.subscribeID))
// 		return
// 	}

// 	old.acceptUpdate(subscription)

// 	slog.Info("updated a subscription", slog.Any("from", old), slog.Any("to", subscription))
// }

// func (sess *session) deleteSubscription(subscription Subscription) {
// 	sess.mu.Lock()
// 	defer sess.mu.Unlock()

// 	delete(sess.subscribeReceivers, subscription.subscribeID)
// }

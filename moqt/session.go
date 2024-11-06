package moqt

import (
	"log/slog"
	"sync"
)

type Session struct {
	Connection    Connection
	SessionStream SessionStream
	//version Version

	/*
	 * Subscription
	 */
	subscription map[SubscribeID]Subscription
	subMu        sync.RWMutex

	terrCh chan TerminateError
}

func (sess *Session) Terminate(terr TerminateError) {
	err := sess.Connection.CloseWithError(SessionErrorCode(terr.TerminateErrorCode()), terr.Error())
	if err != nil {
		slog.Error("failed to close the Session", slog.String("error", err.Error()))
		return
	}

	slog.Info("closed the Session", slog.String("reason", terr.Error()))
}

func (sess *Session) addSubscription(subscription Subscription) {
	sess.subMu.Lock()
	defer sess.subMu.Unlock()

	if old, ok := sess.subscription[subscription.SubscribeID]; ok {
		slog.Info("updated the subscription", slog.Any("from", old), slog.Any("to", subscription))
	}

	sess.subscription[subscription.SubscribeID] = subscription
}

func (sess *Session) removeSubscription(subscription Subscription) {
	sess.subMu.Lock()
	defer sess.subMu.Unlock()

	if _, ok := sess.subscription[subscription.SubscribeID]; ok {
		slog.Info("subscription not found", slog.Any("subscription", subscription))
	}

	delete(sess.subscription, subscription.SubscribeID)
}

package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SessionStream transport.Stream //TODO:

type session struct {
	conn   transport.Connection
	stream SessionStream

	//
	receivedSubscriptionQueue *receivedSubscriptionQueue

	receivedInterestQueue *receivedInterestQueue

	receivedFetchQueue *receivedFetchQueue

	receivedInfoRequestQueue *receivedInfoRequestQueue

	dataReceiveStreamQueue map[SubscribeID]*dataReceiveStreamQueue

	receivedDatagramQueue map[SubscribeID]*receivedDatagramQueue

	publisherManager  *publisherManager
	subscriberManager *subscriberManager
}

func (sess *session) Publisher() *Publisher {
	return &Publisher{
		sess:             sess,
		publisherManager: sess.publisherManager,
	}
}

func (sess *session) Subscriber() *Subscriber {
	return &Subscriber{
		sess:              sess,
		subscriberManager: sess.subscriberManager,
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

	err = sess.conn.CloseWithError(transport.SessionErrorCode(tererr.TerminateErrorCode()), err.Error())
	if err != nil {
		slog.Error("failed to close the Connection", slog.String("error", err.Error()))
		return
	}

	slog.Info("Terminated a session")
}

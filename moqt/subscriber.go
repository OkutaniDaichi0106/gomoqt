package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type Subscriber struct {
	SubscriberHandler

	RemoteTrack [][]string

	// interestCh  chan Interest
	// subscribeCh chan Subscription

	//
	// interestW InterestWriter

	//
	// announceR  AnnounceReader
	// announceRW AnnounceResponceWriter

	// subscribeW SubscribeWriter
}

type SubscriberHandler interface {
	AnnounceHandler
	InfoHandler
	GroupHander
}

func (s Subscriber) run(sess Session) {

	for {

		select {
		case interest := <-s.interestCh:
			// Open a Announce Stream
			stream, err := sess.OpenStream()
			if err != nil {
				slog.Error("failed to open a bidirectional Stream", slog.String("error", err.Error()))
				return
			}

			// Send the Announce Stream Type
			_, err = stream.Write([]byte{byte(protocol.ANNOUNCE)})
			if err != nil {
				slog.Error("failed to open an Announce Stream", slog.String("error", err.Error()))
				return
			}

			//
			err = s.interestW.Interest(interest)
			if err != nil {
				slog.Error("failed to express interest", slog.String("error", err.Error()))
				return
			}
		case subscription := <-s.subscribeCh:
			// Open a Subscribe Stream
			stream, err := sess.OpenStream()
			if err != nil {
				slog.Error("failed to open a bidirectional Stream", slog.String("error", err.Error()))
				return
			}

			// Send the Subscribe Stream Type
			_, err = stream.Write([]byte{byte(protocol.SUBSCRIBE)})
			if err != nil {
				slog.Error("failed to open a Subscribe Stream", slog.String("error", err.Error()))
				return
			}

			err = s.subscribeW.Subscribe(subscription)
		}

	}
}

func (s Subscriber) Subscribe(subscription Subscription) {
	s.subscribeCh <- subscription
}

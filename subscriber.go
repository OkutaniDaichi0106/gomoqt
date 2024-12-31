package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type Subscriber interface {
	Interest(Interest) (*ReceiveAnnounceStream, error)

	Subscribe(Subscription) (*SendSubscribeStream, error)
	// Unsubscribe(*SentSubscription)

	Fetch(Fetch) (ReceiveDataStream, error)

	RequestInfo(InfoRequest) (Info, error)
}

// func (s *Subscriber) Unsubscribe(subscription *SentSubscription) {
// 	// Close gracefully
// 	err := subscription.stream.Close()
// 	if err != nil {
// 		slog.Error("failed to close a subscribe stream", slog.String("error", err.Error()))
// 	}

// 	// Remove the subscription
// 	s.subscriberManager.removeSentSubscription(subscription.subscribeID)

// 	slog.Info("Unsubscribed")
// }

// func (s Subscriber) UnsubscribeWithError(subscription *SentSubscription, err error) {
// 	if err == nil {
// 		s.Unsubscribe(subscription)
// 		slog.Error("unsubscribe with no error")
// 		return
// 	}

// 	// Close with the error
// 	var code transport.StreamErrorCode

// 	var strerr transport.StreamError
// 	if errors.As(err, &strerr) {
// 		code = strerr.StreamErrorCode()
// 	} else {
// 		var ok bool
// 		feterr, ok := err.(FetchError)
// 		if ok {
// 			code = transport.StreamErrorCode(feterr.FetchErrorCode())
// 		} else {
// 			code = ErrInternalError.StreamErrorCode()
// 		}
// 	}

// 	subscription.stream.CancelRead(code)
// 	subscription.stream.CancelWrite(code)

// 	// Remove the subscription
// 	s.subscriberManager.removeSentSubscription(subscription.subscribeID)

// 	slog.Info("Unsubscribed with an error")
// }

func openAnnounceStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("opening an Announce Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_announce,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openSubscribeStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("opening an Subscribe Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_subscribe,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openInfoStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("opening an Info Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_info,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

func openFetchStream(conn transport.Connection) (transport.Stream, error) {
	slog.Debug("opening an Fetch Stream")

	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return nil, err
	}

	stm := message.StreamTypeMessage{
		StreamType: stream_type_fetch,
	}

	err = stm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a Stream Type message", slog.String("error", err.Error()))
		return nil, err
	}

	return stream, nil
}

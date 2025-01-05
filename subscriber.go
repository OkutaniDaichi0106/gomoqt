package moqt

type Subscriber interface {
	OpenAnnounceStream(Interest) (*receiveAnnounceStream, error)

	OpenSubscribeStream(Subscription) (*sendSubscribeStream, error)
	// Unsubscribe(*SentSubscription)

	OpenFetchStream(Fetch) (ReceiveDataStream, error)

	OpenInfoStream(InfoRequest) (Info, error)
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

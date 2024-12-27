package moqt

func newPublisherManager() *publisherManager {
	return &publisherManager{
		receivedSubscriptionQueue: newReceivedSubscriptionQueue(),

		receivedInterestQueue: newReceivedInterestQueue(),

		receivedFetchQueue: newReceivedFetchQueue(),

		receivedInfoRequestQueue: newReceivedInfoRequestQueue(),

		acceptedSubscriptions: make(map[SubscribeID]*ReceivedSubscription),
	}
}

type publisherManager struct {
	//
	receivedSubscriptionQueue *receivedSubscriptionQueue

	//
	receivedInterestQueue *receivedInterestQueue

	receivedFetchQueue *receivedFetchQueue

	receivedInfoRequestQueue *receivedInfoRequestQueue

	//
	acceptedSubscriptions map[SubscribeID]*ReceivedSubscription
}

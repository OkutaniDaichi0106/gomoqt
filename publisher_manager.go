package moqt

func newPublisherManager() *publisherManager {
	return &publisherManager{
		receivedSubscriptionQueue: newReceivedSubscriptionQueue(),

		receivedInterestQueue: newReceivedInterestQueue(),

		receivedFetchQueue: newReceivedFetchQueue(),

		receivedInfoRequestQueue: newReceivedInfoRequestQueue(),
	}
}

type publisherManager struct {
	//
	receivedSubscriptionQueue *receivedSubscriptionQueue

	//
	receivedInterestQueue *receivedInterestQueue

	receivedFetchQueue *receivedFetchQueue

	receivedInfoRequestQueue *receivedInfoRequestQueue
}

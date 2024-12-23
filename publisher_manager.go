package moqt

func newPublisherManager() *publisherManager {
	return &publisherManager{
		// tracks: make(map[string]Track),
	}
}

type publisherManager struct {
	//
	receivedSubscriptionQueue receivedSubscriptionQueue

	//
	receivedInterestQueue receivedInterestQueue

	receivedFetchQueue receivedFetchQueue

	receivedInfoRequestQueue receivedInfoRequestQueue
}

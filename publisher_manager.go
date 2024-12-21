package moqt

import "sync"

type publisherManager struct {
	/*
	 * Tracks
	 * Track Path -> Track
	 */
	tracks map[string]Track
	mu     sync.RWMutex

	//
	receivedSubscriptionQueue receivedSubscriptionQueue

	//
	receivedInterestQueue receivedInterestQueue

	receivedFetchQueue receivedFetchQueue
}

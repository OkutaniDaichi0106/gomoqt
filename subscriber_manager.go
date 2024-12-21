package moqt

import "sync"

type subscriberManager struct {
	dataReceiverQueue dataReceiverQueue

	subscribeIDCounter uint64
	mu                 sync.Mutex
}

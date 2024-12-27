package moqt

import (
	"sync"
	"sync/atomic"
)

func newSubscriberManager() *subscriberManager {
	return &subscriberManager{
		sentSubscritpions:  make(map[SubscribeID]*SentSubscription),
		subscribeIDCounter: 0,
	}
}

type subscriberManager struct {
	sentSubscritpions map[SubscribeID]*SentSubscription
	ssMu              sync.Mutex

	subscribeIDCounter uint64
}

func (sm *subscriberManager) addSubscribeID() SubscribeID {
	new := atomic.AddUint64(&sm.subscribeIDCounter, 1)

	return SubscribeID(new)
}

func (sm *subscriberManager) getSubscribeID() SubscribeID {
	return SubscribeID(atomic.LoadUint64(&sm.subscribeIDCounter))
}

func (sm *subscriberManager) addSentSubscription(ss *SentSubscription) error {
	sm.ssMu.Lock()
	defer sm.ssMu.Unlock()

	if _, ok := sm.sentSubscritpions[ss.SubscribeID()]; ok {
		return ErrDuplicatedSubscribeID
	}

	sm.sentSubscritpions[ss.SubscribeID()] = ss

	return nil
}

func (sm *subscriberManager) getSentSubscription(id SubscribeID) (*SentSubscription, bool) {
	sm.ssMu.Lock()
	defer sm.ssMu.Unlock()

	subscription, ok := sm.sentSubscritpions[id]

	return subscription, ok
}

func (sm *subscriberManager) removeSentSubscription(id SubscribeID) {
	sm.ssMu.Lock()
	defer sm.ssMu.Unlock()

	delete(sm.sentSubscritpions, id)
}

package moqt

import "sync"

type subscriberManager struct {
	//dataReceiveStreamQueue dataReceiveStreamQueue

	//receivedDatagramQueue receivedDatagramQueue

	sentSubscritpions map[SubscribeID]*SentSubscription
	ssMu              sync.Mutex

	subscribeIDCounter uint64
	couterMu           sync.Mutex
}

func (sm *subscriberManager) newSubscribeID() SubscribeID {
	sm.couterMu.Lock()
	defer sm.couterMu.Unlock()

	sm.subscribeIDCounter++
	return SubscribeID(sm.subscribeIDCounter)
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

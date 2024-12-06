package moqt

import (
	"io"
	"sync"
)

type Subscriber interface {
	Interest(Interest) (AnnounceReceiver, error)
	Subscribe(Subscription) (SubscribeSender, error)
	Fetch(FetchRequest) (io.Reader, error)
	RequestInfo(InfoRequest) (Info, error)
	Terminate(error)
}

type subscriberManager struct {
	/*
	 * Sent Subscriptions
	 */
	subscribeSenders map[SubscribeID]*SubscribeSender
	ssMu             sync.RWMutex
}

func (sm *subscriberManager) addSubscribeSender(ss SubscribeSender) error {
	sm.ssMu.Lock()
	defer sm.ssMu.Unlock()

	_, ok := sm.subscribeSenders[ss.subscription.subscribeID]
	if ok {
		return ErrDuplicatedSubscribeID
	}
}

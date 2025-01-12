package moqtransfork

import (
	"sync/atomic"
)

func (sess *session) nextSubscribeID() SubscribeID {
	new := SubscribeID(atomic.LoadUint64(&sess.subscribeIDCounter))

	atomic.AddUint64(&sess.subscribeIDCounter, 1)

	return SubscribeID(new)
}

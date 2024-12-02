package moqt

import (
	"context"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type ClientSession struct {
	*session

	infos map[string]map[string]Info
	iMu   sync.RWMutex
}

func (sess *ClientSession) OpenDataStream(subscription Subscription, sequence GroupSequence, priority PublisherPriority) (moq.SendStream, error) {
	g := Group{
		subscribeID:       subscription.subscribeID,
		groupSequence:     sequence,
		PublisherPriority: priority,
	}

	return sess.openDataStream(g)
}

func (sess *ClientSession) AcceptDataStream(ctx context.Context) (Group, moq.ReceiveStream, error) {
	return sess.acceptDataStream(ctx)
}

func (sess *ClientSession) SendDatagram(subscription Subscription, sequence GroupSequence, priority PublisherPriority, data []byte) error {
	g := Group{
		subscribeID:       subscription.subscribeID,
		groupSequence:     sequence,
		PublisherPriority: priority,
	}

	return sess.sendDatagram(g, data)
}

func (sess *ClientSession) ReceiveDatagram(ctx context.Context) (Group, []byte, error) {
	return sess.receiveDatagram(ctx)
}

func (sess *ClientSession) updateInfo(trackNamespace, trackName string, info Info) {
	sess.iMu.Lock()
	defer sess.iMu.Unlock()

	sess.infos[trackNamespace][trackName] = info
}

func (sess *ClientSession) getInfo(trackNamespace, trackName string) (Info, bool) {
	sess.iMu.Lock()
	defer sess.iMu.Unlock()

	trackMap, ok := sess.infos[trackNamespace]
	if !ok {
		return Info{}, false
	}

	info, ok := trackMap[trackName]

	return info, ok
}

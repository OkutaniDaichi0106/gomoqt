package moqt

import (
	"context"
)

type ClientSession struct {
	*session
}

func (sess *ClientSession) OpenDataStream(subscription Subscription, sequence int, priority PublisherPriority) (SendStream, error) {
	g := Group{
		subscribeID:       subscription.subscribeID,
		groupSequence:     GroupSequence(sequence),
		PublisherPriority: priority,
	}

	return sess.openDataStream(g)
}

func (sess ClientSession) AcceptDataStream(ctx context.Context) (Group, ReceiveStream, error) {
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

func (sess ClientSession) ReceiveDatagram(ctx context.Context) (Group, []byte, error) {
	return sess.receiveDatagram(ctx)
}

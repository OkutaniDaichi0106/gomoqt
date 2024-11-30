package moqt

import (
	"context"

	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type ClientSession struct {
	*session
}

func (sess *ClientSession) OpenDataStream(subscription Subscription, sequence int, priority PublisherPriority) (moq.SendStream, error) {
	g := Group{
		subscribeID:       subscription.subscribeID,
		groupSequence:     GroupSequence(sequence),
		PublisherPriority: priority,
	}

	return sess.openDataStream(g)
}

func (sess ClientSession) AcceptDataStream(ctx context.Context) (Group, moq.ReceiveStream, error) {
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

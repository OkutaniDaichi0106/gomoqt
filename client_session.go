package moqt

import (
	"context"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type ClientSession struct {
	/*
	 * session
	 */
	*session

	/*
	 * Latest Track Informations
	 * The first key is the Track Namespace and the second key is the Track Name
	 */
	infos map[string]Info
	iMu   sync.RWMutex
}

func (clisess *ClientSession) init(conn moq.Connection) error {
	sess := session{
		conn:                  conn,
		subscribeWriters:      make(map[SubscribeID]*SubscribeWriter),
		receivedSubscriptions: make(map[string]Subscription),
	}

	/*
	 * Open a Session Stream
	 */
	stream, err := sess.openSessionStream()
	if err != nil {
		slog.Error("failed to open a Session Stream")
		return err
	}
	// Set the stream
	sess.stream = stream

	*clisess = ClientSession{
		session: &sess,
		infos:   make(map[string]Info),
	}

	return nil
}

func (sess *ClientSession) OpenDataStreams(trackPath string, sequence GroupSequence, priority PublisherPriority) ([]moq.SendStream, error) {
	/*
	 * Find any Subscriptions
	 */
	sess.rsMu.RLock()
	defer sess.rsMu.RUnlock()

	/*
	 *
	 */
	streams := make([]moq.SendStream, 0, 1)

	for _, subscription := range sess.receivedSubscriptions {
		g := Group{
			subscribeID:       subscription.subscribeID,
			groupSequence:     sequence,
			PublisherPriority: priority,
		}

		stream, err := sess.openDataStream(g)
		if err != nil {
			slog.Error("failed to open a data stream", slog.String("error", err.Error()))
			continue
		}

		streams = append(streams, stream)
	}

	/*
	 * Update the Track Information
	 */
	go func() {
		info, ok := sess.getInfo(trackPath)
		if !ok {
			return
		}

		// Update the Track's latest group sequence
		info.LatestGroupSequence = sequence

		sess.updateInfo(trackPath, info)
	}()

	return streams, nil
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

func (sess *ClientSession) updateInfo(trackPath string, info Info) {
	sess.iMu.Lock()
	defer sess.iMu.Unlock()

	sess.infos[trackPath] = info
}

func (sess *ClientSession) getInfo(trackPath string) (Info, bool) {
	sess.iMu.Lock()
	defer sess.iMu.Unlock()

	info, ok := sess.infos[trackPath]
	if !ok {
		return Info{}, false
	}

	return info, ok
}

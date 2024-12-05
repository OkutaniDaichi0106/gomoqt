package moqt

import (
	"context"
	"log/slog"
	"sync"
	"time"

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
		conn:               conn,
		subscribeSenders:   make(map[SubscribeID]*SubscribeSender),
		subscribeReceivers: make(map[SubscribeID]*SubscribeReceiver),
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

func (sess *ClientSession) OpenDataStreams(trackPath string, sequence GroupSequence, priority PublisherPriority, expires time.Duration) ([]moq.SendStream, error) {
	/*
	 * Find any Subscriptions
	 */
	sess.srMu.RLock()
	defer sess.srMu.RUnlock()

	/*
	 *
	 */
	streams := make([]moq.SendStream, 0, 1)

	for _, sr := range sess.subscribeReceivers {
		if sr.subscription.TrackPath == trackPath {
			g := Group{
				subscribeID:       sr.subscription.subscribeID,
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
	}

	/*
	 * Update the Track Information
	 */
	go func() {
		sess.updateInfo(trackPath, Info{
			PublisherPriority:   priority,
			LatestGroupSequence: sequence,
			GroupExpires:        expires,
		})
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

	oldInfo, ok := sess.infos[trackPath]
	if !ok {
		sess.infos[trackPath] = info
	} else {
		if info.PublisherPriority != 0 {
			info.PublisherPriority = oldInfo.PublisherPriority
		}
		if info.LatestGroupSequence != 0 {
			info.LatestGroupSequence = oldInfo.LatestGroupSequence
		}
		if info.GroupOrder != 0 {
			info.GroupOrder = oldInfo.GroupOrder
		}
		if info.GroupExpires != 0 {
			info.GroupExpires = oldInfo.GroupExpires
		}
	}
}

func (sess *ClientSession) getCurrentInfo(trackPath string) (Info, bool) {
	sess.iMu.RLock()
	defer sess.iMu.RUnlock()

	info, ok := sess.infos[trackPath]
	if !ok {
		return Info{}, false
	}

	return info, ok
}

package moqt

type clientSession interface {
	Publisher() *Publisher
	Subscriber() *Subscriber
	Terminate(error)
}

var _ clientSession = (*ClientSession)(nil)

type ClientSession struct {

	/*
	 * session
	 */
	session
}

// func (sess *clientSession) OpenDataStreams(trackPath string, sequence GroupSequence, priority GroupPriority, expires time.Duration) ([]moq.SendStream, error) {
// 	/*
// 	 * Find any Subscriptions
// 	 */
// 	sess.mu.RLock()
// 	defer sess.mu.RUnlock()

// 	/*
// 	 *
// 	 */
// 	streams := make([]moq.SendStream, 0, 1)

// 	for _, sr := range sess.subscribeReceivers {
// 		if sr.subscription.TrackPath == trackPath {
// 			g := Group{
// 				subscribeID:       sr.subscription.subscribeID,
// 				groupSequence:     sequence,
// 				GroupPriority: priority,
// 			}

// 			stream, err := sess.openDataStream(g)
// 			if err != nil {
// 				slog.Error("failed to open a data stream", slog.String("error", err.Error()))
// 				continue
// 			}

// 			streams = append(streams, stream)
// 		}
// 	}

// 	/*
// 	 * Update the Track Information
// 	 */
// 	go func() {
// 		sess.updateInfo(trackPath, Info{
// 			GroupPriority:   priority,
// 			LatestGroupSequence: sequence,
// 			GroupExpires:        expires,
// 		})
// 	}()

// 	return streams, nil
// }

// func (sess *ClientSession) AcceptDataStream(ctx context.Context) (Group, moq.ReceiveStream, error) {
// 	return sess.acceptDataStream(ctx)
// }

// func (sess *ClientSession) SendDatagram(subscription Subscription, sequence GroupSequence, priority GroupPriority, data []byte) error {
// 	g := Group{
// 		subscribeID:       subscription.subscribeID,
// 		groupSequence:     sequence,
// 		GroupPriority: priority,
// 	}

// 	return sess.sendDatagram(g, data)
// }

// func (sess *ClientSession) ReceiveDatagram(ctx context.Context) (Group, []byte, error) {
// 	return sess.receiveDatagram(ctx)
// }

// func (sess *ClientSession) updateInfo(trackPath string, info Info) {
// 	sess.iMu.Lock()
// 	defer sess.iMu.Unlock()

// 	oldInfo, ok := sess.infos[trackPath]
// 	if !ok {
// 		sess.infos[trackPath] = info
// 	} else {
// 		if info.GroupPriority != 0 {
// 			info.GroupPriority = oldInfo.GroupPriority
// 		}
// 		if info.LatestGroupSequence != 0 {
// 			info.LatestGroupSequence = oldInfo.LatestGroupSequence
// 		}
// 		if info.GroupOrder != 0 {
// 			info.GroupOrder = oldInfo.GroupOrder
// 		}
// 		if info.GroupExpires != 0 {
// 			info.GroupExpires = oldInfo.GroupExpires
// 		}
// 	}
// }

// func (sess *ClientSession) getCurrentInfo(trackPath string) (Info, bool) {
// 	sess.iMu.RLock()
// 	defer sess.iMu.RUnlock()

// 	info, ok := sess.infos[trackPath]
// 	if !ok {
// 		return Info{}, false
// 	}

// 	return info, ok
// }

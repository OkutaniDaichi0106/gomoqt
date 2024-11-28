package moqt

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

var defaultRelayManager = NewRelayManager()

type Relayer struct {
	Path string

	//
	RequestHandler RequestHandler

	ServerSessionHandler ServerSessionHandler

	/*
	 * Relay Manager
	 * This field is optional
	 * If no value is set, default RelayManger is used
	 */
	RelayManager *RelayManager

	GoAwayFunc func() (string, time.Duration)

	BufferSize int
}

func (r Relayer) run(sess *ServerSession) {
	if r.RequestHandler == nil {
		panic("no relay manager")
	}
	if r.RelayManager == nil {
		r.RelayManager = defaultRelayManager
	}
	if r.BufferSize < 1 {
		r.BufferSize = 1
	}

	/*
	 * Handle Session
	 */
	go r.ServerSessionHandler.HandleServerSession(sess)

	/*
	 * Listen bidirectional streams
	 */
	go r.listenBiStreams(sess)

	/*
	 * Listen unidirectional streams
	 */
	go r.listenUniStreams(sess)

	go r.listenDatagram(sess)

	select {}
}

func (r Relayer) listenBiStreams(sess *ServerSession) {
	for {
		stream, err := sess.conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		go func(stream Stream) {
			qvr := quicvarint.NewReader(stream)

			num, err := qvr.ReadByte()
			if err != nil {
				slog.Error("failed to read a Stream Type ID", slog.String("error", err.Error()))
			}

			switch StreamType(num) {
			case ANNOUNCE:
				slog.Info("Announce Stream was opened")

				interest, err := getInterest(qvr)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				w := AnnounceWriter{
					doneCh: make(chan struct{}, 1),
					stream: stream,
				}

				// Announce
				announcements, ok := r.RelayManager.GetAnnouncements(interest.TrackPrefix)
				if !ok || announcements == nil {
					announcements = make([]Announcement, 0)
				}

				r.RequestHandler.HandleInterest(interest, announcements, w)
				<-w.doneCh
			case SUBSCRIBE:
				slog.Info("Subscribe Stream was opened")

				subscription, err := getSubscription(qvr)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream gracefully
					err = stream.Close()
					if err != nil {
						slog.Error("failed to close the stream", slog.String("error", err.Error()))
						return
					}

					return
				}

				//
				sw := SubscribeResponceWriter{
					stream: stream,
					doneCh: make(chan struct{}, 1),
				}

				// Get any Infomation of the track
				info, ok := r.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
				if ok {
					// Handle with out
					r.RequestHandler.HandleSubscribe(subscription, &info, sw)
				} else {
					r.RequestHandler.HandleSubscribe(subscription, nil, sw)
				}

				<-sw.doneCh

				/*
				 * Accept the new subscription
				 */
				sess.acceptSubscription(subscription)

				/*
				 * Catch any Subscribe Update or any error from the subscriber
				 */
				for {
					update, err := getSubscribeUpdate(subscription, qvr)
					if err != nil {
						slog.Info("catched an error from the subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Info("received a subscribe update request", slog.Any("subscription", update))

					sw := SubscribeResponceWriter{
						stream: stream,
						doneCh: make(chan struct{}, 1),
					}

					// Get any Infomation of the track
					info, ok := r.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
					if ok {
						r.RequestHandler.HandleSubscribe(update, &info, sw)
					} else {
						r.RequestHandler.HandleSubscribe(update, nil, sw)
					}

					<-sw.doneCh

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription
					 */
					sess.updateSubscription(update)
					subscription = update
				}

				sess.stopSubscription(subscription.subscribeID)
				slog.Info("unsubscribed", slog.Any("subscription", subscription))

				// Close the Stream gracefully
				err = stream.Close()
				if err != nil {
					slog.Error("failed to close the stream", slog.String("error", err.Error()))
					return
				}
			case FETCH:
				slog.Info("Fetch Stream was opened")

				fetchRequest, err := getFetchRequest(qvr)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					return
				}

				w := FetchResponceWriter{
					doneCh: make(chan struct{}, 1),
					stream: stream,
				}

				r.RequestHandler.HandleFetch(fetchRequest, w)

				<-w.doneCh
			case INFO:
				slog.Info("Info Stream is opened")

				infoRequest, err := getInfoRequest(qvr)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				w := InfoWriter{
					doneCh: make(chan struct{}, 1),
					stream: stream,
				}

				info, ok := r.RelayManager.GetInfo(infoRequest.TrackNamespace, infoRequest.TrackName)
				if ok {
					r.RequestHandler.HandleInfoRequest(infoRequest, &info, w)
				} else {
					r.RequestHandler.HandleInfoRequest(infoRequest, nil, w)
				}

				<-w.doneCh
			default:
				err := ErrInvalidStreamType
				slog.Error(err.Error(), slog.Uint64("ID", uint64(num)))

				// Cancel reading and writing
				stream.CancelRead(err.StreamErrorCode())
				stream.CancelWrite(err.StreamErrorCode())

				return
			}
		}(stream)
	}
}

func (r Relayer) listenUniStreams(sess *ServerSession) {
	for {
		group, stream, err := sess.acceptDataStream(context.TODO())
		if err != nil {
			slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
			return
		}

		go func(stream ReceiveStream) {
			/*
			 * Verify if subscribed or not by finding a subscription having the Subscribe ID in the Group
			 */
			sess.rsMu.RLock()
			defer sess.rsMu.RUnlock()

			sw, ok := sess.subscribeWriters[group.subscribeID]
			if !ok {
				slog.Error("received a group of unsubscribed track", slog.Any("group", group))
				return
			}

			// Make a copy to prevent the value from being overwritten along the way
			subscription := sw.subscription

			/*
			 * Distribute data
			 */
			// Find destinations subscribing the Group's Track
			dests, ok := r.RelayManager.findDestinations(strings.Split(subscription.TrackNamespace, "/"), subscription.TrackName, subscription.GroupOrder)
			if !ok {
				slog.Error("no destinations")
				return
			}

			/*
			 * Open Streams
			 */
			streams := make([]SendStream, len(dests))
			var mu sync.Mutex
			for _, destSess := range dests {
				// Skip to send data if the Session is nil
				if destSess == nil {
					continue
				}

				go func(sess *ServerSession) {
					mu.Lock()
					defer mu.Unlock()

					// Verify if the Group is needed
					for _, subscription := range destSess.receivedSubscriptions {
						if subscription.MinGroupSequence > group.groupSequence {
							continue
						}

						if subscription.MaxGroupSequence < group.groupSequence {
							continue
						}

						// Update the Group's Subscribe ID to the one in the Session
						group.subscribeID = subscription.subscribeID

						// Open a Data Stream
						stream, err := destSess.openDataStream(group)
						if err != nil {
							slog.Error("failed to open a data stream")
							// Notify the Group may be dropped
							// TODO:

							return
						}

						streams = append(streams, stream)
					}
				}(sess)

			}

			/*
			 * Read and send data
			 */
			buf := make([]byte, r.BufferSize*(1<<10)) // TODO: tune the size
			for {
				n, err := stream.Read(buf)
				if err != nil && err != io.EOF {
					slog.Error("failed to read data", slog.String("error", err.Error()))
					return
				}

				// Distribute the data
				for _, stream := range streams {
					go func(stream SendStream) {
						_, err := stream.Write(buf[:n])
						if err != nil {
							slog.Error("failed to send data", slog.String("error", err.Error()))
							return
						}
					}(stream)
				}

				if err == io.EOF {
					break
				}
			}

			//
			slog.Info("all data was distributed", slog.Any("group", group))
		}(stream)

	}
}

func (r Relayer) listenDatagram(sess *ServerSession) {
	for {
		group, payload, err := sess.receiveDatagram(context.TODO())
		if err != nil {
			slog.Error("failed to receive a datagram", slog.String("error", err.Error()))
			return
		}

		/*
		 *
		 * Verify if subscribed or not by finding a subscription having the Group's Subscribe ID
		 */
		sess.rsMu.RLock()
		defer sess.rsMu.RUnlock()

		sw, ok := sess.subscribeWriters[group.subscribeID]
		if !ok {
			slog.Error("received a group of unsubscribed track", slog.Any("group", group))
			return
		}

		subscription := sw.subscription

		/*
		 * Distribute data
		 */
		// Find destinations subscribing the Group's Track
		dests, ok := r.RelayManager.findDestinations(strings.Split(subscription.TrackNamespace, "/"), subscription.TrackName, subscription.GroupOrder)
		if !ok {
			slog.Error("")
			return
		}

		/*
		 * Send data
		 */
		var mu sync.Mutex
		for _, destSess := range dests {

			if destSess == nil {
				// Skip
				continue
			}

			go func(sess *session) {
				mu.Lock()
				defer mu.Unlock()

				for _, subscription := range sess.receivedSubscriptions {
					if subscription.MinGroupSequence > group.groupSequence {
						continue
					}

					if subscription.MaxGroupSequence < group.groupSequence {
						continue
					}

					// Update the Group's Subscribe ID to the one in the Session
					group.subscribeID = subscription.subscribeID

					err := sess.sendDatagram(group, payload)
					if err != nil {
						slog.Error("failed to open a Data Stream")
						// Notify the Group may be dropped
						// TODO:

						return
					}
				}
			}(destSess)
		}

		slog.Info("all data was distributed", slog.Any("group", group))
	}

}

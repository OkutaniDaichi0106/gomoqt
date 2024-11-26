package moqt

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

var defaultRelayManager = NewRelayManager()

type Relayer struct {
	Path string

	//
	RequestHandler RequestHandler

	SessionHandler SessionHandler

	/*
	 * Relay Manager
	 * This field is optional
	 * If no value is set, default RelayManger is used
	 */
	RelayManager *RelayManager

	GoAwayFunc func() (string, time.Duration)

	BufferSize int
}

func (r Relayer) run(sess *session) {
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
	go r.SessionHandler.HandleSession(&ServerSession{session: sess})

	/*
	 * Listen bidirectional streams
	 */
	go r.listenBiStreams(sess)

	/*
	 * Listen unidirectional streams
	 */
	go r.listenUniStreams(sess)
}

func (r Relayer) listenBiStreams(sess *session) {
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
				slog.Debug("Announce Stream is opened")

				interest, err := getInterest(qvr)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				w := AnnounceWriter{
					doneCh: make(chan struct{}),
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
				slog.Debug("Subscribe Stream was opened")

				subscription, err := getSubscription(qvr)
				if err != nil {
					slog.Debug("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream gracefully
					err = stream.Close()
					if err != nil {
						slog.Debug("failed to close the stream", slog.String("error", err.Error()))
						return
					}

					return
				}

				//
				sw := SubscribeResponceWriter{
					stream: stream,
					doneCh: make(chan struct{}),
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
						slog.Debug("catched an error from the subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Debug("received a subscribe update request", slog.Any("subscription", update))

					sw := SubscribeResponceWriter{
						stream: stream,
						doneCh: make(chan struct{}),
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
				slog.Debug("unsubscribed", slog.Any("subscription", subscription))

				// Close the Stream gracefully
				err = stream.Close()
				if err != nil {
					slog.Debug("failed to close the stream", slog.String("error", err.Error()))
					return
				}
			case FETCH:
				slog.Debug("Fetch Stream was opened")

				fetchRequest, err := getFetchRequest(qvr)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					return
				}

				w := FetchResponceWriter{
					doneCh: make(chan struct{}),
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
					doneCh: make(chan struct{}),
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

func (r Relayer) listenUniStreams(sess *session) {
	for {
		stream, err := sess.conn.AcceptUniStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		go func(stream ReceiveStream) {
			/*
			 * Get a group
			 */
			group, err := getGroup(quicvarint.NewReader(stream))
			if err != nil {
				slog.Error("failed to get a group", slog.String("error", err.Error()))
				return
			}

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
			for _, destSess := range dests {
				// Skip to send data if the Session is nil
				if destSess == nil {
					continue
				}

				// Verify the Group is
				for _, subscription := range destSess.receivedSubscriptions {
					if subscription.MinGroupSequence > group.groupSequence {
						continue
					}

					if subscription.MaxGroupSequence < group.groupSequence {
						continue
					}

					// make new Group by changing Subscribe ID
					group.subscribeID = subscription.subscribeID

					stream, err := destSess.openDataStream(group)
					if err != nil {
						slog.Error("failed to open a Data Stream")
						// Notify the Group may be dropped // TODO:

						return
					}

					streams = append(streams, stream)
				}

			}

			/*
			 * Read and send data
			 */
			buf := make([]byte, r.BufferSize*(1<<10)) // TODO: tune the size
			for {
				n, err := stream.Read(buf)
				if err != nil && err != io.EOF {
					slog.Debug("failed to read data", slog.String("error", err.Error()))
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
			slog.Debug("all data was distributed", slog.Any("group", group))
		}(stream)

	}
}

func (r Relayer) listenDatagram(sess *session) {
	for {
		data, err := sess.conn.ReceiveDatagram(context.TODO()) //TODO
		if err != nil {
			slog.Debug("failed to receive a datagram")
		}
		qvr := quicvarint.NewReader(bytes.NewReader(data))

		/*
		 * Read the object's group
		 */
		group, err := getGroup(qvr)
		if err != nil {
			slog.Error("failed to get a group", slog.String("error", err.Error()))
			return
		}

		/*
		 * Find a subscription corresponding to the Subscribe ID in the Group
		 * Verify if subscribed or not
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
		// Find a Track corresponding to the Group's Track
		dests, ok := r.RelayManager.findDestinations(strings.Split(subscription.TrackNamespace, "/"), subscription.TrackName, subscription.GroupOrder)
		if !ok {
			slog.Error("")
			return
		}

		/*
		 * Send data
		 */

		for _, destSess := range dests {

			if destSess == nil {
				// Skip
				continue
			}

			go func(sess *session) {

				err := sess.sendDatagram(group, data)
				if err != nil {
					slog.Error("failed to open a Data Stream")
					// Notify the Group may be dropped // TODO:

					return
				}
			}(destSess)
		}

		slog.Debug("all data was distributed", slog.Any("group", group))
	}

}

package moqt

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

var defaultRelayManager = NewRelayManager()

type Relayer struct {
	Path string

	SessionHandler ServerSessionHandler

	/*
	 * Relay Manager
	 * This field is optional
	 * If no value is set, default Relay Manger will be used
	 */
	RelayManager *RelayManager

	BufferSize int

	CacheManager CacheManager
}

func (r Relayer) run(sess *ServerSession) {
	if r.RelayManager == nil {
		r.RelayManager = defaultRelayManager
	}
	if r.BufferSize < 1 {
		r.BufferSize = 1
	}

	/*
	 * Handle Session
	 */
	go r.SessionHandler.HandleServerSession(sess)

	/*
	 * Listen bidirectional streams
	 */
	go r.listenBiStreams(sess)

	/*
	 * Listen unidirectional streams
	 */
	go r.listenUniStreams(sess)

	/*
	 * Listen datagrams
	 */
	go r.listenDatagrams(sess)

	select {}
}

func (r Relayer) listenBiStreams(sess *ServerSession) {
	for {
		stream, err := sess.conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		go func(stream moq.Stream) {
			var stm message.StreamTypeMessage
			err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to read a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			switch stm.StreamType {
			case stream_type_announce:
				slog.Info("Announce Stream was opened")

				interest, err := readInterest(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				slog.Info("received an interest", slog.Any("interest", interest))

				// Initialize a Announce Writer
				w := AnnounceWriter{
					stream: stream,
				}

				/*
				 * Announce
				 */
				// Find any Track Namespace node
				slog.Info("finding any announcements")
				tp := strings.Split(interest.TrackPrefix, "/")
				tnsNode, ok := r.RelayManager.findTrackNamespace(tp)
				if ok {
					// Get any Announcements under the Track Namespace
					announcements := tnsNode.getAnnouncements()

					// Send the Announcements
					for _, ann := range announcements {
						w.Announce(ann)
					}
				}

				// Register the Announce Writer
				r.RelayManager.RegisterFollower(interest.TrackPrefix, w)

				w.Close()
			case stream_type_subscribe:
				slog.Info("Subscribe Stream was opened")

				// Initialize a Subscriber Responce Writer
				sw := SubscribeResponceWriter{
					stream: stream,
				}

				// Get a subscription
				subscription, err := readSubscription(stream)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream
					sw.Reject(err)
					return
				}

				// Get any Infomation of the track
				info, ok := r.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
				if ok {
					// Accept the request if information exists
					sw.Accept(info)
				} else {
					// Reject the request if information does not exist
					sw.Reject(ErrTrackDoesNotExist)
					return
				}

				/*
				 * Accept the new subscription
				 */
				sess.acceptSubscription(subscription)

				/*
				 * Register the session as destinations of the Track
				 */

				/*
				 * Catch any Subscribe Update or any error from the subscriber
				 */
				for {
					update, err := readSubscribeUpdate(subscription, stream)
					if err != nil {
						slog.Info("catched an error from the subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Info("received a subscribe update request", slog.Any("subscription", update))

					sw := SubscribeResponceWriter{
						stream: stream,
					}

					// Get any Infomation of the track
					info, ok := r.RelayManager.GetInfo(subscription.TrackNamespace, subscription.TrackName)
					if ok {
						// Accept the request if information exists
						sw.Accept(info)
					} else {
						// Reject the request if information does not exist
						sw.Reject(ErrTrackDoesNotExist)
						return
					}

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription
					 */
					sess.updateSubscription(update)
					subscription = update
				}

				sess.removeSubscription(subscription)

				slog.Info("subscription has ended", slog.Any("subscription", subscription))

				// Close the Stream gracefully
				sw.Close()
				return
			case stream_type_fetch:
				// Handle the Fecth Stream

				slog.Info("Fetch Stream was opened")

				// Initialize a fetch responce writer
				frw := FetchResponceWriter{
					stream: stream,
				}

				// Get a fetch request
				req, err := readFetchRequest(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				slog.Info("get a fetch request", slog.Any("fetch request", req))

				// Get a data rader
				r, err := r.CacheManager.GetFrame(req.TrackNamespace, req.TrackName, req.GroupSequence, req.FrameSequence)
				if err != nil {
					slog.Error("failed to get a frame", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				// Verify if subscriptions corresponding to the ftch request exists
				for _, subscription := range sess.receivedSubscriptions {
					if subscription.TrackNamespace != req.TrackNamespace {
						continue
					}
					if subscription.TrackName != req.TrackName {
						continue
					}

					// Send the group data
					w, err := frw.SendGroup(Group{
						subscribeID:       subscription.subscribeID,
						groupSequence:     req.GroupSequence,
						PublisherPriority: PublisherPriority(req.SubscriberPriority), // TODO: Handle Publisher Priority
					})
					if err != nil {
						slog.Error("failed to send a group", slog.String("error", err.Error()))
						frw.Reject(err)
						return
					}

					// Send the data by copying it from the reader
					io.Copy(w, r)
				}

				// Close the Fetch Stream gracefully
				frw.Close()
				return
			case stream_type_info:
				// Handle the Info Stream
				slog.Info("Info Stream was opened")

				// Get a info request
				infoRequest, err := readInfoRequest(stream)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				iw := InfoWriter{
					stream: stream,
				}

				info, ok := r.RelayManager.GetInfo(infoRequest.TrackNamespace, infoRequest.TrackName)
				if ok {
					iw.Answer(info)
				} else {
					iw.Reject(ErrTrackDoesNotExist)
					return
				}

				// Close the Info Stream gracefully
				iw.Close()
				return
			default:
				// Terminate the session if invalid Stream Type was detected
				sess.Terminate(ErrInvalidStreamType)

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

		go func(stream moq.ReceiveStream) {
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
			streams := make([]moq.SendStream, len(dests))
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
					go func(stream moq.SendStream) {
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

func (r Relayer) listenDatagrams(sess *ServerSession) {
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

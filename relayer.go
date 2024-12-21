package moqt

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type relayer interface {
}

var _ relayer = (*Relayer)(nil)

// func newRelayer(path string, upstream ServerSession) *Relayer {
// 	return &Relayer{
// 		TrackPath:   path,
// 		upstream:    upstream,
// 		downstreams: make([]ServerSession, 0),
// 		// BufferSize: 1,
// 	}
// }

type Relayer struct {
	TrackPath string

	upstream *Subscriber

	downstreams []*Publisher
	dsMu        sync.RWMutex

	BufferSize int

	//CacheManager CacheManager
}

func (r *Relayer) run() {
	// Listen bidirectional streams
	go r.listenBiStreams()

	go r.listenUniStreams()

}

func (r *Relayer) AddDownstream(sess ServerSession) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	r.downstreams = append(r.downstreams, sess)
}

// func (r *Relayer) listen(sess *ServerSession) {
// 	if r.BufferSize < 1 {
// 		r.BufferSize = 1
// 	}

// 	/*
// 	 * Handle Session
// 	 */

// 	/*
// 	 * Listen bidirectional streams
// 	 */
// 	go r.listenBiStreams(sess)

// 	/*
// 	 * Listen unidirectional streams
// 	 */
// 	go r.listenUniStreams(sess)

// 	/*
// 	 * Listen datagrams
// 	 */
// 	go r.listenDatagrams(sess)

// 	select {}
// }

func (r *Relayer) listenBiStreams(sess *ServerSession) {
	for {
		stream, err := sess.conn.AcceptStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			return
		}

		slog.Debug("some control stream was opened")

		go func(stream transport.Stream) {
			var stm message.StreamTypeMessage
			err := stm.Decode(stream)
			if err != nil {
				slog.Error("failed to read a Stream Type ID", slog.String("error", err.Error()))
				return
			}

			switch stm.StreamType {
			case stream_type_announce:
				slog.Debug("announce stream was opened")

				interest, err := readInterest(stream)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				slog.Info("Received an interest", slog.Any("interest", interest))

				// Initialize a Announce Writer
				aw := AnnounceSender{
					stream: stream,
				}

				annCh := make(chan Announcement, 1) // TODO: Tune the size

				// Register the Announce Writer
				r.RelayManager.registerFollower(interest.TrackPrefix, annCh)

				/*
				 * Announce
				 */
				// Find anu current announcements
				slog.Info("Finding any announcements")
				anns := r.RelayManager.GetAnnouncements(interest.TrackPrefix)
				// Send the Announcements
				for _, ann := range anns {
					aw.Announce(ann)
				}

				for ann := range annCh {
					aw.Announce(ann)
				}

				aw.Close()
				return
			case stream_type_subscribe:
				slog.Debug("subscribe stream was opened")

				// Initialize a Subscriber Responce Writer
				sr := ReceivedSubscription{
					stream: stream,
				}

				// Get a subscription
				subscription, err := readSubscription(stream)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream
					sr.CancelRead(err)
					return
				}

				// Get any Infomation of the track
				info, ok := r.RelayManager.GetInfo(subscription.TrackPath)
				if ok {
					// Accept the request if information exists
					sr.Inform(info)
				} else {
					// Reject the request if information does not exist
					sr.CancelRead(ErrTrackDoesNotExist)
					return
				}

				// Set the subscription to the receiver
				sr.subscription = subscription

				/*
				 * Accept the new subscription
				 */
				sess.acceptNewSubscription(&sr)

				/*
				 * Register the session as destinations of the Track
				 */
				//TODO

				/*
				 * Catch any Subscribe Update or any error from the subscriber
				 */
				for {
					update, err := sr.ReceiveUpdate()
					if err != nil {
						slog.Info("failed to read a subscribe update", slog.String("error", err.Error()))
						break
					}

					slog.Info("received a subscribe update", slog.Any("update", update))

					// Get any Infomation of the track
					info, ok := r.RelayManager.GetInfo(subscription.TrackPath)
					if ok {
						// Accept the request if information exists
						sr.Inform(info)
					} else {
						// Reject the request if information does not exist
						sr.CancelRead(ErrTrackDoesNotExist)
						return
					}

					slog.Info("updated a subscription", slog.Any("from", subscription), slog.Any("to", update))

					/*
					 * Update the subscription by overwriting the receiver
					 */
					sr.updateSubscription(update)
				}

				sess.deleteSubscription(subscription)

				slog.Info("subscription has ended", slog.Any("subscription", subscription))

				// Close the Stream gracefully
				sr.Close()
				return
			case stream_type_fetch:
				slog.Debug("fetch stream was opened")

				// Initialize a fetch responce writer
				frw := ReceivedFetch{
					stream: stream,
				}

				// Get a fetch request
				req, err := readFetch(stream)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				slog.Info("Got a fetch request", slog.Any("fetch request", req))

				// Get a data rader
				r, err := r.CacheManager.GetFrame(req.TrackPath, req.GroupSequence, req.FrameSequence)
				if err != nil {
					slog.Error("failed to get a frame", slog.String("error", err.Error()))
					frw.Reject(err)
					return
				}

				// Verify if subscriptions corresponding to the ftch request exists
				for _, sr := range sess.subscribeReceivers {
					if sr.subscription.TrackPath != req.TrackPath {
						// Initialize a group
						group := Group{
							subscribeID:   sr.subscription.subscribeID,
							groupSequence: req.GroupSequence,
							GroupPriority: TrackPriority(req.TrackPriority), // TODO: Handle Publisher Priority
						}

						// Send the group data
						w, err := frw.SendGroup(group)
						if err != nil {
							slog.Error("failed to send a group", slog.String("error", err.Error()))
							frw.Reject(err)
							return
						}

						// Send the data by copying it from the reader
						io.Copy(w, r)
					}
				}

				// Close the Fetch Stream gracefully
				frw.Close()
				return
			case stream_type_info:
				slog.Debug("info stream was opened")

				// Get a info request
				infoRequest, err := readInfoRequest(stream)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				iw := ReceivedInfoRequest{
					stream: stream,
				}

				info, ok := r.RelayManager.GetInfo(infoRequest.TrackPath)
				if ok {
					iw.Inform(info)
				} else {
					iw.CancelInform(ErrTrackDoesNotExist)
					return
				}

				// Close the Info Stream gracefully
				iw.Close()
				return
			default:
				slog.Debug("unknown stream was opend")
				// Terminate the session
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

		go func(stream transport.ReceiveStream) {
			/*
			 * Verify if subscribed or not by finding a subscription having the Subscribe ID in the Group
			 */
			sess.srMu.RLock()
			defer sess.srMu.RUnlock()

			sw, ok := sess.subscribeSenders[group.subscribeID]
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
			tp := strings.Split(subscription.TrackPath, "/")
			dests, ok := r.RelayManager.findDestinations(tp[:len(tp)-1], tp[len(tp)-1], subscription.GroupOrder)
			if !ok {
				slog.Error("no destinations")
				return
			}

			/*
			 * Open Streams
			 */
			streams := make([]transport.SendStream, len(dests))
			var mu sync.Mutex
			for _, destSess := range dests {
				// Skip to send data if the Session is nil
				if destSess == nil {
					continue
				}

				go func(sess *ServerSession) {
					mu.Lock()
					defer mu.Unlock()

					// Verify if the group is required for the subscription
					for _, sr := range destSess.subscribeReceivers {
						subscription := sr.subscription

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
					go func(stream transport.SendStream) {
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
		sess.srMu.RLock()
		defer sess.srMu.RUnlock()

		sw, ok := sess.subscribeSenders[group.subscribeID]
		if !ok {
			slog.Error("received a group of unsubscribed track", slog.Any("group", group))
			return
		}

		subscription := sw.subscription

		/*
		 * Distribute data
		 */
		// Find destinations subscribing the Group's Track
		tp := strings.Split(subscription.TrackPath, "/")
		dests, ok := r.RelayManager.findDestinations(tp[:len(tp)-1], tp[len(tp)-1], subscription.GroupOrder)
		if !ok {
			slog.Error("destinations not found")
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

				for _, sr := range sess.subscribeReceivers {
					if sr.subscription.MinGroupSequence > group.groupSequence {
						continue
					}

					if sr.subscription.MaxGroupSequence < group.groupSequence {
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

package moqt

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/quic-go/quic-go/quicvarint"
)

var defaultRelayManager = NewRelayManager()

type Relayer struct {
	Path string

	//
	RequestHandler

	/*
	 * Relay Manager
	 * This field is optional
	 * If no value is set, default RelayManger is used
	 */
	RelayManager *RelayManager
}

func (r Relayer) listen(sess *Session) {
	if r.RequestHandler == nil {
		panic("no relay manager")
	}
	if r.RelayManager == nil {
		r.RelayManager = defaultRelayManager
	}

	go r.listenBiStreams(sess)
	go r.listenUniStreams(sess)

	terr := <-sess.terrCh
	slog.Error("Session was terminated", slog.String("error", terr.Error()))

	sess.Terminate(terr)
}

func (r Relayer) listenBiStreams(sess *Session) {
	for {
		stream, err := sess.Connection.AcceptStream(context.Background())
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

				r.HandleInterest(interest, announcements, w)
				<-w.doneCh
			case SUBSCRIBE:
				slog.Debug("Subscribe Stream was opened")

				subscription, err := getSubscription(qvr)
				if err != nil {
					slog.Debug("failed to get a subscription", slog.String("error", err.Error()))

					// Close the Stream gracefully
					slog.Debug("closing a Subscribe Stream", slog.String("error", err.Error()))
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
					r.HandleSubscribe(subscription, &info, sw)
				} else {
					r.HandleSubscribe(subscription, nil, sw)
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
						r.HandleSubscribe(update, &info, sw)
					} else {
						r.HandleSubscribe(update, nil, sw)
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

				r.HandleFetch(fetchRequest, w)

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
					r.HandleInfoRequest(infoRequest, &info, w)
				} else {
					r.HandleInfoRequest(infoRequest, nil, w)
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

func (r Relayer) listenUniStreams(sess *Session) {
	for {
		stream, err := sess.Connection.AcceptUniStream(context.Background())
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
			 * Find a subscription corresponding to the Subscribe ID in the Group
			 * Verify if subscribed or not
			 */
			sess.rsMu.RLock()
			defer sess.rsMu.RUnlock()

			sw, ok := sess.subscribeWriters[group.SubscribeID]
			if !ok {
				slog.Error("received a group of unsubscribed track", slog.Any("group", group))
				return
			}
			subscription := sw.subscription

			/*
			 * Distribute data
			 */
			// Find a Track corresponding to the Group's Track
			tnNode, ok := r.RelayManager.findTrack(strings.Split(subscription.TrackNamespace, "/"), subscription.TrackName)
			if !ok {
				slog.Error("")
				return
			}

			dataCh := make(chan []byte, 1<<5) // TODO: Tune the size

			/*
			 * Read data
			 */
			go func() {
				buf := make([]byte, 1<<10)
				for {
					n, err := stream.Read(buf)
					if err != nil {
						if err == io.EOF {
							dataCh <- buf[:n]
						}
						return
					}

					dataCh <- buf[:n]
				}
			}()

			/*
			 * Send data
			 */
			for data := range dataCh {
				go func(data []byte) {
					for _, destSess := range tnNode.destinations {

						if destSess == nil {
							// Skip
							continue
						}

						go func(sess *Session) {
							stream, err := sess.OpenDataStream(group)
							if err != nil {
								slog.Error("failed to open a Data Stream")
								// Notify the Group may be dropped // TODO:

								return
							}

							_, err = stream.Write(data)
							if err != nil {
								// Notify the Group may be dropped // TODO:
								return
							}
						}(destSess)
					}
				}(data)
			}

			slog.Debug("all data was distributed", slog.Any("group", group))
		}(stream)
	}
}

package moqt

import (
	"context"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type Relayer struct {
	Path string

	//
	Publisher Publisher

	Subscriber Subscriber

	RelayManager *RelayManager
}

func (r Relayer) listen(sess *Session) {
	if r.RelayManager == nil {
		r.RelayManager = defaultRelayManager
	}

	go r.listenBiStreams(sess)
	go r.listenUniStreams(sess)

	err := <-sess.terrCh
	slog.Error("Session was terminated", slog.String("error", err.Error()))

	sess.Terminate(err)
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
				slog.Info("Announce Stream is opened")

				interest, err := getInterest(qvr)
				if err != nil {
					slog.Error("failed to get an Interest", slog.String("error", err.Error()))
					return
				}

				w := defaultAnnounceWriter{
					errCh:  make(chan error),
					stream: stream,
				}

				r.Publisher.Handler.HandleInterest(interest, w)

				err = <-w.errCh
				if err != nil {
					slog.Error("rejected an interest", slog.Any("interest", interest))
					return
				}

				// Catch any error
				for {
					_, err := stream.Read([]byte{})
					if err != nil {
						slog.Error("receives an error in announce stream from client", slog.String("error", err.Error()))
						break
					}
				}

				// Close the Stream
				err = stream.Close()
				if err != nil {
					slog.Error("failed to close the stream", slog.String("error", err.Error()))
					return
				}

			case SUBSCRIBE:
				slog.Info("Subscribe Stream is opened")

				subscription, err := getSubscription(qvr)
				if err != nil {
					slog.Error("failed to get a subscription", slog.String("error", err.Error()))
					return
				}

				sw := defaultSubscribeResponceWriter{
					stream: stream,
					errCh:  make(chan error),
				}

				r.Publisher.Handler.HandleSubscribe(subscription, sw)

				err = <-sw.errCh
				if err != nil {
					slog.Error("failed to subscribe", slog.String("error", err.Error()))
					return
				}

				// Register the subscription
				sess.addSubscription(subscription)

				/*
				 * Notice the track information
				 */
				iw := defaultInfoWriter{
					errCh:  make(chan error),
					stream: stream,
				}
				// Handle
				r.Publisher.Handler.HandleInfoRequest(InfoRequest{
					TrackNamespace: subscription.TrackNamespace,
					TrackName:      subscription.TrackName,
				}, iw)

				err = <-iw.errCh
				if err != nil {
					slog.Error("catch an error", slog.String("error", err.Error()))
				}

				// Catch any Subscribe Update or any error from the subscriber
				for {
					subscription, err := getSubscribeUpdate(subscription, qvr)
					if err != nil {
						slog.Error("receives an error in subscriber stream from a subscriber", slog.String("error", err.Error()))
						break
					}

					slog.Info("received a subscription", slog.Any("subscription", subscription))

					sw := defaultSubscribeResponceWriter{
						stream: stream,
						errCh:  make(chan error),
					}

					r.Publisher.Handler.HandleSubscribe(subscription, sw)

					err = <-sw.errCh
					if err != nil {
						slog.Error("reject a subscription", slog.Any("subscription", subscription), slog.String("error", err.Error()))
						return
					}

					slog.Info("accept a subscribe update", slog.Any("new subscription", subscription))

					// Register the subscription
					sess.addSubscription(subscription)
				}

				// Close the Stream gracefully
				err = stream.Close()
				if err != nil {
					slog.Error("failed to close the stream", slog.String("error", err.Error()))
					return
				}

			case FETCH:
				slog.Info("Fetch Stream is opened")

				fetchRequest, err := getFetchRequest(qvr)
				if err != nil {
					slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
					return
				}

				w := defaultFetchRequestWriter{
					errCh:  make(chan error),
					stream: stream,
				}

				r.Publisher.Handler.HandleFetch(fetchRequest, w)

				err = <-w.errCh
				if err != nil {
					slog.Error("catch an error", slog.String("error", err.Error()))
					return
				}

				// Close the Stream
				err = stream.Close()
				if err != nil {
					slog.Error("failed to close the stream", slog.String("error", err.Error()))
					return
				}

			case INFO:
				slog.Info("Info Stream is opened")

				infoRequest, err := getInfoRequest(qvr)
				if err != nil {
					slog.Error("failed to get a info-request", slog.String("error", err.Error()))
					return
				}

				w := defaultInfoWriter{
					errCh:  make(chan error),
					stream: stream,
				}

				r.Publisher.Handler.HandleInfoRequest(infoRequest, w)

				err = <-w.errCh

				if err != nil {
					slog.Error("catch an error", slog.String("error", err.Error()))
				}

				// Close the Stream
				err = stream.Close()
				if err != nil {
					slog.Error("failed to close the stream", slog.String("error", err.Error()))
					return
				}

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
			group, err := getGroup(quicvarint.NewReader(stream))
			if err != nil {
				slog.Error("failed to get a group", slog.String("error", err.Error()))
				return
			}

			r.Subscriber.Handler.HandleData(group, stream)
		}(stream)
	}
}

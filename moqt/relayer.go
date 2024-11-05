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

	manager relayManager
}

func (r Relayer) init() {
	r.manager = newRelayManager()
}

func (r Relayer) listen(sess Session) {
	r.init()

	biCh := make(chan Stream, 1<<3)         // TODO: Tune the size
	uniCh := make(chan ReceiveStream, 1<<3) // TODO: Tune the size

	go func() {
		for {
			stream, err := sess.Connection.AcceptStream(context.Background())
			if err != nil {
				slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
				return
			}

			biCh <- stream
		}
	}()

	go func() {
		for {
			stream, err := sess.Connection.AcceptUniStream(context.Background())
			if err != nil {
				slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
				return
			}

			uniCh <- stream
		}
	}()

	/*
	 *
	 */
	for {
		select {
		case stream := <-biCh:
			// Handle the Stream
			go func(stream Stream) {
				qvr := quicvarint.NewReader(stream)

				num, err := qvr.ReadByte()
				if err != nil {
					slog.Error("failed to read a Stream Type ID", slog.String("error", err.Error()))
				}

				switch StreamType(num) {
				case ANNOUNCE:
					interest, err := getInterest(qvr)
					if err != nil {
						slog.Error("failed to get an Interest", slog.String("error", err.Error()))
						return
					}

					r.Publisher.Handler.HandleInterest(interest, defaultAnnounceWriter{
						stream: stream,
					})

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

					// Notice the track information
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
							slog.Error("receives an error in announce stream from a subscriber", slog.String("error", err.Error()))
							break
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
					}

					// Close the Stream gracefully
					err = stream.Close()
					if err != nil {
						slog.Error("failed to close the stream", slog.String("error", err.Error()))
						return
					}
				case FETCH:
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
					slog.Error("invalid Stream Type ID", slog.Uint64("ID", uint64(num)))
					return
				}
			}(stream)
		case stream := <-uniCh:
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
}

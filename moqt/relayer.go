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
}

func (r Relayer) listen(sess Session) {
	r.Publisher.init()
	r.Subscriber.init()

	go func() {
		for {
			stream, err := sess.Connection.AcceptStream(context.Background())
			if err != nil {
				slog.Error("failed to accept a bidirectional stream", slog.String("error", err.Error()))
			}

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

					r.Publisher.interestCh <- struct {
						Interest
						AnnounceWriter
					}{
						Interest: interest,
						AnnounceWriter: defaultAnnounceWriter{
							stream: stream,
						},
					}

					// Listen any error
					for {
						_, err := stream.Read([]byte{})
						if err != nil {
							slog.Error("receives an error in announce stream from client", slog.String("error", err.Error()))
							return
						}
					}
				case SUBSCRIBE:
					subscription, err := getSubscription(qvr)
					if err != nil {
						slog.Error("failed to get a subscription", slog.String("error", err.Error()))
						return
					}

					r.Publisher.subscriptionCh <- struct {
						Subscription
						SubscribeResponceWriter
					}{
						Subscription: subscription,
						SubscribeResponceWriter: defaultSubscribeResponceWriter{
							stream: stream,
						},
					}

					// Listen any error
					for {
						_, err := stream.Read([]byte{})
						if err != nil {
							slog.Error("receives an error in announce stream from client", slog.String("error", err.Error()))
							return
						}
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

					r.Publisher.fetchReqCh <- struct {
						FetchRequest
						FetchResponceWriter
					}{
						FetchRequest:        fetchRequest,
						FetchResponceWriter: w,
					}

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

					r.Publisher.infoReqCh <- struct {
						InfoRequest
						InfoWriter
					}{
						InfoRequest: infoRequest,
						InfoWriter:  w,
					}

					err = <-w.errCh

					if err != nil {
						slog.Error("catch an error", slog.String("error", err.Error()))
					}

					// Close the stream
					stream.Close()
				default:
					slog.Error("invalid Stream Type ID", slog.Uint64("ID", uint64(num)))
					return
				}
			}(stream)
		}
	}()

	go r.Publisher.listen()
	go r.Subscriber.listen()

}

// type RelayHandler interface {
// 	PublisherHandler
// 	SubscriberHandler
// }

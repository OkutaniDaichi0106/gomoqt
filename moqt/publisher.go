package moqt

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type Publisher struct {
	// Handlers
	PublisherHandler

	LocalTrack []string

	// //
	// interestR InterestReader
	// announceW AnnounceWriter

	// //
	// subscribeR  SubscribeReader
	// subscribeRW SubscribeResponceWriter

	// fetchRR FetchRequestReader
	// fetchRW FetchResponceWriter
}

type PublisherHandler interface {
	InterestHandler
	SubscribeHandler
	FetchHandler
	InfoRequestHandler
}

func (p Publisher) run(sess Session) {
	/*
	 * Handle Announce, Subscribe, Fetch, Info Stream
	 */
	for {
		// Accept a Stream
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			slog.Error(err.Error())
			return
		}

		// Handle the Stream
		go func(stream Stream) {
			// Read the first byte and get Stream Type
			buf := make([]byte, 1)
			_, err = stream.Read(buf)
			if err != nil {
				slog.Error("failed to read a Stream Type", slog.String("error", err.Error()))
				return
			}
			// Verify if the Stream Type is valid
			switch protocol.StreamType(buf[0]) {
			case protocol.ANNOUNCE:
				req, err := defaultInterestReader{}.Read(quicvarint.NewReader(stream))
				if err != nil {
					slog.Error("failed to get a set-up request", slog.String("error", err.Error()))
					return
				}

				// Handle the set-up request
				p.HandleInterest(req, defaultAnnounceWriter{stream: stream})

			case protocol.SUBSCRIBE:
				subscription, err := defaultSubscribeReader{streaem: stream}.Read()
				if err != nil {
					slog.Error("failed to get a subscribe request", slog.String("error", err.Error()))
					return
				}

				p.HandleSubscribe(subscription, defaultSubscribeResponceWriter{
					stream: stream,
				})

			// case protocol.FETCH:
			// 	req, err := defaulFetchRequestHandler.Read(quicvarint.NewReader(stream))
			// 	if err != nil {
			// 		slog.Error("failed to get a fetch request", slog.String("error", err.Error()))
			// 		return
			// 	}

			// 	p.HandleFetch(req, defaulFetchRequestHandler{})
			case protocol.INFO:
				req, err := defaultInfoRequestReader{}.Read()
			default:
				slog.Error("unexpected Stream Type ID", slog.Uint64("ID", uint64(buf[0]))) // TODO
			}
		}(stream)
	}
}

package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func main() {
	server := moqt.Server{
		Addr: "localhost:4444",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			InsecureSkipVerify: true, // TODO: Not recommended for production
		},
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
	}
	// Serve moq over webtransport
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		mux := moqt.NewTrackMux()

		sess, err := server.AcceptWebTransport(w, r, mux)
		if err != nil {
			slog.Error("failed to serve moq over webtransport", "error", err)
		}

		anns, err := sess.OpenAnnounceStream("")
		if err != nil {
			slog.Error("failed to open announce stream", "error", err)
		}

		for {
			ann, err := anns.ReceiveAnnouncement(context.Background())
			if err != nil {
				slog.Error("failed to receive announcements", "error", err)
				break
			}
			path := ann.BroadcastPath()
			if path.Extension() != ".echo" {
				continue
			}

			sub, err := sess.OpenTrackStream(path, "index", nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				break
			}

			moqt.HandleFunc(context.Background(), sub.BroadcastPath, func(pub *moqt.Publisher) {
				defer sub.TrackReader.Close()

				for {
					gr, err := sub.TrackReader.AcceptGroup(context.Background())
					if err != nil {
						slog.Error("failed to accept group", "error", err)
						break
					}

					gw, err := pub.TrackWriter.OpenGroup(gr.GroupSequence())
					if err != nil {
						slog.Error("failed to open group", "error", err)
						break
					}

					for {
						f, err := gr.ReadFrame()
						if err != nil {
							slog.Error("failed to accept frame", "error", err)
							break
						}
						err = gw.WriteFrame(f)
						if err != nil {
							slog.Error("failed to write frame", "error", err)
							break
						}
					}
				}
			})
		}

	})

	err := server.ListenAndServeTLS("examples/cert/localhost.pem", "examples/cert/localhost-key.pem")
	if err != nil {
		slog.Error("failed to listen and serve", "error", err)
	}
}

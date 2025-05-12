package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
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

		sess, err := server.AcceptWebTransport(w, r)
		if err != nil {
			slog.Error("failed to serve web transport", "error", err)
		}

		info, tr, err := sess.OpenTrackStream("/data", nil)
		if err != nil {
			slog.Error("failed to open track stream", "error", err)
		}
		defer tr.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tw := moqt.BuildTrack(ctx, ann.TrackPath(), info, 0)

		for {
			gr, err := tr.AcceptGroup(context.Background())
			if err != nil {
				slog.Error("failed to accept group", "error", err)
				break
			}

			gw, err := tw.OpenGroup(gr.GroupSequence())
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

	err := server.ListenAndServeTLS("examples/cert/localhost.pem", "examples/cert/localhost-key.pem")
	if err != nil {
		slog.Error("failed to listen and serve", "error", err)
	}
}

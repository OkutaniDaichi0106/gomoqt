package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	server := moqt.Server{
		Addr: "localhost:4444",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			Certificates:       []tls.Certificate{generateCert()},
			InsecureSkipVerify: true, // TODO: Not recommended for production
		},
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},

		SessionHandlerFunc: handleSession,
	}

	// Serve moq over webtransport
	http.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		err := server.ServeWebTransport(w, r)
		if err != nil {
			slog.Error("failed to serve web transport", "error", err)
		}
	})

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("failed to listen and serve", "error", err)
	}
}

func handleSession(path string, sess moqt.Session) {
	slog.Info("handling a session", slog.String("path", path))

	// Accept a track stream
	stream, err := sess.AcceptTrackStream(context.Background(), func(path moqt.TrackPath) (moqt.Info, error) {
		slog.Info("subscribed to a track", slog.String("track_path", path.String()))
		info := moqt.Info{}
		slog.Info("accepted a subscription", slog.String("track_info", info.String()))
		return info, nil
	})
	if err != nil {
		slog.Error("failed to accept track stream", "error", err)
		return
	}

	seq := moqt.FirstSequence
	for {
		w, err := stream.OpenGroup(seq)
		if err != nil {
			slog.Error("failed to accept group", "error", err)
			break
		}

		slog.Info("group opened", slog.String("group sequence", seq.String()))

		frame := moqt.NewFrame([]byte(fmt.Sprintf("Hello!!. Group: {%s}", seq.String())))
		for {
			err := w.WriteFrame(frame)
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				break
			}

			slog.Info("frame written", slog.String("message", string(frame.CopyBytes())))
		}

		seq = seq.Next()
	}
}

func generateCert() tls.Certificate {
	// Load certificates from the examples/cert directory (project root)
	cert, err := tls.LoadX509KeyPair("examples/cert/localhost.pem", "examples/cert/localhost-key.pem")
	if err != nil {
		panic(err)
	}
	return cert
}

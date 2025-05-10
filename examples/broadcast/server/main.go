package main

import (
	"crypto/tls"
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
	}

	// Serve moq over webtransport
	http.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		sess, err := server.AcceptWebTransport(w, r)
		if err != nil {
			slog.Error("failed to serve web transport", "error", err)
		}
	})

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("failed to listen and serve", "error", err)
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

package main

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	moqt_quic "github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/quic-go/quic-go"
)

// var mux *moqt.TrackMux

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	server := moqt.Server{
		Addr: "localhost:4444",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			Certificates:       []tls.Certificate{generateCert()},
			InsecureSkipVerify: true, // TODO: Not recommended for production
		},
		QUICConfig: (*moqt_quic.Config)(&quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		}),
	}
	// Serve moq over webtransport
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		_, err := server.AcceptWebTransport(w, r, nil)
		if err != nil {
			slog.Error("failed to serve web transport", "error", err)
			return
		}
	})

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("failed to listen and serve", "error", err)
		return
	}
}

func handleSession(path string, sess moqt.Session) {
	if path != "/chat" {
		slog.Error("invalid path", slog.String("path", path))
		return
	}
	defer sess.Terminate(0, "session completed")

	slog.Info("handling a session", slog.String("path", path))

	var wg sync.WaitGroup
	wg.Add(2)
	go runPublisher(sess, &wg)
	go runSubscriber(sess, &wg)

	wg.Wait()

}

func generateCert() tls.Certificate {
	// Load certificates from the examples/cert directory (project root)
	cert, err := tls.LoadX509KeyPair("examples/cert/localhost.pem", "examples/cert/localhost-key.pem")
	if err != nil {
		panic(err)
	}
	return cert
}

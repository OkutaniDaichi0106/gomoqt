package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func main() {
	server := moqt.Server{
		Addr: "localhost:4469",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			Certificates:       []tls.Certificate{generateCert()},
			InsecureSkipVerify: true, // TODO: Not recommended for production
		},
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
		Logger: slog.Default(),
	}

	// Register the relay handler with the default mux
	moqt.PublishFunc(context.Background(), "/server.relay", func(ctx context.Context, tw *moqt.TrackWriter) {
		seq := moqt.GroupSequenceFirst
		for {
			time.Sleep(100 * time.Millisecond)

			gw, err := tw.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			frame := moqt.NewFrame([]byte("FRAME " + seq.String()))
			err = gw.WriteFrame(frame)
			if err != nil {
				gw.CancelWrite(moqt.InternalGroupErrorCode) // TODO: Handle error properly
				slog.Error("failed to write frame", "error", err)
				return
			}

			// TODO: Release the frame after writing
			// This is important to avoid memory leaks

			gw.Close()

			seq = seq.Next()
		}
	})

	moqt.HandleFunc("/relay", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
		sess, err := server.Accept(w, r, nil)
		if err != nil {
			slog.Error("failed to accept session", "error", err)
			return
		}
		anns, err := sess.AcceptAnnounce("/")
		if err != nil {
			slog.Error("failed to open announce stream", "error", err)
			return
		}

		for {
			ann, err := anns.ReceiveAnnouncement(context.Background())
			if err != nil {
				slog.Error("failed to receive announcement", "error", err)
				return
			}

			go func(ann *moqt.Announcement) {
				if !ann.IsActive() {
					return
				}
				tr, err := sess.Subscribe(ann.BroadcastPath(), "index", nil)
				if err != nil {
					slog.Error("failed to open track stream", "error", err)
					return
				}

				moqt.Announce(ann, newRelayHandler(tr))
			}(ann)
		}
	})

	// Serve moq over webtransport
	http.HandleFunc("/relay", func(w http.ResponseWriter, r *http.Request) {
		err := server.ServeWebTransport(w, r)
		if err != nil {
			slog.Error("failed to serve web transport", "error", err)
			return
		}

	})

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("failed to listen and serve", "error", err)
	}
}

func generateCert() tls.Certificate {
	// Find project root by looking for go.mod file
	projectRoot, err := findProjectRoot()
	if err != nil {
		panic(err)
	}

	// Load certificates from the examples/cert directory (project root)
	certPath := filepath.Join(projectRoot, "examples", "cert", "localhost.pem")
	keyPath := filepath.Join(projectRoot, "examples", "cert", "localhost-key.pem")

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		panic(err)
	}
	return cert
}

// findProjectRoot searches for the project root by looking for go.mod file
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func main() {
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
		Logger: slog.Default(),
	}

	moqt.HandleFunc("/echo", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
		mux := moqt.NewTrackMux()
		sess, err := moqt.Accept(w, r, mux)
		if err != nil {
			slog.Error("failed to accept session", "error", err)
			return
		}

		anns, err := sess.AcceptAnnounce("/")
		if err != nil {
			slog.Error("failed to open announce stream", "error", err)
		}

		for {
			ann, err := anns.ReceiveAnnouncement(context.Background())
			if err != nil {
				slog.Error("failed to receive announcements", "error", err)
				break
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

				mux.Announce(ann, moqt.TrackHandlerFunc(func(tw *moqt.TrackWriter) {
					defer tr.Close()

					for {
						gr, err := tr.AcceptGroup(context.Background())
						if err != nil {
							slog.Error("failed to accept group", "error", err)
							return
						}

						go func(gr *moqt.GroupReader) {
							defer gr.CancelRead(moqt.InternalGroupErrorCode)

							gw, err := tw.OpenGroup(gr.GroupSequence())
							if err != nil {
								slog.Error("failed to open group", "error", err)
								return
							}
							defer gw.Close()

							for frame := range gr.Frames(nil) {
								err = gw.WriteFrame(frame)
								if err != nil {
									slog.Error("failed to write frame", "error", err)
									return
								}

								// TODO: Release the frame after writing
								// This is important to avoid memory leaks
							}
						}(gr)

					}
				}))
			}(ann)
		}

	})

	// Serve moq over webtransport
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		err := server.ServeWebTransport(w, r)
		if err != nil {
			slog.Error("failed to serve moq over webtransport", "error", err)
		}
	})

	err := server.ListenAndServeTLS("../../cert/localhost.pem", "../../cert/localhost-key.pem")
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

package main

import (
	"context"
	"crypto/tls"
	"io"
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
		Addr: "moqt.example.com:9000",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			Certificates:       []tls.Certificate{generateCert()},
			InsecureSkipVerify: true, // TODO: Not recommended for production
		},
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
		Config: &moqt.Config{
			CheckHTTPOrigin: func(r *http.Request) bool {
				return true // TODO: Implement proper origin check
			},
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	moqt.HandleFunc("/interop", func(w moqt.ResponseWriter, r *moqt.Request) {
		sess, err := server.Accept(w, r, nil)
		if err != nil {
			w.Reject(moqt.ProtocolViolationErrorCode)
			slog.Error("failed to accept session", "error", err)
			return
		}

		anns, err := sess.AcceptAnnounce("/")
		if err != nil {
			slog.Error("failed to open announce stream", "error", err)
			return
		}

		ann, err := anns.ReceiveAnnouncement(context.Background())
		if err != nil {
			slog.Error("failed to receive announcements", "error", err)
			return
		}

		if !ann.IsActive() {
			return
		}

		slog.Info("Received announcement", "path", ann.BroadcastPath())

		tr, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
		if err != nil {
			slog.Error("failed to open track stream", "error", err)
			return
		}

		for {
			gr, err := tr.AcceptGroup(context.Background())
			if err != nil {
				slog.Error("failed to accept group", "error", err)
				return
			}

			slog.Info("Accepted group", "group_sequence", gr.GroupSequence())

			go func(gr *moqt.GroupReader) {
				frame := moqt.NewFrame(nil)
				for {
					err := gr.ReadFrame(frame)
					if err != nil {
						if err == io.EOF {
							return
						}
						slog.Error("failed to read frame", "error", err)
						return
					}

					slog.Info("Received a frame", "frame", string(frame.Bytes()))

					// TODO: Release the frame after processing
					// This is important to avoid memory leaks
				}
			}(gr)
		}
	})

	http.HandleFunc("/interop", func(w http.ResponseWriter, r *http.Request) {
		err := server.ServeWebTransport(w, r)
		if err != nil {
			slog.Error("failed to serve moq over webtransport", "error", err)
			return
		}
	})

	moqt.PublishFunc(context.Background(), "/server.interop", func(ctx context.Context, tw *moqt.TrackWriter) {
		seq := moqt.GroupSequenceFirst
		for range 10 {
			group, err := tw.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				break
			}

			slog.Info("Opened group successfully", "group_sequence", group.GroupSequence())

			frame := moqt.NewFrame([]byte("Hello from interop server in Go!"))
			err = group.WriteFrame(frame)
			if err != nil {
				group.CancelWrite(moqt.InternalGroupErrorCode) // TODO: Handle error properly
				slog.Error("failed to write frame", "error", err)
				break
			}

			slog.Info("Sent frame successfully", "frame", string(frame.Bytes()))

			group.Close()

			seq = seq.Next()

			time.Sleep(100 * time.Millisecond)
		}

		tw.Close()
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

	// Load certificates from the interop/cert directory (project root)
	certPath := filepath.Join(projectRoot, "interop", "server", "moqt.example.com.pem")
	keyPath := filepath.Join(projectRoot, "interop", "server", "moqt.example.com-key.pem")

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

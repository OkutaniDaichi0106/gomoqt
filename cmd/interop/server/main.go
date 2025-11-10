package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func main() {
	slog.Info("[Server] Starting server on moqt.example.com:9000")

	server := moqt.Server{
		Addr: "moqt.example.com:9000", // TODO: Use given address
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
	}

	// Serve MOQ over WebTransport
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := server.HandleWebTransport(w, r)
		if err != nil {
			slog.Error("[Server] failed to serve moq over webtransport: " + err.Error())
			return
		}
	})

	moqt.HandleFunc("/", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
		slog.Info("[Server] Accepting session")

		// Create a custom mux for this session
		mux := moqt.NewTrackMux()

		// Register the handler BEFORE accepting session
		slog.Info("[Server] Registering /interop/server handler")
		mux.PublishFunc(context.Background(), "/interop/server", func(tw *moqt.TrackWriter) {
			slog.Info("[Server] Client subscribed, sending data...")

			group, err := tw.OpenGroup(moqt.GroupSequenceFirst)
			if err != nil {
				slog.Error("[Server] failed to open group: " + err.Error())
				return
			}
			defer group.Close()

			frame := moqt.NewFrame(1024)
			frame.Write([]byte("HELLO"))

			err = group.WriteFrame(frame)
			if err != nil {
				slog.Error("[Server] failed to write frame: " + err.Error())
				return
			}

			slog.Info("[Server] Data sent to client")
		})

		// Accept the session with the custom mux
		sess, err := moqt.Accept(w, r, mux)
		if err != nil {
			w.Reject(moqt.ProtocolViolationErrorCode)
			slog.Error("[Server] failed to accept session: " + err.Error())
			return
		}

		// Don't use defer for Terminate since we want to shutdown the server first
		var wg sync.WaitGroup

		// Discover announcements from client
		wg.Add(1)
		go func() {
			defer wg.Done()

			slog.Info("[Server] Starting to accept client announcements...")
			anns, err := sess.AcceptAnnounce("/")
			if err != nil {
				slog.Error("[Server] failed to accept announce: " + err.Error())
				return
			}
			defer anns.Close()

			slog.Info("[Server] Waiting for announcement from client...")
			ann, err := anns.ReceiveAnnouncement(context.Background())
			if err != nil {
				slog.Error("[Server] failed to receive announcement: " + err.Error())
				return
			}

			slog.Info("[Server] Discovered broadcast: " + string(ann.BroadcastPath()))

			track, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
			if err != nil {
				slog.Error("[Server] failed to subscribe: " + err.Error())
				return
			}
			defer track.Close()

			slog.Info("[Server] Subscribed to a track")

			group, err := track.AcceptGroup(context.Background())
			if err != nil {
				slog.Error("[Server] failed to accept group: " + err.Error())
				return
			}
			defer group.CancelRead(moqt.InternalGroupErrorCode)

			slog.Info("[Server] Received a group")

			frame := moqt.NewFrame(1024)

			err = group.ReadFrame(frame)
			if err != nil {
				slog.Error("[Server] failed to read frame: " + err.Error())
				return
			}

			slog.Info("[Server] Received frame: " + string(frame.Body()))
		}()

		// Wait for all operations to complete
		slog.Info("[Server] Waiting for operations to complete...")
		wg.Wait()

		slog.Info("[Server] Operations completed")

		// Terminate session before shutting down server
		sess.CloseWithError(moqt.NoError, "no error")

		// Trigger server shutdown after handling one session
		// Give time for session cleanup and final data transmission
		time.AfterFunc(100*time.Millisecond, func() {
			if shutdownErr := server.Shutdown(context.Background()); shutdownErr != nil {
				slog.Error("[Server] Shutdown error: " + shutdownErr.Error())
				os.Exit(1)
			}
			os.Exit(0)
		})
	})

	err := server.ListenAndServe()
	// Ignore expected shutdown errors
	if err != nil && err != moqt.ErrServerClosed && err.Error() != "quic: server closed" {
		slog.Error("[Server] failed to listen and serve: " + err.Error())
		os.Exit(1)
	}
}

func generateCert() tls.Certificate {
	// Find project root by looking for go.mod file
	root, err := findRootDir()
	if err != nil {
		slog.Error("[Server] failed to find project root: " + err.Error())
		os.Exit(1)
	}

	// Load certificates from the interop/cert directory (project root)
	certPath := filepath.Join(root, "cmd", "interop", "server", "moqt.example.com.pem")
	keyPath := filepath.Join(root, "cmd", "interop", "server", "moqt.example.com-key.pem")

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		slog.Error("[Server] failed to load certificates: " + err.Error())
		os.Exit(1)
	}
	return cert
}

// findRootDir searches for the project root by looking for go.mod file
func findRootDir() (string, error) {
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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func main() {
	if err := mkcert(); err != nil {
		fmt.Fprintf(os.Stderr, "[Server] Setting up certificates...failed\n  Error: %v\n", err)
		return
	}

	// Print startup message directly
	fmt.Println("[Server] ✓ Started on 127.0.0.1:9000")

	server := moqt.Server{
		Addr: "127.0.0.1:9000", // Use localhost
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
			fmt.Fprintf(os.Stderr, "[Server] failed to serve moq over webtransport: %v\n", err)
			return
		}
	})

	moqt.HandleFunc("/", func(w moqt.SetupResponseWriter, r *moqt.SetupRequest) {
		fmt.Println("[Server] Setting up session")

		// Create a custom mux for this session
		mux := moqt.NewTrackMux()

		// Accept the session with the custom mux
		fmt.Print("[Server] Accepting session...")
		sess, err := moqt.Accept(w, r, mux)
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			w.Reject(moqt.ProtocolViolationErrorCode)
			return
		}
		fmt.Println("...ok")

		var publishedWG sync.WaitGroup
		publishedWG.Add(1)
		mux.PublishFunc(context.Background(), "/interop/server", func(tw *moqt.TrackWriter) {
			defer publishedWG.Done()

			fmt.Println("[Server] Serving broadcast...")

			fmt.Print("[Server] Opening group...")
			group, err := tw.OpenGroup(moqt.GroupSequenceFirst)
			if err != nil {
				fmt.Printf("...failed\n  Error: %v\n", err)
				return
			}
			defer group.Close()
			fmt.Println("...ok")

			fmt.Print("[Server] Writing frame to client...")
			frame := moqt.NewFrame(1024)
			frame.Write([]byte("HELLO"))

			err = group.WriteFrame(frame)
			if err != nil {
				fmt.Printf("...failed\n  Error: %v\n", err)
				return
			}
			fmt.Println("...ok")

			fmt.Println("[Server] ✓ Data sent to client")
		})

		// Wait for the broadcast to be published
		publishedWG.Wait()

		// Discover broadcasts from client
		fmt.Print("[Server] Accepting client announcements...")
		anns, err := sess.AcceptAnnounce("/")
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		defer anns.Close()
		fmt.Println("...ok")

		fmt.Print("[Server] Receiving announcement...")
		ann, err := anns.ReceiveAnnouncement(context.Background())
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		fmt.Println("...ok")

		fmt.Printf("[Server] Discovered broadcast: %s\n", string(ann.BroadcastPath()))

		fmt.Print("[Server] Subscribing to broadcast...")
		track, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		defer track.Close()
		fmt.Println("...ok")

		fmt.Print("[Server] Accepting group...")
		group, err := track.AcceptGroup(context.Background())
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		fmt.Println("...ok")

		fmt.Print("[Server] Reading frame from client...")
		frame := moqt.NewFrame(1024)

		err = group.ReadFrame(frame)
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		fmt.Println("...ok")

		fmt.Printf("[Server] ✓ Received data from client: %s\n", string(frame.Body()))

		fmt.Print("[Server] Closing session...")
		sess.CloseWithError(moqt.NoError, "no error")
		fmt.Println("...ok")

		fmt.Println("[Server] Operation completed successfully")
	})

	fmt.Println("[Server] Listening...")
	err := server.ListenAndServe()
	// Ignore expected shutdown errors
	if err != nil && err != moqt.ErrServerClosed && err.Error() != "quic: server closed" {
		fmt.Fprintf(os.Stderr, "[Server] failed to listen and serve: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[Server] Exited normally")
}

func generateCert() tls.Certificate {
	// Find project root by looking for go.mod file
	root, err := findRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Server] failed to find project root: %v\n", err)
		os.Exit(1)
	}

	// Load certificates from the interop/cert directory (project root)
	certPath := filepath.Join(root, "cmd", "interop", "server", "moqt.example.com.pem")
	keyPath := filepath.Join(root, "cmd", "interop", "server", "moqt.example.com-key.pem")

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Server] failed to load certificates: %v\n", err)
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

func mkcert() error {
	// Resolve paths from project root so the program works regardless of CWD
	root, err := findRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Server] failed to find project root for mkcert: %v\n", err)
		return err
	}

	serverCertPath := filepath.Join(root, "cmd", "interop", "server", "moqt.example.com.pem")

	// Check if server certificates exist
	if _, err := os.Stat(serverCertPath); os.IsNotExist(err) {
		fmt.Print("[Server] Setting up certificates...")
		cmd := exec.Command("mkcert", "moqt.example.com")
		// Ensure mkcert runs in the server directory where cert files should be generated
		cmd.Dir = filepath.Join(root, "cmd", "interop", "server")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return err
		}
		fmt.Println("...ok")
	}
	return nil
}

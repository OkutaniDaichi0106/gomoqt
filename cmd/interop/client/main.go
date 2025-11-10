package main

import (
	"context"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	client := &moqt.Client{}

	// Create a custom mux for this session (like http.NewServeMux())
	mux := moqt.NewTrackMux()

	// Publish to the interop broadcast so server can discover it
	// Register the handler BEFORE dialing so it's ready when server requests announcements
	slog.Info("[Client] Registering /interop/client handler")
	mux.PublishFunc(context.Background(), "/interop/client", func(tw *moqt.TrackWriter) {
		slog.Info("[Client] Server subscribed, sending data...")

		group, err := tw.OpenGroup(moqt.GroupSequenceFirst)
		if err != nil {
			slog.Error("[Client] failed to open group: " + err.Error())
			return
		}
		defer group.Close()

		frame := moqt.NewFrame(1024)
		frame.Write([]byte("HELLO"))

		err = group.WriteFrame(frame)
		if err != nil {
			slog.Error("[Client] failed to write frame: " + err.Error())
			return
		}

		slog.Info("[Client] Data sent to server")
	})

	sess, err := client.Dial(context.Background(), "https://moqt.example.com:9000", mux)
	if err != nil {
		slog.Error("[Client] failed to dial: " + err.Error())
		return
	}

	slog.Info("[Client] Connected to server")

	defer sess.CloseWithError(moqt.NoError, "no error")

	var wg sync.WaitGroup

	// Discover announcements and receive data from server
	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("[Client] Starting to accept server announcements...")
		anns, err := sess.AcceptAnnounce("/")
		if err != nil {
			slog.Error("[Client] failed to accept announce: " + err.Error())
			return
		}
		defer anns.Close()

		slog.Info("[Client] Waiting for announcement from server...")
		ann, err := anns.ReceiveAnnouncement(context.Background())
		if err != nil {
			slog.Error("[Client] failed to receive announcement: " + err.Error())
			return
		}

		slog.Info("[Client] Discovered broadcast: " + string(ann.BroadcastPath()))

		// Subscribe to the interop broadcast
		track, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
		if err != nil {
			slog.Error("[Client] failed to subscribe: " + err.Error())
			return
		}
		defer track.Close()

		slog.Info("[Client] Subscribed to a track")

		group, err := track.AcceptGroup(context.Background())
		if err != nil {
			slog.Error("[Client] failed to accept group: " + err.Error())
			return
		}
		defer group.CancelRead(moqt.InternalGroupErrorCode)

		slog.Info("[Client] Received a group")

		frame := moqt.NewFrame(1024)

		err = group.ReadFrame(frame)
		if err != nil {
			slog.Error("[Client] failed to read frame: " + err.Error())
			return
		}

		slog.Info("[Client] Received frame: " + string(frame.Body()))
	}()

	// Wait for all operations to complete
	slog.Info("[Client] Waiting for operations to complete...")
	wg.Wait()

	slog.Info("[Client] Operations completed")
}

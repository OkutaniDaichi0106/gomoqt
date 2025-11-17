package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	addr := flag.String("addr", "https://moqt.example.com:9000", "server URL for MOQ (https://host:port)")
	flag.Parse()
	client := &moqt.Client{}

	// Create a custom mux for this session (like http.NewServeMux())
	mux := moqt.NewTrackMux()

	sess, err := client.Dial(context.Background(), *addr, mux)
	if err != nil {
		fmt.Printf("[Client] Connecting to server...failed\n  Error: %v\n", err)
		return
	}
	defer sess.CloseWithError(moqt.NoError, "no error")

	fmt.Println("[Client] ✓ Connected to server")

	// Step 1: Accept announcements from server
	fmt.Print("[Client] Accepting server announcements...")
	anns, err := sess.AcceptAnnounce("/")
	if err != nil {
		fmt.Printf("...failed\n  Error: %v\n", err)
		return
	}
	defer anns.Close()

	fmt.Println("...ok")

	fmt.Print("[Client] Receiving announcement...")
	announceCtx, cancelReceive := context.WithTimeout(sess.Context(), 5*time.Second)
	defer cancelReceive()
	ann, err := anns.ReceiveAnnouncement(announceCtx)
	if err != nil {
		fmt.Printf("...failed\n  Error: %v\n", err)
		return
	}
	fmt.Println("...ok")

	fmt.Printf("[Client] Discovered broadcast: %s\n", string(ann.BroadcastPath()))

	// Step 2: Subscribe to the server's broadcast and receive data (in goroutine)
	fmt.Print("[Client] Subscribing to server broadcast...")
	track, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
	if err != nil {
		fmt.Printf("...failed\n  Error: %v\n", err)
		return
	}
	defer track.Close()

	fmt.Println("...ok")

	fmt.Print("[Client] Accepting group from server...")
	groupCtx, cancelGroup := context.WithTimeout(sess.Context(), 5*time.Second)
	defer cancelGroup()
	group, err := track.AcceptGroup(groupCtx)
	if err != nil {
		fmt.Printf("...failed\n  Error: %v\n", err)
		return
	}
	fmt.Println("...ok")

	fmt.Print("[Client] Reading the first frame from server...")
	frame := moqt.NewFrame(1024)

	err = group.ReadFrame(frame)
	if err != nil {
		fmt.Printf("...failed\n  Error: %v\n", err)
		return
	}
	fmt.Println("...ok")

	fmt.Printf("[Client] ✓ Received data from server: %s\n", string(frame.Body()))

	var publishedWG sync.WaitGroup
	// Publish to the interop broadcast so server can discover it
	// Register the handler BEFORE dialing so it's ready when server requests announcements
	publishedWG.Add(1)
	mux.PublishFunc(context.Background(), "/interop/client", func(tw *moqt.TrackWriter) {
		defer publishedWG.Done()

		fmt.Print("[Client] Opening group...")
		group, err := tw.OpenGroup(moqt.GroupSequenceFirst)
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		defer group.Close()
		fmt.Println("...ok")

		fmt.Print("[Client] Writing frame to server...")
		frame := moqt.NewFrame(1024)
		frame.Write([]byte("HELLO"))

		err = group.WriteFrame(frame)
		if err != nil {
			fmt.Printf("...failed\n  Error: %v\n", err)
			return
		}
		fmt.Println("...ok")

		fmt.Println("[Client] ✓ Data sent to server")
	})

	publishedWG.Wait()

	fmt.Println("[Client] Operations completed")
}

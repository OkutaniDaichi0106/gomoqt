package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"golang.org/x/term"
)

var name *string

var history []string
var historyMutex sync.Mutex

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	name = flag.String("name", "unknown", "your name in chat")
	flag.Parse()

	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), state)

	runClient()
}

func runClient() {
	client := moqt.Client{}

	sess, _, err := client.Dial("https://localhost:4444/chat", context.Background())
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}
	defer sess.Terminate(nil)

	// track := moqt.BuildTrack()
	// mux.Handle("/kiu", track)

	var wg sync.WaitGroup
	wg.Add(2)
	go runSubscriber(sess, &wg)
	go runPublisher(sess, &wg)

	wg.Wait()
}

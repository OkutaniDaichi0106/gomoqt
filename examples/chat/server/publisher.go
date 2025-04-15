package main

import (
	"context"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func runPublisher(sess moqt.Session, wg *sync.WaitGroup) {
	defer wg.Done()

	// Accept an announce stream
	annstr, annconf, err := sess.AcceptAnnounceStream(context.Background())
	if err != nil {
		slog.Error("failed to accept announce stream", "error", err)
		return
	}

	moqt.ServeAnnouncements(annstr, annconf)

	for {
		wg.Add(1)
		go serveTrack(sess, wg)
	}
}

func serveTrack(sess moqt.Session, wg *sync.WaitGroup) {
	defer wg.Done()
	// Accept a track stream
	stream, subconf, err := sess.AcceptTrackStream(context.Background())
	if err != nil {
		slog.Error("failed to accept track stream", "error", err)
		return
	}

	moqt.ServeTrack(stream, subconf)
}

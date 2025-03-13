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
	annstr, err := sess.AcceptAnnounceStream(context.Background(), func(ac moqt.AnnounceConfig) error {
		slog.Debug("accepted an announce stream", slog.String("config", ac.String()))
		return nil
	})
	if err != nil {
		slog.Error("failed to accept announce stream", "error", err)
		return
	}

	mux.ServeAnnouncement(annstr)

	for {
		wg.Add(1)
		go serveTrack(sess, wg)
	}
}

func serveTrack(sess moqt.Session, wg *sync.WaitGroup) {
	defer wg.Done()
	// Accept a track stream
	stream, err := sess.AcceptTrackStream(context.Background(), func(sc moqt.SubscribeConfig) (moqt.Info, error) {
		slog.Debug("subscribed to a track", slog.String("config", sc.String()))
		infoReq := moqt.InfoRequest{
			TrackPath: sc.TrackPath,
		}
		infoCh := make(chan moqt.Info, 1)

		// Find the track info
		mux.ServeInfo(infoCh, infoReq)

		info := <-infoCh
		slog.Debug("accepted a subscription", slog.String("info", info.String()))
		return info, nil
	})
	if err != nil {
		slog.Error("failed to accept track stream", "error", err)
		return
	}

	mux.ServeTrack(stream, stream.SubscribeConfig())
}

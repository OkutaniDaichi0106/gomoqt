package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func runSubscriber(sess moqt.Session, wg *sync.WaitGroup) {
	defer wg.Done()

	annstr, err := sess.OpenAnnounceStream(moqt.AnnounceConfig{
		TrackPattern: "/kiu",
	})
	if err != nil {
		slog.Error("failed to open announce stream", "error", err)
		return
	}

	// Listen to announcements
	annCh := make(chan *moqt.Announcement)
	go func() {
		for {
			anns, err := annstr.NextAnnouncements(context.Background())
			if err != nil {
				slog.Error("failed to read announcements", "error", err)
				close(annCh)
				return
			}
			for _, ann := range anns {
				annCh <- ann
			}
		}
	}()

	// Subscribe to tracks
	for ann := range annCh {
		slog.Debug("received an announcement", slog.String("announcement", ann.String()))
		go func(ann *moqt.Announcement) {
			// Subscribe to the track
			config := moqt.SubscribeConfig{
				TrackPath: ann.TrackPath(),
			}
			info, stream, err := sess.OpenTrackStream(config)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}
			slog.Info("subscribed to a track", slog.String("config", config.String()), slog.String("info", info.String()))

			name := strings.TrimPrefix(string(ann.TrackPath()), "/kiu/")
			slog.Info("a new member joined", slog.String("name", name))

			for {
				r, err := stream.AcceptGroup(context.Background())
				if err != nil {
					slog.Error("failed to accept group", "error", err)
					break
				}

				//
				typingRaw := appendHistory(name + " is typing...")

				frame, err := r.ReadFrame()
				if err != nil {
					slog.Error("failed to read a frame", "error", err)
					break
				}

				message := string(frame.CopyBytes())

				if message == "" {
					deleteHistory(typingRaw)
				} else {
					updateHistory(typingRaw, name+": "+message)
				}

				r.CancelRead(nil)
			}
		}(ann)
	}

	// subconfig := moqt.SubscribeConfig{
	// 	TrackPath: moqt.TrackPath("/kiu"),
	// }
	// info, stream, err := sess.OpenTrackStream(subconfig)
	// if err != nil {
	// 	slog.Error("failed to open track stream", "error", err,
	// 	return
	// }

	// slog.Info("subscribed to a track", slog.String("config", subconfig.String()), slog.String("info", info.String()))

	// for {
	// 	r, err := stream.AcceptGroup(context.Background())
	// 	if err != nil {
	// 		slog.Error("failed to accept group", "error", err,
	// 		break
	// 	}

	// 	frame, err := r.ReadFrame()
	// 	if err != nil {
	// 		slog.Error("failed to read frame", "error", err,
	// 		break
	// 	}

	// 	fmt.Printf("")
	// }
}

func redraw(input string) {
	fmt.Print("\033[H\033[2J")
	for _, line := range history {
		fmt.Println(line)
	}

	fmt.Print("> " + input)
}

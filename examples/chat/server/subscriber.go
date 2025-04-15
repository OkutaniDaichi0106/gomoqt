package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func runSubscriber(sess moqt.Session, wg *sync.WaitGroup) {
	defer wg.Done()

	annstr, err := sess.OpenAnnounceStream(&moqt.AnnounceConfig{
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
			anns, err := annstr.ReceiveAnnouncements(context.Background())
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
			path := ann.TrackPath()
			info, stream, err := sess.OpenTrackStream(path, nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}
			slog.Info("subscribed to a track",
				"track_path", path,
				"info", info,
			)

			track := moqt.NewTrackBuffer(ann.TrackPath(), info, 5*time.Second)

			moqt.Handle(path, track)

			for {
				r, err := stream.AcceptGroup(context.Background())
				if err != nil {
					slog.Error("failed to accept group", "error", err)
					break
				}

				w, err := track.OpenGroup(r.GroupSequence())
				if err != nil {
					slog.Error("failed to open group", "error", err)
					break
				}

				// Just read one frame
				frame, err := r.ReadFrame()
				if err != nil {
					slog.Error("failed to read a frame", "error", err)
					break
				}
				r.CancelRead(nil)

				// Write the frame to the track buffer
				err = w.WriteFrame(frame)
				if err != nil {
					slog.Error("failed to write a frame", "error", err)
					break
				}

				slog.Debug("successfully wrote a frame", slog.String("groupSequence", r.GroupSequence().String()))
			}
		}(ann)
	}

}

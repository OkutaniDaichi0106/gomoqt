package main

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	client := moqt.Client{
		Logger: slog.Default(),
	}

	sess, err := client.Dial(context.Background(), "https://localhost:4469/broadcast", nil)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}

	//
	annRecv, err := sess.AcceptAnnounce("/")
	if err != nil {
		slog.Error("failed to open announce stream", "error", err)
		return
	}
	defer annRecv.Close()

	for {
		ann, err := annRecv.ReceiveAnnouncement(context.Background())
		if err != nil {
			slog.Error("failed to receive announcement", "error", err)
			break
		}

		slog.Info("received announcement", "announcement", ann)

		go func(ann *moqt.Announcement) {
			if !ann.IsActive() {
				return
			}

			tr, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}
			defer tr.Close()

			for {
				gr, err := tr.AcceptGroup(context.Background())
				if err != nil {
					slog.Error("failed to accept group", "error", err)
					return
				}

				go func(gr *moqt.GroupReader) {
					defer gr.CancelRead(moqt.InternalGroupErrorCode)

					for frame := range gr.Frames(nil) {
						slog.Info("received a frame", "frame", string(frame.Body()))

						// TODO: Release the frame after processing
						// This is important to avoid memory leaks
					}
				}(gr)

			}

		}(ann)
	}
}

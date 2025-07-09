package main

import (
	"context"
	"io"
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
	annRecv, err := sess.OpenAnnounceStream("/")
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

			sub, err := sess.OpenTrackStream(ann.BroadcastPath(), "", nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}
			defer sub.Controller.Close()

			for {
				gr, err := sub.TrackReader.AcceptGroup(context.Background())
				if err != nil {
					slog.Error("failed to accept group", "error", err)
					return
				}

				go func(gr moqt.GroupReader) {
					for {
						frame, err := gr.ReadFrame()
						if err != nil {
							if err == io.EOF {
								return
							}
							slog.Error("failed to read frame", "error", err)
							gr.CancelRead(moqt.InternalGroupErrorCode)
							return
						}

						slog.Info("received a frame", "frame", string(frame.CopyBytes()))

						// TODO: Release the frame after processing
						// This is important to avoid memory leaks
						frame.Release()
					}
				}(gr)

			}

		}(ann)
	}
}

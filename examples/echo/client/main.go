package main

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	moqt.HandleFunc(context.Background(), "client.echo", func(pub *moqt.Publisher) {
		if pub.TrackName != "index" {
			return
		}

		seq := moqt.GroupSequenceFirst
		for {
			gw, err := pub.TrackWriter.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			err = gw.WriteFrame(moqt.NewFrame([]byte("FRAME " + seq.String())))
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				return
			}

			gw.Close()

			seq = seq.Next()
		}
	})

	client := moqt.Client{}

	sess, err := client.Dial(context.Background(), "https://localhost:4444/broadcast", nil)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}

	annstr, err := sess.OpenAnnounceStream("/client")
	if err != nil {
		slog.Error("failed to open announce stream", "error", err)
		return
	}
	for {
		announcements, err := annstr.ReceiveAnnouncements(context.TODO())
		if err != nil {
			slog.Error("failed to receive announcements", "error", err)
			return
		}
		for _, ann := range announcements {
			go func(ann *moqt.Announcement) {
				sub, err := sess.OpenTrackStream(ann.BroadcastPath(), "index", nil)
				if err != nil {
					slog.Error("failed to open track stream", "error", err)
					return
				}

				for {
					gr, err := sub.TrackReader.AcceptGroup(context.Background())
					if err != nil {
						slog.Error("failed to accept group", "error", err)
						return
					}

					for {
						f, err := gr.ReadFrame()
						if err != nil {
							slog.Error("failed to read frame", "error", err)
							return
						}

						slog.Info("received frame", "frame", string(f.CopyBytes()))
					}
				}
			}(ann)

		}
	}

}

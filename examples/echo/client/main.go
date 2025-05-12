package main

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	client := moqt.Client{}

	sess, _, err := client.Dial("https://localhost:4444/broadcast", context.Background())
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}

	annstr, err := sess.OpenAnnounceStream(&moqt.AnnounceConfig{TrackPattern: "/data"})
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
			info, tr, err := sess.OpenTrackStream(ann.TrackPath(), nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}

		}
	}

}

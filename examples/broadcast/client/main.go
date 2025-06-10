package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	client := moqt.Client{}

	sess, err := client.Dial(context.Background(), "https://localhost:4444/broadcast", nil)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}

	info, stream, err := sess.OpenTrackStream("/chat", nil)
	if err != nil {
		slog.Error("failed to open track stream", "error", err)
		return
	}

	slog.Info("track stream opened", slog.String("info", info.String()))

	for {
		r, err := stream.AcceptGroup(context.Background())
		if err != nil {
			slog.Error("failed to accept group", "error", err)
			break
		}

		slog.Info("group accepted", slog.String("group sequence", r.GroupSequence().String()))

		for {
			frame, err := r.ReadFrame()
			if err != nil {
				slog.Error("failed to accept frame", "error", err)
				break
			}

			slog.Info("frame accepted", slog.String("message", string(frame.CopyBytes())))
		}
	}
}

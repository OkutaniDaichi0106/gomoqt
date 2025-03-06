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

	sess, _, err := client.Dial("https://localhost:4444/push", context.Background())
	if err != nil {
		slog.Error("failed to dial", slog.String("error", err.Error()))
		return
	}

	info, stream, err := sess.OpenTrackStream(moqt.SubscribeConfig{
		TrackPath: moqt.TrackPath("/text"),
	})
	if err != nil {
		slog.Error("failed to open track stream", slog.String("error", err.Error()))
		return
	}

	slog.Info("track stream opened", slog.String("info", info.String()))

	for {
		r, err := stream.AcceptGroup(context.Background())
		if err != nil {
			slog.Error("failed to accept group", slog.String("error", err.Error()))
			break
		}

		slog.Info("group accepted", slog.String("group sequence", r.GroupSequence().String()))

		for {
			frame, err := r.ReadFrame()
			if err != nil {
				slog.Error("failed to accept frame", slog.String("error", err.Error()))
				break
			}

			slog.Info("frame accepted", slog.String("message", string(frame.CopyBytes())))
		}
	}
}

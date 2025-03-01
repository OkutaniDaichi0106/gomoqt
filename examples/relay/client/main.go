package main

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	client := moqt.Client{}

	sess, _, err := client.Dial("https://localhost:4444/relay", context.Background())
	if err != nil {
		slog.Error("failed to dial", slog.String("error", err.Error()))
		return
	}
}

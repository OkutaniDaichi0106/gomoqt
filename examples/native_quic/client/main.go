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

	_, err := client.Dial(context.Background(), "moqt://localhost:4469/nativequic", nil)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}
}

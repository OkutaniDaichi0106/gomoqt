package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"time"

	moqt "github.com/OkutaniDaichi0106/gomoqt"
	"github.com/quic-go/quic-go"
)

var echoTrackPrefix = "japan/kyoto"
var echoTrackPath = "japan/kyoto/kiu/text"

func main() {
	/*
	 * Set Log Level to "DEBUG"
	 */
	// logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	// slog.SetDefault(logger)

	c := moqt.Client{
		TLSConfig:  &tls.Config{},
		QUICConfig: &quic.Config{},
	}

	sess, _, err := c.Dial("https://localhost:8080/path", context.Background())
	if err != nil {
		slog.Error(err.Error())
		return
	}

	go func() {
		pub := sess.Publisher()

		pub.Announce(moqt.Announcement{TrackPathSuffix: "japan/kyoto/kiu"})
		moqt.NewTrack()
		for {
			stream, err := pub.OpenDataStream(track, moqt.Group{})
			if err != nil {
				slog.Error("failed to open a data stream", slog.String("error", err.Error()))
				return
			}

			_, err = stream.Write([]byte("HELLO!!"))
			if err != nil {
				slog.Error("failed to write data", slog.String("error", err.Error()))
				return
			}

			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		sub := sess.Subscriber()

		sub.Interest(moqt.Interest{TrackPrefix: "japan/kyoto"})

	}()

}

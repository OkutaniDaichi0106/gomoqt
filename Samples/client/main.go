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

	// Get a setup request
	req := moqt.SetupRequest{
		URL: "https://localhost:8080/path",
	}

	// Dial to the server with the setup request
	sess, _, err := c.Dial(req, context.Background())
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// Run a publisher
	go func() {
		pub := sess.Publisher()

		interest, err := pub.AcceptInterest(context.Background())
		if err != nil {
			slog.Error("failed to accept an interest", slog.String("error", err.Error()))
			return
		}

		tracks := moqt.NewTracks([]moqt.Track{{
			TrackPath:     echoTrackPath,
			TrackPriority: 0,
			GroupOrder:    0,
			GroupExpires:  1 * time.Second,
		}, {
			TrackPath:     "japan/kyoto/kiu/image",
			TrackPriority: 0,
			GroupOrder:    0,
			GroupExpires:  1 * time.Second,
		}})

		// Announce the tracks
		interest.Announce(tracks)

		subscription, err := pub.AcceptSubscription(context.Background())
		if err != nil {
			slog.Error("failed to accept a subscription", slog.String("error", err.Error()))
			return
		}

		if subscription.TrackPath != echoTrackPath {
			slog.Error("failed to get a track path", slog.String("error", "track path is invalid"))
			return
		}

		for sequence := moqt.GroupSequence(0); sequence < 30; sequence++ {
			stream, err := subscription.OpenDataStream(sequence, 0)
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

	// Run a subscriber
	func() {
		sub := sess.Subscriber()

		interest, err := sub.Interest(moqt.Interest{TrackPrefix: echoTrackPrefix})
		if err != nil {
			slog.Error("failed to get an interest", slog.String("error", err.Error()))
			return
		}

		tracks, err := interest.NextActiveTracks()
		if err != nil {
			slog.Error("failed to get active tracks", slog.String("error", err.Error()))
			return
		}

		if _, ok := tracks.Get(echoTrackPath); !ok {
			slog.Error("failed to get a track", slog.String("error", "track is not found"))
			return
		}

		subscription, err := sub.Subscribe(moqt.Subscription{
			Track: moqt.Track{
				TrackPath:     echoTrackPath,
				TrackPriority: 0,
				GroupOrder:    0,
				GroupExpires:  1 * time.Second,
			},
		})
		if err != nil {
			slog.Error("failed to subscribe", slog.String("error", err.Error()))
			return
		}

		for {
			stream, err := subscription.AcceptDataStream(context.Background())
			if err != nil {
				slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
				return
			}

			buf := make([]byte, 1024)
			n, err := stream.Read(buf)
			if err != nil {
				slog.Error("failed to read data", slog.String("error", err.Error()))
				return
			}

			slog.Debug("received data", slog.String("data", string(buf[:n])))
		}
	}()

}

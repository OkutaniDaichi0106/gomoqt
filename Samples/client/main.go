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

		annstr, err := sess.AcceptAnnounceStream(context.Background())
		if err != nil {
			slog.Error("failed to accept an interest", slog.String("error", err.Error()))
			return
		}

		// Announce
		announcements := []moqt.Announcement{
			{
				TrackPath: echoTrackPath,
			},
			{
				TrackPath: "japan/kyoto/kiu/audio", //
			},
		}
		err = annstr.SendAnnouncement(announcements)
		if err != nil {
			slog.Error("failed to announce", slog.String("error", err.Error()))
			return
		}

		substr, err := sess.AcceptSubscribeStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a subscription", slog.String("error", err.Error()))
			return
		}

		if substr.Subscription().TrackPath != echoTrackPath {
			slog.Error("failed to get a track path", slog.String("error", "track path is invalid"))
			return
		}

		for sequence := moqt.GroupSequence(0); sequence < 30; sequence++ {
			stream, err := sess.OpenDataStream(substr, sequence, 0)
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
	go func() {
		annstr, err := sess.OpenAnnounceStream(moqt.Interest{TrackPrefix: echoTrackPrefix})
		if err != nil {
			slog.Error("failed to get an interest", slog.String("error", err.Error()))
			return
		}

		announcements, err := annstr.ReceiveAnnouncements()
		if err != nil {
			slog.Error("failed to get active tracks", slog.String("error", err.Error()))
			return
		}

		slog.Info("Active Tracks", slog.Any("announcements", announcements))

		subscription := moqt.Subscription{
			Track: moqt.Track{
				TrackPath:     echoTrackPath,
				TrackPriority: 0,
				GroupOrder:    0,
				GroupExpires:  1 * time.Second,
			},
		}
		substr, err := sess.OpenSubscribeStream(subscription)
		if err != nil {
			slog.Error("failed to subscribe", slog.String("error", err.Error()))
			return
		}

		for {
			stream, err := sess.AcceptDataStream(substr, context.Background())
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

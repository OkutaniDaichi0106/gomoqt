package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

var echoTrackPrefix = []string{"japan", "kyoto"}
var echoTrackPath = []string{"japan", "kyoto", "kiu", "text"}

func main() {
	/*
	 * Set Log Level to "DEBUG"
	 */
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	c := moqt.Client{
		TLSConfig:  &tls.Config{},
		QUICConfig: &quic.Config{},
	}

	// Get a setup request
	req := moqt.SetupRequest{
		URL: "https://localhost:8443/path",
	}

	// Dial to the server with the setup request
	slog.Info("Dial to the server")
	sess, _, err := c.Dial(req, context.Background())
	if err != nil {
		slog.Error(err.Error())
		return
	}

	wg := new(sync.WaitGroup)

	// Run a publisher
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Runing a publisher")

		slog.Info("Waiting an Announce Stream")
		// Accept an Announce Stream
		annstr, err := sess.AcceptAnnounceStream(context.Background(), func(ac moqt.AnnounceConfig) error {
			if !moqt.HasPrefix(echoTrackPath, ac.TrackPrefix) {
				return moqt.ErrTrackDoesNotExist
			}
			return nil
		})

		if err != nil {
			slog.Error("failed to accept an interest", slog.String("error", err.Error()))
			return
		}

		slog.Info("Accepted an Announce Stream")

		// Send Announcements
		announcements := []moqt.Announcement{
			{
				TrackPath: echoTrackPath,
			},
		}

		slog.Info("Send Announcements")

		// Send Announcements
		err = annstr.SendAnnouncement(announcements)
		if err != nil {
			slog.Error("failed to announce", slog.String("error", err.Error()))
			return
		}

		slog.Info("Announced")

		// Accept a subscription
		slog.Info("Waiting a subscribe stream")

		substr, err := sess.AcceptSubscribeStream(context.Background(), func(sc moqt.SubscribeConfig) (moqt.Info, error) {
			slog.Info("Received a subscribe request", slog.Any("config", sc))

			if !moqt.IsSamePath(sc.TrackPath, echoTrackPath) {
				return moqt.Info{}, moqt.ErrTrackDoesNotExist
			}

			return moqt.Info{
				TrackPriority:       0,
				LatestGroupSequence: 0,
				GroupOrder:          0,
			}, nil
		})
		if err != nil {
			slog.Error("failed to accept a subscribe stream", slog.String("error", err.Error()))
			return
		}

		if moqt.IsSamePath(substr.SubscribeConfig().TrackPath, echoTrackPath) {
			slog.Error("failed to get a track path", slog.String("error", "track path is invalid"))
			substr.CloseWithError(moqt.ErrTrackDoesNotExist)
			return
		}

		for sequence := moqt.GroupSequence(0); sequence < 30; sequence++ {
			stream, err := sess.OpenGroupStream(substr, sequence)
			if err != nil {
				slog.Error("failed to open a data stream", slog.String("error", err.Error()))
				return
			}

			err = stream.WriteFrame([]byte("HELLO!!"))
			if err != nil {
				slog.Error("failed to write data", slog.String("error", err.Error()))
				return
			}

			time.Sleep(3 * time.Second)
		}
	}()

	// Run a subscriber
	wg.Add(1)
	go func() {
		defer wg.Done()

		slog.Info("Run a subscriber")

		slog.Info("Receive Announcements")
		annstr, err := sess.OpenAnnounceStream(moqt.AnnounceConfig{TrackPrefix: echoTrackPrefix})
		if err != nil {
			slog.Error("failed to get an interest", slog.String("error", err.Error()))
			return
		}

		announcements, err := annstr.ReceiveAnnouncements()
		if err != nil {
			slog.Error("failed to get active tracks", slog.String("error", err.Error()))
			return
		}

		slog.Info("Announced", slog.Any("announcements", announcements))

		config := moqt.SubscribeConfig{
			TrackPath:     echoTrackPath,
			TrackPriority: 0,
			GroupOrder:    0,
		}

		slog.Info("Subscribing", slog.Any("config", config))

		substr, info, err := sess.OpenSubscribeStream(config)
		if err != nil {
			slog.Error("failed to subscribe", slog.String("error", err.Error()))
			return
		}

		slog.Info("Subscribed", slog.Any("info", info))

		for {
			stream, err := sess.AcceptGroupStream(context.Background(), substr)
			if err != nil {
				slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
				return
			}

			buf, err := stream.ReadFrame()
			if len(buf) > 0 {
				slog.Info("received data", slog.String("data", string(buf)))
			}

			if err != nil {
				slog.Error("failed to read data", slog.String("error", err.Error()))
				return
			}
		}
	}()

	wg.Wait()
}

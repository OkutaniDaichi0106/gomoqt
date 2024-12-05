package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"strconv"
	"time"

	moqt "github.com/OkutaniDaichi0106/gomoqt"
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
		URL:               "https://localhost:8443/path",
		SupportedVersions: []moqt.Version{moqt.Default},
		TLSConfig:         &tls.Config{},
		SessionHandler:    moqt.ClientSessionHandlerFunc(handleClientSession),
	}

	c.AddAnnouncement(moqt.Announcement{TrackPath: echoTrackPath})

	err := c.Run(context.Background())
	if err != nil {
		log.Print(err)
		return
	}

}

/*
 * Client Session Handler
 */
func handleClientSession(sess *moqt.ClientSession) {
	/*
	 * Publish data
	 */
	go func() {
		var sequence moqt.GroupSequence = 1
		for i := 0; i < 10; i++ {
			//
			time.Sleep(3 * time.Second)

			streams, err := sess.OpenDataStreams(echoTrackPath, sequence, 0, 1*time.Second)
			if err != nil {
				slog.Error("failed to open a data stream", slog.String("error", err.Error()))
				return
			}

			text := "hello!!!" + strconv.Itoa(i)

			for _, stream := range streams {
				_, err = stream.Write([]byte(text))
				if err != nil {
					slog.Error("failed to send data", slog.String("error", err.Error()))
					return
				}

				slog.Info("sent data", slog.String("text", text))

				stream.Close()
			}

			sequence++
		}
	}()

	/*
	 * Subscribe data
	 */
	// Interest
	annstr, err := sess.Interest(moqt.Interest{
		TrackPrefix: echoTrackPrefix,
	})
	if err != nil {
		slog.Error("failed to interest", slog.String("error", err.Error()))
		return
	}

	//  Get Announcements
	for {
		ann, err := annstr.Read()
		if err != nil {
			slog.Error("failed to read an announcement", slog.String("error", err.Error()))
			return
		}
		slog.Info("Received an announcement", slog.Any("announcement", ann))

		if ann.TrackPath == echoTrackPath {
			break
		}
	}

	// Subscribe
	subscription := moqt.Subscription{
		TrackPath: echoTrackPath,
	}
	_, info, err := sess.Subscribe(subscription)
	if err != nil {
		slog.Error("failed to subscribe", slog.String("error", err.Error()))
		return
	}
	slog.Info("Successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))

	/*
	 * Receive data
	 */
	buf := make([]byte, 1<<5)
	for {
		group, stream, err := sess.AcceptDataStream(context.Background())
		if err != nil {
			slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
			return
		}

		_, err = stream.Read(buf)
		if err != nil {
			slog.Error("failed to receive data", slog.String("error", err.Error()))
			return
		}

		slog.Info("received data", slog.Any("group", group), slog.String("text", string(buf)))
	}
}

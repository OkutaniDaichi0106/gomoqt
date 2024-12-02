package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	moqt "github.com/OkutaniDaichi0106/gomoqt"
)

func main() {
	/*
	 * Set Log Level to "DEBUG"
	 */
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	handler := sessionHandler{
		subscribedCh: make(chan moqt.Subscription, 1),
	}

	c := moqt.Client{
		URL:               "https://localhost:8443/path",
		SupportedVersions: []moqt.Version{moqt.Default},
		TLSConfig:         &tls.Config{},
		Announcements: []moqt.Announcement{
			moqt.Announcement{
				TrackNamespace: "japan/kyoto/student",
			},
		},
		SessionHandler: &handler,
	}

	err := c.Run(context.Background())
	if err != nil {
		log.Print(err)
		return
	}

}

/*
 * Client Session Handler
 */
var _ moqt.ClientSessionHandler = (*sessionHandler)(nil)

type sessionHandler struct {
	subscribedCh chan moqt.Subscription
}

func (h *sessionHandler) HandleClientSession(sess *moqt.ClientSession) {
	slog.Info("Subscribing")

	/*
	 * Send data
	 */
	go func() {
		var sequence moqt.GroupSequence = 1
		for i := 0; i < 10; i++ {
			//
			time.Sleep(33 * time.Millisecond)

			stream, err := sess.OpenDataStream(subscription, sequence, 0)
			if err != nil {
				slog.Error("failed to open a data stream", slog.String("error", err.Error()))
				return
			}

			text := "hello!!!" + strconv.Itoa(i)

			_, err = stream.Write([]byte(text))
			if err != nil {
				slog.Error("failed to send data", slog.String("error", err.Error()))
				return
			}

			slog.Info("sent data", slog.String("text", text))

			stream.Close()

			sequence++
		}
	}()

	/*
	 * Interest
	 */
	interest := moqt.Interest{
		TrackPrefix: h.localTrack.TrackNamespace,
	}

	annstr, err := sess.Interest(interest)
	if err != nil {
		slog.Error("failed to interest", slog.String("error", err.Error()))
		return
	}

	/*
	 * Get Announcements
	 */
	ann, err := annstr.ReadAnnouncement()
	if err != nil {
		slog.Error("failed to read an announcement", slog.String("error", err.Error()))
		return
	}

	log.Print("ECHO REACH")
	//
	slog.Info("received an announcement", slog.Any("announcement", ann))

	/*
	 * Subscribe
	 */
	_, info, err := sess.Subscribe(subscription)
	if err != nil {
		slog.Error("failed to subscribe", slog.String("error", err.Error()))
		return
	}
	//
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

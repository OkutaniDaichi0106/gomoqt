package main

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	handler := requestHandler{}
	c := moqt.Client{
		URL:                  "https://localhost:8443/path",
		SupportedVersions:    []moqt.Version{moqt.Devlop},
		TLSConfig:            &tls.Config{},
		RequestHandler:       handler,
		ClientSessionHandler: handler,
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
var _ moqt.ClientSessionHandler = (*requestHandler)(nil)

func (h requestHandler) HandleClientSession(sess *moqt.ClientSession) {
	subscription := <-h.subscribedCh

	slog.Info("Subscribed!!!")

	if subscription.TrackName != "text" {
		return
	}

	/*
	 * Send data
	 */
	go func() {
		sequence := 0
		for {
			stream, err := sess.OpenDataStream(subscription, sequence, 0)
			if err != nil {
				slog.Error("failed to open a data stream", slog.String("error", err.Error()))
				return
			}

			text := "hello!!!"

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
	//
	slog.Info("interest", slog.Any("interest", interest))

	/*
	 * Get Announcements
	 */
	ann, err := annstr.ReadAnnouncement()
	if err != nil {
		slog.Error("failed to read an announcement", slog.String("error", err.Error()))
		return
	}
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
	slog.Info("successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))

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

/*
 * Request Handler
 */
var _ moqt.RequestHandler = (*requestHandler)(nil)

type requestHandler struct {
	localTrack moqt.Announcement

	subscribedCh chan moqt.Subscription
}

func (requestHandler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
	// Reject all fetch request
	w.Reject(nil)
}

func (requestHandler) HandleInfoRequest(r moqt.InfoRequest, i *moqt.Info, w moqt.InfoWriter) {
	if i == nil {
		// Reject when information not found
		w.Reject(moqt.ErrNoGroup)
		return
	}

	// Answer the information
	w.Answer(*i)
}

func (h requestHandler) HandleInterest(i moqt.Interest, a []moqt.Announcement, w moqt.AnnounceWriter) {
	h.localTrack = moqt.Announcement{
		TrackNamespace: i.TrackPrefix + "room-0x000001/user-0x000001",
	}

	w.Announce(h.localTrack)

	// Close the Announce Stream after 30 minutes
	time.Sleep(30 * time.Minute)

	w.Close(nil)
}

func (h requestHandler) HandleSubscribe(s moqt.Subscription, i *moqt.Info, w moqt.SubscribeResponceWriter) {
	slog.Info("Subscribed", slog.Any("subscription", s))

	if h.localTrack.TrackNamespace != s.TrackNamespace {
		// Reject if get a subscription with an unknown Track Namespace
		w.Reject(nil)
	}

	if i == nil {
		// Reject the subscription if track was not found
		w.Reject(moqt.ErrTrackDoesNotExist)
		return
	}

	// Accept the subscription if a track was found
	w.Accept(*i)

	h.subscribedCh <- s
}

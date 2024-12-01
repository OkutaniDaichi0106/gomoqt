package main

import (
	"crypto/tls"
	"log"
	"log/slog"
	"os"
	"time"

	moqt "github.com/OkutaniDaichi0106/gomoqt"
	"github.com/quic-go/quic-go"
)

func main() {
	/*
	 * Set Log Level to "DEBUG"
	 */
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	/*
	 * Set certification config
	 */
	certs, err := getCertificates("localhost.pem", "localhost-key.pem")
	if err != nil {
		return
	}

	/*
	 * Initialize a Server
	 */
	moqs := moqt.Server{
		Addr: "localhost:8443",
		TLSConfig: &tls.Config{
			NextProtos:         []string{"h3", "moq-00"},
			Certificates:       certs,
			InsecureSkipVerify: true, // TODO:
		},
		QUICConfig: &quic.Config{
			Allow0RTT:       true,
			EnableDatagrams: true,
		},
	}

	// Initialize a Relayer
	relayer := moqt.Relayer{
		Path:           "/path",
		RequestHandler: requestHandler{},
		SessionHandler: serverSessionHandler{},
		RelayManager:   nil,
	}

	// Run the Relayer on WebTransport
	moqs.RunOnWebTransport(relayer)

	moqs.ListenAndServe()
}

var _ moqt.ServerSessionHandler = (*serverSessionHandler)(nil)

type serverSessionHandler struct{}

func (serverSessionHandler) HandleServerSession(sess *moqt.ServerSession) {
	/*
	 * Interest
	 */
	interest := moqt.Interest{
		TrackPrefix: "relayer",
	}
	slog.Info("interest", slog.Any("interest", interest))
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

	slog.Info("received an announcement", slog.Any("announcement", ann))

	/*
	 * Subscribe
	 */
	subscription := moqt.Subscription{
		TrackNamespace: ann.TrackNamespace,
		TrackName:      "text",
	}

	_, info, err := sess.Subscribe(subscription)
	if err != nil {
		slog.Error("failed to subscribe", slog.String("error", err.Error()))
		return
	}

	slog.Info("successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))
}

var _ moqt.RequestHandler = (*requestHandler)(nil)

type requestHandler struct{}

func (requestHandler) HandleInterest(i moqt.Interest, a []moqt.Announcement, w moqt.AnnounceWriter) {
	if a == nil {
		// Close the Announce Stream if track was not found
		w.Close(moqt.ErrTrackDoesNotExist)
	}

	log.Print("Announcements", a, len(a))

	for _, announcement := range a {
		w.Announce(announcement)
	}

	// Close the Announce Stream after 30 minutes
	time.Sleep(30 * time.Minute)

	w.Close(nil)
}

func (requestHandler) HandleSubscribe(s moqt.Subscription, info *moqt.Info, w moqt.SubscribeResponceWriter) {
	if info == nil {
		/*
		 * When info is nil, it means the subscribed track was not found.
		 * Reject the subscription or make a new subscrition to upstream.
		 */
		w.Reject(moqt.ErrTrackDoesNotExist)
		return
	}

	w.Accept(*info)
}

func (requestHandler) HandleFetch(r moqt.FetchRequest, w moqt.FetchResponceWriter) {
	// Reject all fetch request
	w.Reject(moqt.ErrNoGroup)
}

func (requestHandler) HandleInfoRequest(r moqt.InfoRequest, i *moqt.Info, w moqt.InfoWriter) {
	if i == nil {
		/*
		 * When info is nil, it means the subscribed track was not found.
		 * Reject the request or make a new subscrition to upstream and request information of the track.
		 */
		w.Reject(moqt.ErrTrackDoesNotExist)
		return
	}

	w.Answer(*i)
}

func getCertificates(certFile, keyFile string) ([]tls.Certificate, error) {
	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return certs, nil

}

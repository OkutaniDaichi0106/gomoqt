package main

import (
	"crypto/tls"
	"log/slog"
	"os"

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
		SessionHandler: moqt.ServerSessionHandlerFunc(handleServerSession),
		RelayManager:   nil,
	}

	// Run the Relayer on WebTransport
	moqs.RunOnWebTransport(relayer)

	moqs.ListenAndServe()
}

func handleServerSession(sess *moqt.ServerSession) {
	echoTrackPrefix := "japan/kyoto"
	echoTrackPath := "japan/kyoto/kiu/text"

	/*
	 * Interest
	 */
	interest := moqt.Interest{
		TrackPrefix: echoTrackPrefix,
	}
	annstr, err := sess.Interest(interest)
	if err != nil {
		slog.Error("failed to interest", slog.String("error", err.Error()))
		return
	}

	/*
	 * Get Announcements
	 */
	for {
		ann, err := annstr.ReadAnnouncement()
		if err != nil {
			slog.Error("failed to read an announcement", slog.String("error", err.Error()))
			return
		}
		slog.Info("received an announcement", slog.Any("announcement", ann))

		if ann.TrackPath == echoTrackPath {
			break
		}
	}

	/*
	 * Subscribe
	 */
	subscription := moqt.Subscription{
		TrackPath: echoTrackPath,
	}

	_, info, err := sess.Subscribe(subscription)
	if err != nil {
		slog.Error("failed to subscribe", slog.String("error", err.Error()))
		return
	}

	slog.Info("successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))
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

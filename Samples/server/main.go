package main

import (
	"crypto/tls"
	"log/slog"

	moqt "github.com/OkutaniDaichi0106/gomoqt"
	"github.com/quic-go/quic-go"
)

func main() {
	/*
	 * Set Log Level to "DEBUG"
	 */
	// logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	// slog.SetDefault(logger)

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
	moqServer := moqt.Server{
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

	/*
	 * Set a handler function
	 */
	moqt.HandleFunc("/path", func(ss moqt.ServerSession) {
		echoTrackPrefix := "japan/kyoto"
		echoTrackPath := "japan/kyoto/kiu/text"

		subscriber := ss.Subscriber()
		/*
		 * Interest
		 */
		interest := moqt.Interest{
			TrackPrefix: echoTrackPrefix,
		}
		annstr, err := subscriber.Interest(interest)
		if err != nil {
			slog.Error("failed to interest", slog.String("error", err.Error()))
			return
		}

		/*
		 * Get Announcements
		 */
		for {
			ann, err := annstr.Read()
			if err != nil {
				slog.Error("failed to read an announcement", slog.String("error", err.Error()))
				return
			}
			slog.Info("Received an announcement", slog.Any("announcement", ann))

			/*
			 * Subscribe
			 */
			subscription := moqt.Subscription{
				TrackPath: echoTrackPath,
			}

			_, info, err := subscriber.Subscribe(subscription)
			if err != nil {
				slog.Error("failed to subscribe", slog.String("error", err.Error()))
				return
			}

			slog.Info("successfully subscribed", slog.Any("subscription", subscription), slog.Any("info", info))
		}
	})

	moqServer.ListenAndServe()
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

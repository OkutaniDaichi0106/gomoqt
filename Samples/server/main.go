package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"time"

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

		interest, err := subscriber.Interest(moqt.Interest{
			TrackPrefix: echoTrackPrefix,
		})
		if err != nil {
			slog.Error("failed to interest", slog.String("error", err.Error()))
			return
		}

		tracks, err := interest.NextActiveTracks()
		if err != nil {
			slog.Error("failed to get active tracks", slog.String("error", err.Error()))
			return
		}

		_, ok := tracks.Get(echoTrackPath)
		if !ok {
			slog.Error("failed to get the active track", slog.String("error", "track is not active"))
			return
		}

		subscription, err := subscriber.Subscribe(moqt.Subscription{
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

		// Receive data

		for {
			stream, err := subscription.AcceptDataStream(context.Background())
			if err != nil {
				slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
				return
			}

			go func(stream moqt.ReceiveDataStream) {
				for {
					buf := make([]byte, 1024)
					n, err := stream.Read(buf)
					if err != nil {
						slog.Error("failed to read data", slog.String("error", err.Error()))
						return
					}
					slog.Info("Received", slog.String("data", string(buf[:n])))
				}
			}(stream)
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

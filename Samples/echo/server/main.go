package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"os"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransfork"
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
	moqServer := moqtransfork.Server{
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
	slog.Info("Server runs on path: \"/path\"")
	moqtransfork.HandleFunc("/path", func(sess moqtransfork.ServerSession) {
		echoTrackPrefix := "japan/kyoto"
		echoTrackPath := "japan/kyoto/kiu/text"

		dataCh := make(chan []byte, 1<<3)

		/*
		 * Subscriber
		 */
		go func() {
			/*
			 * Request Announcements
			 */
			slog.Info("Request Announcements")
			interest := moqtransfork.Interest{
				TrackPrefix: echoTrackPrefix,
			}
			annstr, err := sess.OpenAnnounceStream(interest)
			if err != nil {
				slog.Error("failed to interest", slog.String("error", err.Error()))
				return
			}

			/*
			 * Receive Announcements
			 */
			slog.Info("Receive Announcements")
			announcements, err := annstr.ReceiveAnnouncements()
			if err != nil {
				slog.Error("failed to get active tracks", slog.String("error", err.Error()))
				return
			}

			slog.Info("Announcements", slog.Any("announcements", announcements))

			/*
			 * Subscribe
			 */
			slog.Info("Subscribe")
			subscription := moqtransfork.Subscription{
				TrackPath:     echoTrackPath,
				TrackPriority: 0,
				GroupOrder:    0,
				GroupExpires:  1 * time.Second,
			}
			substr, err := sess.OpenSubscribeStream(subscription)
			if err != nil {
				slog.Error("failed to subscribe", slog.String("error", err.Error()))
				return
			}

			/*
			 * Receive data
			 */
			slog.Info("Receive data")
			for {
				stream, err := sess.AcceptDataStream(substr, context.Background())
				if err != nil {
					slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
					return
				}

				go func(stream moqtransfork.ReceiveDataStream) {
					for {
						buf := make([]byte, 1024)
						n, err := stream.Read(buf)
						if err != nil {
							slog.Error("failed to read data", slog.String("error", err.Error()))
							return
						}
						slog.Info("Received a frame", slog.String("frame", string(buf[:n])))

						dataCh <- buf[:n]
					}
				}(stream)
			}
		}()

		/*
		 * Publisher
		 */
		go func() {
			/*
			 * Announce
			 */
			slog.Info("Waiting an Announce Stream")
			annstr, err := sess.AcceptAnnounceStream(context.Background())
			if err != nil {
				slog.Error("failed to accept an announce stream", slog.String("error", err.Error()))
				return
			}
			slog.Info("Accepted an Announce Stream")

			slog.Info("Announce")
			announcements := []moqtransfork.Announcement{
				{
					TrackPath: echoTrackPath,
				},
				{
					TrackPath: "japan/kyoto/kiu/audio", //
				},
			}
			err = annstr.SendAnnouncement(announcements)
			if err != nil {
				slog.Error("failed to send an announcement", slog.String("error", err.Error()))
				return
			}

			/*
			 * Accept subscription
			 */
			slog.Info("Accept subscription")
			substr, err := sess.AcceptSubscribeStream(context.Background())
			if err != nil {
				slog.Error("failed to accept a subscription", slog.String("error", err.Error()))
				return
			}

			if substr.Subscription().TrackPath != echoTrackPath {
				slog.Error("failed to get a track path", slog.String("error", "track path is invalid"))
				return
			}

			/*
			 * Send data
			 */
			for sequence := moqtransfork.GroupSequence(1); sequence < 30; sequence++ {
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
	})

	slog.Info("Start a server")
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

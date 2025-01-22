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
	slog.Info("Server runs on path: \"/path\"")
	moqt.HandleFunc("/path", func(sess moqt.Session) {
		echoTrackPrefix := []string{"japan", "kyoto"}
		echoTrackPath := []string{"japan", "kyoto", "kiu", "text"}

		dataCh := make(chan []byte, 1<<3)

		wg := new(sync.WaitGroup)
		/*
		 * Subscriber
		 */
		wg.Add(1)
		go func() {
			defer wg.Done()
			/*
			 * Request Announcements
			 */
			slog.Info("Request Announcements")
			interest := moqt.AnnounceConfig{
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
			subcfg := moqt.SubscribeConfig{
				TrackPath:     echoTrackPath,
				TrackPriority: 0,
				GroupOrder:    0,
			}
			substr, info, err := sess.OpenSubscribeStream(subcfg)
			if err != nil {
				slog.Error("failed to subscribe", slog.String("error", err.Error()))
				return
			}

			slog.Info("Subscribed", slog.Any("info", info))

			/*
			 * Receive data
			 */
			slog.Info("Receive data")
			for {
				stream, err := sess.AcceptGroupStream(context.Background(), substr)
				if err != nil {
					slog.Error("failed to accept a data stream", slog.String("error", err.Error()))
					return
				}

				go func(stream moqt.ReceiveGroupStream) {
					for {
						buf, err := stream.ReadFrame()
						if err != nil {
							slog.Error("failed to read data", slog.String("error", err.Error()))
							return
						}
						slog.Info("Received a frame", slog.String("frame", string(buf)))

						dataCh <- buf
					}
				}(stream)
			}
		}()

		/*
		 * Publisher
		 */
		wg.Add(1)
		go func() {
			defer wg.Done()
			/*
			 * Announce
			 */
			slog.Info("Waiting an Announce Stream")

			annstr, err := sess.AcceptAnnounceStream(context.Background(), func(ac moqt.AnnounceConfig) error {
				slog.Info("Received an announce request", slog.Any("config", ac))

				if !moqt.HasPrefix(echoTrackPath, ac.TrackPrefix) {
					return moqt.ErrTrackDoesNotExist
				}

				return nil
			})

			if err != nil {
				slog.Error("failed to accept an announce stream", slog.String("error", err.Error()))
				return
			}

			slog.Info("Accepted an Announce Stream")

			slog.Info("Announcing")

			announcements := []moqt.Announcement{
				{
					TrackPath: echoTrackPath,
				},
			}
			err = annstr.SendAnnouncement(announcements)
			if err != nil {
				slog.Error("failed to send an announcement", slog.String("error", err.Error()))
				return
			}

			slog.Info("Announced")

			/*
			 * Accept a subscription
			 */
			slog.Info("Waiting a subscribe stream")

			substr, err := sess.AcceptSubscribeStream(context.Background(), func(sc moqt.SubscribeConfig) (moqt.Info, error) {
				slog.Info("Received a subscribe request", slog.Any("config", sc))

				if !moqt.IsSamePath(sc.TrackPath, echoTrackPath) {
					return moqt.Info{}, moqt.ErrTrackDoesNotExist
				}

				return moqt.Info{}, nil
			})
			if err != nil {
				slog.Error("failed to accept a subscription", slog.String("error", err.Error()))
				return
			}

			slog.Info("Accepted a subscribe stream")

			/*
			 * Send data
			 */
			for sequence := moqt.GroupSequence(1); sequence < 30; sequence++ {
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

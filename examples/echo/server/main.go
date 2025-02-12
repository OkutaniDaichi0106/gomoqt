package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/quic-go/quic-go"
)

const (
	lightPink   = "\033[38;5;218m"
	lightOrange = "\033[38;5;208m"
	reset       = "\033[0m"
)

const (
	CertFile = "./examples/cert/localhost.pem"
	KeyFile  = "./examples/cert/localhost-key.pem"
)

type colorTextHandler struct {
	out   io.Writer
	opts  *slog.HandlerOptions
	color string
}

func newColorTextHandler(out io.Writer, opts *slog.HandlerOptions, color string) *colorTextHandler {
	return &colorTextHandler{
		out:   out,
		opts:  opts,
		color: color,
	}
}

func (h *colorTextHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	timeStr := r.Time.Format(time.RFC3339)

	// Format attributes
	attrs := ""
	r.Attrs(func(a slog.Attr) bool {
		if stringer, ok := a.Value.Any().(fmt.Stringer); ok {
			attrs += fmt.Sprintf("%s=%s ", a.Key, stringer.String())
		} else {
			attrs += fmt.Sprintf("%s=%v ", a.Key, a.Value)
		}
		return true
	})

	// Color the entire line
	fmt.Fprintf(h.out, "%stime=%s level=%s\n    msg=%q\n    %s %s\n", h.color, timeStr, level, r.Message, attrs, reset)

	return nil
}

func (h *colorTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *colorTextHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *colorTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func main() {
	/*
	 * Set Log Level to "DEBUG"
	 */
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	/*
	 * Set certification config
	 */
	certs, err := getCertificates(CertFile, KeyFile)
	if err != nil {
		slog.Error("failed to get certificates", slog.String("error", err.Error()))
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
	slog.Info("Server runs on path: \"/echo\"")
	moqt.HandleFunc("/echo", func(sess moqt.Session) {
		echoTrackPrefix := []string{"japan", "kyoto"}
		echoTrackPath := moqt.NewTrackPath("japan", "kyoto", "text")

		groupBufferCh := make(chan *moqt.GroupBuffer, 1<<2)

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
			subLogger := slog.New(newColorTextHandler(os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
				lightOrange))

			subLogger.Info("Running a subscriber")

			subLogger.Info("Opening an Announce Stream")

			annstr, err := sess.OpenAnnounceStream(moqt.AnnounceConfig{
				TrackPrefix: echoTrackPrefix,
			})
			if err != nil {
				subLogger.Error("failed to open an announce stream", slog.String("error", err.Error()))
				return
			}

			subLogger.Info("Opened an Announce Stream")

			subLogger.Info("Receiving announcements")

			announcements, err := annstr.ReceiveAnnouncements()
			if err != nil {
				subLogger.Error("failed to receive announcements", slog.String("error", err.Error()))
				return
			}

			subLogger.Info("Received announcements", slog.Any("announcements", announcements))

			/*
			 * Subscribe
			 */
			subLogger.Info("Subscribing")

			substr, info, err := sess.OpenSubscribeStream(moqt.SubscribeConfig{
				TrackPath:     echoTrackPath,
				TrackPriority: 0,
				GroupOrder:    0,
			})
			if err != nil {
				subLogger.Error("failed to subscribe", slog.String("error", err.Error()))
				return
			}

			subLogger.Info("Subscribed", slog.Any("info", info))

			/*
			 * Receive data
			 */
			subLogger.Info("Receiving data")

			for {
				stream, err := sess.AcceptGroupStream(context.Background(), substr)
				if err != nil {
					subLogger.Error("failed to accept a data stream", slog.String("error", err.Error()))
					return
				}

				gb := moqt.NewGroupBuffer(stream.GroupSequence(), 1<<2)

				groupBufferCh <- gb

				moqt.RelayGroup(stream, gb)
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
			pubLogger := slog.New(newColorTextHandler(os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
				lightPink))

			pubLogger.Info("Waiting for an Announce Stream")

			annstr, err := sess.AcceptAnnounceStream(context.Background(), func(ac moqt.AnnounceConfig) error {
				pubLogger.Info("Received an announce request", slog.Any("config", ac))

				if !echoTrackPath.HasPrefix(ac.TrackPrefix) {
					return moqt.ErrTrackDoesNotExist
				}

				return nil
			})

			if err != nil {
				pubLogger.Error("Failed to accept an announce stream", slog.String("error", err.Error()))
				return
			}

			pubLogger.Info("Accepted an Announce Stream")

			pubLogger.Info("Announcing")

			err = annstr.SendAnnouncement([]moqt.Announcement{
				{
					AnnounceStatus: moqt.ACTIVE,
					TrackPath:      echoTrackPath,
				},
			})
			if err != nil {
				pubLogger.Error("Failed to send an announcement", slog.String("error", err.Error()))
				return
			}

			pubLogger.Info("Successfully Announced")

			/*
			 * Accept a subscription
			 */
			pubLogger.Info("Waiting for a subscribe stream")

			substr, err := sess.AcceptSubscribeStream(context.Background(), func(sc moqt.SubscribeConfig) (moqt.Info, error) {
				pubLogger.Info("Received a subscribe request", slog.Any("config", sc))

				if !sc.TrackPath.Equal(echoTrackPath) {
					return moqt.Info{}, moqt.ErrTrackDoesNotExist
				}

				return moqt.Info{}, nil
			})
			if err != nil {
				pubLogger.Error("Failed to accept a subscription", slog.String("error", err.Error()))
				return
			}

			pubLogger.Info("Accepted a subscribe stream")

			/*
			 * Send data
			 */
			for groupBuffer := range groupBufferCh {
				stream, err := sess.OpenGroupStream(substr, groupBuffer.GroupSequence())
				if err != nil {
					pubLogger.Error("Failed to open a data stream", slog.String("error", err.Error()))

					return
				}

				err = moqt.RelayGroup(groupBuffer, stream)
				if err != nil {
					pubLogger.Error("Failed to write data", slog.String("error", err.Error()))
					return
				}

				groupBuffer.Close()
			}
		}()
	})

	slog.Info("Starting the server")
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

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

var echoTrackPrefix = []string{"japan", "kyoto"}
var echoTrackPath = moqt.NewTrackPath("japan", "kyoto", "text")

func main() {
	/*
	 * Set Log Level to "INFO"
	 */
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	c := moqt.Client{
		TLSConfig:  &tls.Config{},
		QUICConfig: &quic.Config{},
	}

	// Dial to the server with the setup request
	slog.Info("Dial to the server")

	sess, _, err := c.Dial("https://localhost:8443/echo", context.Background())
	if err != nil {
		slog.Error(err.Error())
		return
	}

	wg := new(sync.WaitGroup)

	// Define frames for animation
	frames := []string{
		fmt.Sprintf(`
    %s    ☆         o
         ☆    ☆
          ☆       ☆    
    %s    ☆       o   ☆
         ☆       ☆
      o         %s☆   %s
`, "\033[31m", "\033[32m", "\033[33m", "\033[0m"),
		fmt.Sprintf(`
    %s    ☆            ☆
         ☆   o
       o     ☆
    %s ☆       ☆   ☆
       ☆   o        ☆
    %s       ☆     ☆   %s
`, "\033[36m", "\033[34m", "\033[33m", "\033[0m"),
		fmt.Sprintf(`
    %s  o          ☆   
       ☆       ☆    
         ☆     ☆
    %s☆         ☆      o
       ☆  o         ☆ 
    %s     ☆         %s
`, "\033[33m", "\033[31m", "\033[32m", "\033[0m"),
		fmt.Sprintf(`
    %s    ☆         ☆
         ☆   ☆
      o       ☆     
    %s   ☆       ☆    ☆
         ☆      o
     ☆           ☆  %s
`, "\033[34m", "\033[32m", "\033[0m"),
		fmt.Sprintf(`
    %s    o        ☆
      ☆       ☆   
    ☆     ☆       ☆
    %s  ☆      o       ☆
        o     ☆    ☆
      ☆         %s☆   %s
`, "\033[33m", "\033[36m", "\033[32m", "\033[0m"),
	}

	// Run a publisher
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Create light pink pubLogger with custom handler
		pubLogger := slog.New(newColorTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo},
			lightPink))

		pubLogger.Info("Running a publisher")

		pubLogger.Info("Waiting an Announce Stream")
		// Accept an Announce Stream
		annstr, err := sess.AcceptAnnounceStream(context.Background(), func(ac moqt.AnnounceConfig) error {

			pubLogger.Info("Received an announce request", slog.Any("config", ac))

			if !echoTrackPath.HasPrefix(ac.TrackPrefix) {
				return moqt.ErrTrackDoesNotExist
			}

			return nil
		})
		if err != nil {
			pubLogger.Error("failed to accept an interest", slog.String("error", err.Error()))
			return
		}

		pubLogger.Info("Accepted an Announce Stream")

		// Send Announcements

		pubLogger.Info("Send Announcements")

		// Send Announcements
		err = annstr.SendAnnouncement([]moqt.Announcement{
			{
				AnnounceStatus: moqt.ACTIVE,
				TrackPath:      echoTrackPath,
			},
		})
		if err != nil {
			pubLogger.Error("failed to announce", slog.String("error", err.Error()))
			return
		}

		pubLogger.Info("Announced")

		// Accept a subscription
		pubLogger.Info("Waiting a subscribe stream")

		substr, err := sess.AcceptSubscribeStream(context.Background(), func(sc moqt.SubscribeConfig) (moqt.Info, error) {
			pubLogger.Info("Received a subscribe request", slog.Any("config", sc))

			if !sc.TrackPath.Equal(echoTrackPath) {
				return moqt.Info{}, moqt.ErrTrackDoesNotExist
			}

			return moqt.Info{
				TrackPriority:       0,
				LatestGroupSequence: 0,
				GroupOrder:          0,
			}, nil
		})
		if err != nil {
			pubLogger.Error("failed to accept a subscribe stream", slog.String("error", err.Error()))
			return
		}

		if !substr.SubscribeConfig().TrackPath.Equal(echoTrackPath) {
			pubLogger.Error("failed to get a track path", slog.String("error", "track path is invalid"))
			substr.CloseWithError(moqt.ErrTrackDoesNotExist)
			return
		}

		for seq := moqt.FirstSequence; seq < 300; seq++ {
			for _, frame := range frames {
				stream, err := sess.OpenGroupStream(substr, seq)
				if err != nil {
					pubLogger.Error("failed to open a data stream", slog.String("error", err.Error()))
					return
				}

				err = stream.WriteFrame([]byte(frame))
				if err != nil {
					pubLogger.Error("failed to write data", slog.String("error", err.Error()))
					return
				}

				stream.Close()

				time.Sleep(250 * time.Millisecond)

				seq = seq.Next()
			}
		}
	}()

	// Run a subscriber
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Create light orange subLogger with custom handler
		subLogger := slog.New(newColorTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo},
			lightOrange))

		subLogger.Info("Running a subscriber")

		subLogger.Info("Opening an Announce Stream")

		annstr, err := sess.OpenAnnounceStream(moqt.AnnounceConfig{TrackPrefix: echoTrackPrefix})
		if err != nil {
			subLogger.Error("failed to open an announce stream", slog.String("error", err.Error()))
			return
		}

		subLogger.Info("Opened an Announce Stream")

		subLogger.Info("Receiving announcements")

		announcements, err := annstr.ReceiveAnnouncements()
		if err != nil {
			subLogger.Error("failed to get announcements", slog.String("error", err.Error()))
			return
		}

		subLogger.Info("Received announcements", slog.Any("announcements", announcements))

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

		for {
			stream, err := sess.AcceptGroupStream(context.Background(), substr)
			if err != nil {
				subLogger.Error("failed to accept a data stream", slog.String("error", err.Error()))
				return
			}

			buf, err := stream.ReadFrame()
			if len(buf) > 0 {
				subLogger.Info("Received data", slog.String("data", string(buf)))
			}

			if err != nil {
				subLogger.Error("failed to read data", slog.String("error", err.Error()))
				return
			}
		}
	}()

	wg.Wait()
}

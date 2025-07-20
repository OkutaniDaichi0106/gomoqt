package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	path := flag.String("path", "", "operation path (e.g., /subscribe, /publish)")
	flag.Parse()

	client := &moqt.Client{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	switch *path {
	case "/subscribe":
		if err := subscribe(client); err != nil {
			slog.Error("subscribe failed", "error", err)
		}
	case "/publish":
		if err := publish(client); err != nil {
			slog.Error("publish failed", "error", err)
		}
	default:
		slog.Error("invalid path specified", "path", *path)
		flag.Usage()
	}
}

func subscribe(client *moqt.Client) error {
	sess, err := client.Dial(context.Background(), "https://moqt.example.com:9000/subscribe", nil)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return err
	}

	//
	anns, err := sess.OpenAnnounceStream("/")
	if err != nil {
		slog.Error("failed to open announce stream", "error", err)
		return err
	}
	defer anns.Close()

	for {
		ann, err := anns.ReceiveAnnouncement(context.Background())
		if err != nil {
			slog.Error("failed to receive announcement", "error", err)
			break
		}

		slog.Info("received announcement", "announcement", ann)

		go func(ann *moqt.Announcement) {
			if !ann.IsActive() {
				return
			}

			tr, err := sess.OpenTrackStream(ann.BroadcastPath(), "index", nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}
			defer tr.Close()

			for {
				gr, err := tr.AcceptGroup(context.Background())
				if err != nil {
					slog.Error("failed to accept group", "error", err)
					return
				}

				go func(gr *moqt.GroupReader) {
					defer gr.CancelRead(moqt.InternalGroupErrorCode)

					for {
						frame, err := gr.ReadFrame()
						if err != nil {
							if err == io.EOF {
								return
							}
							slog.Error("failed to read frame", "error", err)
							break
						}

						slog.Info("received a frame", "frame", string(frame.CopyBytes()))

						// TODO: Release the frame after processing
						// This is important to avoid memory leaks
						frame.Release()
					}
				}(gr)

			}

		}(ann)
	}

	return nil
}

func publish(client *moqt.Client) error {
	mux := moqt.NewTrackMux()
	mux.HandleFunc(context.Background(), "/interop.client", func(tw *moqt.TrackWriter) {
		seq := moqt.GroupSequenceFirst
		for {
			group, err := tw.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			frame := moqt.NewFrame([]byte("Hello from client"))
			err = group.WriteFrame(frame)
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				return
			}

			frame.Release()
			group.Close()

			seq = seq.Next()
		}
	})
	_, err := client.Dial(context.Background(), "https://moqt.example.com:9000/publish", mux)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return err
	}
	return nil
}

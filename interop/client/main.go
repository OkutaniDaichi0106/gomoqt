package main

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	moqt.HandleFunc(context.Background(), "/interop.client", func(tw *moqt.TrackWriter) {
		seq := moqt.GroupSequenceFirst
		for range 10 {
			group, err := tw.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			slog.Info("Opened group successfully", "group_sequence", group.GroupSequence())

			frame := moqt.NewFrame([]byte("Hello from client"))
			err = group.WriteFrame(frame)
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				return
			}

			slog.Info("Sent frame successfully", "frame", string(frame.Bytes()))

			group.Close()

			slog.Info("Closed group successfully", "group_sequence", seq)

			seq = seq.Next()

			time.Sleep(100 * time.Millisecond)
		}
	})

	client := &moqt.Client{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	sess, err := client.Dial(context.Background(), "https://moqt.example.com:9000/interop", nil)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}

	slog.Info("Connected to the server successfully")

	//
	anns, err := sess.OpenAnnounceStream("/")
	if err != nil {
		slog.Error("failed to open announce stream", "error", err)
		return
	}
	defer anns.Close()

	slog.Info("Opened announce stream successfully")

	ann, err := anns.ReceiveAnnouncement(context.Background())
	if err != nil {
		slog.Error("failed to receive announcement", "error", err)
		return
	}

	slog.Info("Received announcement", "announcement", ann)

	if !ann.IsActive() {
		slog.Info("Announcement is not active", "announcement", ann)
		return
	}

	tr, err := sess.OpenTrackStream(ann.BroadcastPath(), "", nil)
	if err != nil {
		slog.Error("failed to open track stream", "error", err)
		return
	}

	slog.Info("Opened track stream successfully", "path", ann.BroadcastPath())

	for {
		gr, err := tr.AcceptGroup(context.Background())
		if err != nil {
			slog.Error("failed to accept group", "error", err)
			break
		}

		slog.Info("Accepted a group", "group_sequence", gr.GroupSequence())

		go func(gr *moqt.GroupReader) {
			frame := moqt.NewFrame(nil)
			for {
				err := gr.ReadFrame(frame)
				if err != nil {
					if err == io.EOF {
						return
					}
					slog.Error("failed to read frame", "error", err)
					break
				}

				slog.Info("Received a frame", "frame", string(frame.Bytes()))
			}
		}(gr)
	}

	sess.Terminate(moqt.NoError, moqt.NoError.String())
}

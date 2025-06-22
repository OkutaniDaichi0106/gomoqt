package main

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	moqt.HandleFunc(context.Background(), "/client.echo", func(pub *moqt.Publisher) {
		seq := moqt.GroupSequenceFirst
		for {
			time.Sleep(100 * time.Millisecond)

			gw, err := pub.TrackWriter.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			err = gw.WriteFrame(moqt.NewFrame([]byte("FRAME " + seq.String())))
			if err != nil {
				gw.CancelWrite(moqt.InternalGroupErrorCode)
				slog.Error("failed to write frame", "error", err)
				return
			}

			gw.Close()

			seq = seq.Next()
		}
	})

	client := moqt.Client{
		Logger: slog.Default(),
	}

	sess, err := client.Dial(context.Background(), "https://localhost:4444/echo", nil)
	if err != nil {
		slog.Error("failed to dial",
			"error", err,
		)
		return
	}

	annstr, err := sess.OpenAnnounceStream("/")
	if err != nil {
		slog.Error("failed to open announce stream",
			"error", err,
		)
		return
	}

	for {
		ann, err := annstr.ReceiveAnnouncement(context.Background())
		if err != nil {
			slog.Error("failed to receive announcements",
				"error", err,
			)
			return
		}

		go func(ann *moqt.Announcement) {
			if !ann.IsActive() {
				return
			}

			sub, err := sess.OpenTrackStream(ann.BroadcastPath(), "index", nil)
			if err != nil {
				slog.Error("failed to open track stream", "error", err)
				return
			}

			for {
				gr, err := sub.TrackReader.AcceptGroup(context.Background())
				if err != nil {
					slog.Error("failed to accept group", "error", err)
					return
				}

				go func(gr moqt.GroupReader) {
					defer gr.CancelRead(moqt.InternalGroupErrorCode)

					for {
						frame, err := gr.ReadFrame()
						if err != nil {
							if err == io.EOF {
								return
							}
							slog.Error("failed to read frame", "error", err)
							return
						}

						slog.Info("received a frame", "frame", string(frame.CopyBytes()))

						// TODO: Release the frame after reading
						// This is important to avoid memory leaks
						frame.Release()
					}
				}(gr)

			}
		}(ann)
	}
}

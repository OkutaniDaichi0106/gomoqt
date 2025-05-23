package main

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func main() {
	mux := moqt.NewTrackMux()

	audio := mux.BuildTrack(context.TODO(), "/audio", moqt.Info{}, 0)
	video := mux.BuildTrack(context.TODO(), "/video", moqt.Info{}, 0)

	go func() {
		seq := moqt.FirstGroupSequence
		for {
			gw, err := audio.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			err = gw.WriteFrame(moqt.NewFrame([]byte("AUDIO_FRAME" + seq.String())))
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				return
			}

			gw.Close()

			seq = seq.Next()
		}
	}()

	go func() {
		seq := moqt.FirstGroupSequence
		for {
			gw, err := video.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			err = gw.WriteFrame(moqt.NewFrame([]byte("VIDEO_FRAME" + seq.String())))
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				return
			}

			gw.Close()

			seq = seq.Next()
		}
	}()

	client := moqt.Client{}

	sess, _, err := client.Dial(context.Background(), "https://localhost:4444/broadcast", mux)
	if err != nil {
		slog.Error("failed to dial", "error", err)
		return
	}

	annstr, err := sess.OpenAnnounceStream(&moqt.AnnounceConfig{TrackPattern: "/**"})
	if err != nil {
		slog.Error("failed to open announce stream", "error", err)
		return
	}
	for {
		announcements, err := annstr.ReceiveAnnouncements(context.TODO())
		if err != nil {
			slog.Error("failed to receive announcements", "error", err)
			return
		}
		for _, ann := range announcements {
			go func(ann *moqt.Announcement) {
				_, tr, err := sess.OpenTrackStream(ann.TrackPath(), nil)
				if err != nil {
					slog.Error("failed to open track stream", "error", err)
					return
				}

				for {
					gr, err := tr.AcceptGroup(context.Background())
					if err != nil {
						slog.Error("failed to accept group", "error", err)
						return
					}

					for {
						f, err := gr.ReadFrame()
						if err != nil {
							slog.Error("failed to read frame", "error", err)
							return
						}

						slog.Info("received frame", "frame", string(f.CopyBytes()))
					}
				}
			}(ann)

		}
	}

}

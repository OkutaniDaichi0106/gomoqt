package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/magefile/mage/mg"
)

type Client mg.Namespace

func (c *Client) Dial(ctx context.Context) error {
	flag.Parse()

	client := &moqt.Client{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	sess, err := client.Dial(ctx, "https://moqt.example.com:9000/interop", nil)
	if err != nil {
		return err
	}

	slog.Info("Connected to the server successfully")

	// Close the session when done
	return nil
}

func (c *Client) Discover(ctx context.Context) error {
	err := c.Dial(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Subscribe(ctx context.Context) error {
	err := c.Dial(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PublishAndAnnounce(ctx context.Context) error {
	err := c.Dial(ctx)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	moqt.PublishFunc(context.Background(), "/interop.client", func(tw *moqt.TrackWriter) {
		seq := moqt.GroupSequenceFirst
		frame := moqt.NewFrame(1024)
		for range 10 {
			group, err := tw.OpenGroup(seq)
			if err != nil {
				slog.Error("failed to open group", "error", err)
				return
			}

			slog.Info("Opened group successfully", "group_sequence", group.GroupSequence())

			frame.Reset()
			frame.Write([]byte("Hello from interop client in Go!"))
			err = group.WriteFrame(frame)
			if err != nil {
				slog.Error("failed to write frame", "error", err)
				return
			}

			group.Close()

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
	anns, err := sess.AcceptAnnounce("/")
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

	tr, err := sess.Subscribe(ann.BroadcastPath(), "", nil)
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
			for frame := range gr.Frames(nil) {
				slog.Info("Received a frame", "frame", string(frame.Body()))
			}
		}(gr)
	}

	sess.Terminate(moqt.NoError, moqt.NoError.String())
}

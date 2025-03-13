package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func runPublisher(sess moqt.Session, wg *sync.WaitGroup) {
	defer wg.Done()

	// Get the path of the track
	myPath := moqt.TrackPath("/kiu" + "/" + *name)

	// Accept an announce stream
	annstr, err := sess.AcceptAnnounceStream(context.Background())
	if err != nil {
		slog.Error("failed to accept announce stream", "error", err)
		return
	}

	// Announce the track
	err = annstr.SendAnnouncements([]*moqt.Announcement{moqt.NewAnnouncement(myPath)})
	if err != nil {
		slog.Error("failed to write announcements", "error", err)
		return
	}

	// Accept a track stream
	stream, err := sess.AcceptTrackStream(context.Background(), func(path moqt.TrackPath) (moqt.Info, error) {
		slog.Debug("subscribed to a track", slog.String("track_path", path.String()))
		// Accept a track stream only if the track path matches
		if myPath != path {
			return moqt.Info{}, moqt.ErrTrackDoesNotExist
		}

		info := moqt.Info{}
		slog.Debug("accepted a subscription", slog.String("info", info.String()))
		return info, nil
	})
	if err != nil {
		slog.Error("failed to accept track stream", "error", err)
		return
	}

	seq := moqt.FirstSequence
	for {
		w, err := stream.OpenGroup(seq)
		if err != nil {
			slog.Error("failed to open group", "error", err)
			break
		}

		typingRaw := appendHistory("<")

		input := make([]byte, 0, 1024)
		b := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(b)
			if err != nil {
				slog.Error("failed to read from stdin", "error", err)
				break
			}

			if b[0] == '\n' {
				break
			} else if b[0] == 127 { // backspace
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			} else {
				input = append(input, b[0])
			}

			// Redraw the screen
			updateHistory(typingRaw, string(input)+"<")
		}

		frame := moqt.NewFrame(input)

		err = w.WriteFrame(frame)
		if err != nil {
			slog.Error("failed to write frame", "error", err)
			break
		}

		slog.Debug("sent a frame", slog.String("frame", string(frame.CopyBytes())))

		// Release the frame
		frame.Release()

		// Append the message to the history
		updateHistory(typingRaw, *name+": "+string(input))

		seq = seq.Next()
	}
}

func appendHistory(s string) int {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	history = append(history, s)

	redraw("")

	return len(history)
}

func deleteHistory(raw int) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	if raw < 0 || raw >= len(history) {
		return
	}

	history = append(history[:raw], history[raw+1:]...)

	redraw("")
}

func updateHistory(raw int, s string) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	history[raw] = s
	historyMutex.Unlock()

	redraw("")
}

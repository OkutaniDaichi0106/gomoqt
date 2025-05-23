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
	myPath := moqt.BroadcastPath("/kiu" + "/" + *name)
	slog.Info("my path", slog.String("path", myPath.String()))
	track := moqt.BuildTrack(myPath, moqt.Info{}, 0)

	wg.Add(1)
	go func() {
		defer wg.Done()

		seq := moqt.FirstGroupSequence
		for {
			w, err := track.OpenGroup(seq)
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

			// Append the message to the history
			updateHistory(typingRaw, *name+": "+string(input))

			seq = seq.Next()
		}
	}()

	// Accept an announce stream
	annstr, annconf, err := sess.AcceptAnnounceStream(context.Background())
	if err != nil {
		slog.Error("failed to accept announce stream", "error", err)
		return
	}

	// Serve announcements
	wg.Add(1)
	go func() {
		defer wg.Done()
		moqt.ServeAnnouncements(annstr, annconf)
	}()

	for {
		// Accept a track stream
		stream, subconf, err := sess.AcceptTrackStream(context.Background())
		if err != nil {
			slog.Error("failed to accept track stream", "error", err)
			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("serving a track", slog.String("config", subconf.String()))
			moqt.ServeTrack(stream, subconf)
		}()
	}
	// // Accept a track stream
	// subconf, stream, err := sess.AcceptTrackStream(context.Background())
	// if err != nil {
	// 	slog.Error("failed to accept track stream", "error", err)
	// 	return
	// }

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

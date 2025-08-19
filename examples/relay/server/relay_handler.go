package main

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

var _ moqt.TrackHandler = (*relayHandler)(nil)

func newRelayHandler(tr *moqt.TrackReader) moqt.TrackHandler {
	ctx, cancel := context.WithCancel(context.Background())

	h := &relayHandler{
		ctx:   ctx,
		dests: make(map[*moqt.TrackWriter]struct{}),
	}

	go func() {
		defer cancel()
		defer tr.Close()

		for {
			gr, err := tr.AcceptGroup(ctx)
			if err != nil {
				return
			}

			go h.relayGroup(gr)
		}
	}()

	return h
}

type relayHandler struct {
	ctx context.Context

	mu    sync.RWMutex
	dests map[*moqt.TrackWriter]struct{}
}

func (h *relayHandler) ServeTrack(ctx context.Context, tw *moqt.TrackWriter) {
	h.mu.Lock()
	h.dests[tw] = struct{}{}
	h.mu.Unlock()

	<-h.ctx.Done()
}

func (h *relayHandler) relayGroup(gr *moqt.GroupReader) {

	seq := gr.GroupSequence()

	writers := make(chan *moqt.GroupWriter, len(h.dests))
	go func() {
		failedPubs := make([]*moqt.TrackWriter, 0, len(h.dests))

		h.mu.RLock()
		defer h.mu.RUnlock()

		for tw := range h.dests {
			gw, err := tw.OpenGroup(seq)
			if err != nil {
				failedPubs = append(failedPubs, tw)
				continue
			}
			writers <- gw
		}

		for _, pub := range failedPubs {
			delete(h.dests, pub) // Remove publishers that failed to open group
		}

		close(writers)
	}()

	// Relay frames
	frame := moqt.NewFrame(nil)
	for {
		err := gr.ReadFrame(frame)
		if err != nil {
			if err == io.EOF {
				for gw := range writers {
					gw.Close() // Close all writers if the group reader is closed
				}
				return
			}

			errCode := moqt.InternalGroupErrorCode

			var groupErr moqt.GroupError
			if errors.As(err, &groupErr) {
				errCode = groupErr.GroupErrorCode()
			}
			// If we fail to read a frame, we should cancel all writers.
			for gw := range writers {
				gw.CancelWrite(errCode)
			}

			return
		}

		for gw := range writers {
			gw.WriteFrame(frame)

			writers <- gw // Reinsert the writer back into the channel
		}
	}
}

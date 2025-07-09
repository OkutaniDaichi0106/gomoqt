package main

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

var _ moqt.TrackHandler = (*relayHandler)(nil)

func newRelayHandler(sub *moqt.Subscription) moqt.TrackHandler {
	ctx, cancel := context.WithCancel(context.Background())

	h := &relayHandler{
		ctx:   ctx,
		dests: make(map[*moqt.Publication]struct{}),
	}

	go func() {
		defer cancel()
		defer sub.TrackReader.Close()

		for {
			gr, err := sub.TrackReader.AcceptGroup(ctx)
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
	dests map[*moqt.Publication]struct{}
}

func (h *relayHandler) ServeTrack(pub *moqt.Publication) {
	h.mu.Lock()
	h.dests[pub] = struct{}{}
	h.mu.Unlock()

	<-h.ctx.Done()
}

func (h *relayHandler) relayGroup(gr moqt.GroupReader) {

	seq := gr.GroupSequence()

	writers := make(chan moqt.GroupWriter, len(h.dests))
	go func() {
		failedPubs := make([]*moqt.Publication, 0, len(h.dests))

		h.mu.RLock()
		defer h.mu.RUnlock()

		for pub := range h.dests {
			gw, err := pub.TrackWriter.OpenGroup(seq)
			if err != nil {
				failedPubs = append(failedPubs, pub)
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
	for {
		frame, err := gr.ReadFrame()
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

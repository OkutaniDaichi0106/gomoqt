package main

import (
	"context"
	"sync"

	"github.com/okdaichi/gomoqt/moqt"
)

func newRelayHandler(path moqt.BroadcastPath, sess *moqt.Session) *relayHandler {
	handler := &relayHandler{
		path:     path,
		sess:     sess,
		relaying: make(map[moqt.TrackName]*trackRelayer),
	}
	return handler
}

var _ moqt.TrackHandler = (*relayHandler)(nil)

type relayHandler struct {
	path moqt.BroadcastPath
	sess *moqt.Session

	mu       sync.RWMutex
	relaying map[moqt.TrackName]*trackRelayer
}

func (h *relayHandler) ServeTrack(tw *moqt.TrackWriter) {
	h.mu.Lock()
	if h.relaying == nil {
		h.relaying = make(map[moqt.TrackName]*trackRelayer)
	}

	tr, ok := h.relaying[tw.TrackName]
	if !ok {
		tr = h.fetch(tw.TrackName)
	}
	h.mu.Unlock()

	if tr == nil {
		tw.CloseWithError(moqt.TrackNotFoundErrorCode)
		return
	}

	tr.addDestination(tw)

	<-tw.Context().Done()

	tw.Close()

	tr.removeDestination(tw)
}

func (h *relayHandler) fetch(name moqt.TrackName) *trackRelayer {
	src, err := h.sess.Subscribe(h.path, name, nil)
	if err != nil {
		return nil
	}
	return newTrackRelayer(src, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.relaying, name)
	})
}

func newTrackRelayer(src *moqt.TrackReader, onClose func()) *trackRelayer {
	relayer := &trackRelayer{
		srcs:           src,
		dests:          make(map[*moqt.TrackWriter]struct{}),
		onClose:        onClose,
		lastGroupCache: newGroupCache(0),
	}

	go relayer.relay(context.Background())

	return relayer
}

type trackRelayer struct {
	mu    sync.RWMutex
	srcs  *moqt.TrackReader
	dests map[*moqt.TrackWriter]struct{}

	onClose func()

	lastGroupCache *groupCache
}

func (r *trackRelayer) addDestination(dest *moqt.TrackWriter) {
	r.mu.Lock()
	r.dests[dest] = struct{}{}
	r.mu.Unlock()

	gw, err := dest.OpenGroup(r.lastGroupCache.seq)
	if err != nil {
		r.removeDestination(dest)
		return
	}

	r.lastGroupCache.flush(gw)
}

func (r *trackRelayer) removeDestination(dest *moqt.TrackWriter) {
	r.mu.Lock()

	delete(r.dests, dest)

	if len(r.dests) == 0 {
		r.close()
	}
	r.mu.Unlock()
}

func (r *trackRelayer) close() {
	r.srcs.Close()
	for dest := range r.dests {
		dest.Close()
	}
	r.onClose()
}

func (r *trackRelayer) relay(ctx context.Context) {
	defer r.close()

	for {
		gr, err := r.srcs.AcceptGroup(ctx)
		if err != nil {
			return
		}

		currentSeq := gr.GroupSequence()

		r.lastGroupCache.update(currentSeq)

		r.mu.RLock()
		groups := make([]*moqt.GroupWriter, 0, len(r.dests))
		var failed []*moqt.TrackWriter
		for dest := range r.dests {
			gw, err := dest.OpenGroup(currentSeq)
			if err != nil {
				if failed == nil {
					failed = make([]*moqt.TrackWriter, 0, 1<<3)
				}
				failed = append(failed, dest)
				continue
			}

			groups = append(groups, gw)
		}
		r.mu.RUnlock()

		frame := moqt.NewFrame(0)
		for {
			err := gr.ReadFrame(frame)
			if err != nil {
				break
			}

			for _, gw := range groups {
				if err := gw.WriteFrame(frame); err != nil {
					continue
				}
			}

			if r.lastGroupCache.seq == currentSeq {
				r.lastGroupCache.addFrame(frame)
			}
		}
	}
}

func newGroupCache(seq moqt.GroupSequence) *groupCache {
	return &groupCache{
		seq:    seq,
		frames: make([]*moqt.Frame, 0),
		dests:  make(map[int][]*moqt.GroupWriter),
	}
}

type groupCache struct {
	mu     sync.RWMutex
	seq    moqt.GroupSequence
	frames []*moqt.Frame

	dests map[int][]*moqt.GroupWriter
}

func (gc *groupCache) addFrame(frame *moqt.Frame) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.frames = append(gc.frames, frame)

	frameCount := len(gc.frames)

	var err error
	for _, gw := range gc.dests[frameCount-1] {
		err = gw.WriteFrame(frame)
		if err != nil {
			continue
		}

		// Shift the current group writers for the next frame
		gc.dests[frameCount] = append(gc.dests[frameCount], gw)
	}

	// Clear the previous group writers
	gc.dests[frameCount-1] = gc.dests[frameCount-1][:0]
}

func (gc *groupCache) flush(gw *moqt.GroupWriter) {
	gc.mu.Lock()

	if gc.seq != gw.GroupSequence() {
		gc.mu.Unlock()
		return
	}

	frameCount := len(gc.frames)

	gc.dests[frameCount] = append(gc.dests[frameCount], gw)

	// Create a snapshot of current frames to avoid race conditions
	// This is efficient - just copies the slice header, not the data
	frames := make([]*moqt.Frame, len(gc.frames))
	copy(frames, gc.frames)

	gc.mu.Unlock()

	for _, frame := range frames {
		gw.WriteFrame(frame)
	}
}

func (gc *groupCache) update(seq moqt.GroupSequence) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if seq < gc.seq {
		return
	}

	gc.seq = seq
	frameCount := len(gc.frames)
	for k, groups := range gc.dests {
		last := k == frameCount
		for _, gw := range groups {
			if last {
				gw.Close()
			} else {
				gw.CancelWrite(moqt.ExpiredGroupErrorCode) // TODO: Use more appropriate error code
			}
		}
	}
	gc.frames = gc.frames[:0]
}

package moqt

import (
	"errors"
	"sync"
	"time"
)

func BuildTrack(path TrackPath, expires time.Duration) *TrackBuffer {
	return &TrackBuffer{
		path:           path,
		groupMap:       make(map[GroupSequence]*GroupBuffer),
		notifyChannels: make([]chan GroupSequence, 0, 1), // TODO: Tune the size
		expires:        expires,
		mu:             &sync.Mutex{},
	}
}

type TrackBuffer struct {
	path TrackPath
	// priority            TrackPriority
	latestGroupSequence GroupSequence
	// order               GroupOrder
	expires time.Duration

	groupMap map[GroupSequence]*GroupBuffer
	mapMu    *sync.RWMutex

	notifyChannels []chan GroupSequence
	chMu           *sync.Mutex

	closed    bool
	closedErr error

	mu *sync.Mutex
}

func (t *TrackBuffer) TrackPath() TrackPath {
	return t.path
}

// func (t *TrackBuffer) Info() Info {
// 	t.mu.Lock()
// 	info := Info{
// 		TrackPriority:       t.priority,
// 		GroupOrder:          t.order,
// 		LatestGroupSequence: t.latestGroupSequence,
// 	}
// 	t.mu.Unlock()
// 	return info
// }

// func (t *TrackBuffer) TrackPriority() TrackPriority {
// 	return t.priority
// }

// func (t *TrackBuffer) GroupOrder() GroupOrder {
// 	return t.order
// }

func (t *TrackBuffer) LatestGroupSequence() GroupSequence {
	t.mu.Lock()
	defer t.mu.Unlock()

	seq := t.latestGroupSequence
	return seq
}

func (t *TrackBuffer) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	count := len(t.groupMap)
	return count
}

func (t *TrackBuffer) NewTrackWriter(priority TrackPriority, order GroupOrder) TrackWriter {
	return &trackBufferWriter{
		trackBuffer: t,
	}
}

func (t *TrackBuffer) NewTrackReader(priority TrackPriority, order GroupOrder) TrackReader {
	t.chMu.Lock()
	defer t.chMu.Unlock()

	return newTrackBufferReader(t, order)
}

func (t *TrackBuffer) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

func (t *TrackBuffer) CloseWithError(err error) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrClosedTrack
	}
	t.closedErr = err
	t.closed = true
	return nil
}

func (t *TrackBuffer) storeGroup(gb *GroupBuffer) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrClosedTrack
	}
	t.groupMap[gb.GroupSequence()] = gb
	if gb.GroupSequence() > t.latestGroupSequence {
		t.latestGroupSequence = gb.GroupSequence()
	}
	// Always enqueue GroupSequence into notifyChannels (blocking send)
	for _, ch := range t.notifyChannels {
		ch <- gb.GroupSequence()
	}
	time.AfterFunc(t.expires, func() {
		t.removeGroup(gb.GroupSequence())
	})
	return nil
}

func (t *TrackBuffer) getGroup(seq GroupSequence) (*GroupBuffer, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.groupMap) == 0 {
		return nil, errors.New("no group buffers available")
	}
	gb, ok := t.groupMap[seq]
	if !ok {
		return nil, errors.New("group buffer not found")
	}
	return gb, nil
}

func (t *TrackBuffer) removeGroup(seq GroupSequence) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.groupMap, seq)
}

func (t *TrackBuffer) addNotifyChannel() chan GroupSequence {
	t.chMu.Lock()
	defer t.chMu.Unlock()

	ch := make(chan GroupSequence, 1)
	t.notifyChannels = append(t.notifyChannels, ch)
	return ch
}

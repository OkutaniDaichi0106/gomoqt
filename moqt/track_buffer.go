package moqt

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// var DefaultExpires = 10 * time.Second // TODO: Tune the value

func NewTrackBuffer(info Info, expires time.Duration) *TrackBuffer {
	// if expires == 0 {
	// 	expires = DefaultExpires
	// }
	buf := &TrackBuffer{
		expires:        expires,
		priority:       info.TrackPriority,
		order:          info.GroupOrder,
		groupMap:       make(map[GroupSequence]*GroupBuffer),
		notifyChannels: make([]chan GroupSequence, 0, 1<<4), // TODO: Tune the size
	}

	buf.latestGroupSequence.Store(uint64(info.LatestGroupSequence))

	return buf
}

var _ TrackWriter = (*TrackBuffer)(nil)
var _ TrackHandler = (*TrackBuffer)(nil)

type TrackBuffer struct {
	announcement        *Announcement
	latestGroupSequence atomic.Uint64
	expires             time.Duration

	priority TrackPriority
	order    GroupOrder

	groupMap map[GroupSequence]*GroupBuffer
	mapMu    sync.RWMutex

	notifyChannels []chan GroupSequence
	chMu           sync.Mutex

	closed    atomic.Bool
	closedErr error
}

func (t *TrackBuffer) Info() (Info, bool) {
	if !t.isReadable() {
		return Info{}, false
	}
	return Info{
		LatestGroupSequence: GroupSequence(t.latestGroupSequence.Load()),
		TrackPriority:       t.priority,
		GroupOrder:          t.order,
	}, true
}

func (t *TrackBuffer) Announcemnt() *Announcement {
	return t.announcement
}

func (t *TrackBuffer) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if !t.isWritable() {
		return nil, ErrClosedTrack
	}

	gb := newGroupBuffer(seq, DefaultGroupBufferSize)

	err := t.storeGroup(gb)
	if err != nil {
		return nil, err
	}

	return gb, nil
}

func (t *TrackBuffer) SendGap(gap Gap) error {
	if !t.isWritable() {
		return ErrClosedTrack
	}

}

func (t *TrackBuffer) Close() error {
	if t.closed.Load() {
		return ErrClosedTrack
	}

	t.closed.Store(true)
	for _, ch := range t.notifyChannels {
		close(ch)
	}
	t.notifyChannels = nil

	return nil
}

func (t *TrackBuffer) CloseWithError(err error) error {
	if t.closed.Load() {
		return ErrClosedTrack
	}

	t.closedErr = err
	t.closed.Store(true)

	for _, ch := range t.notifyChannels {
		close(ch)
	}
	t.notifyChannels = nil

	return nil
}

func (t *TrackBuffer) HandleTrack(w TrackWriter, sub SendTrackStream) {
	if !t.isReadable() {
		w.CloseWithError(ErrEndedTrack)
		return
	}

	r := newTrackBufferReader(t, config)
	defer r.Close()

	ctx := context.Background()
	for {
		gr, err := r.AcceptGroup(ctx)
		if err != nil {
			return
		}

		// Check if GroupSequence is in range
		if !config.IsInRange(gr.GroupSequence()) {
			continue
		}

		// Open a new group writer
		gw, err := w.OpenGroup(gr.GroupSequence())
		if err != nil {
			return
		}

		for {
			f, err := gr.ReadFrame()
			if err != nil {
				break
			}

			err = gw.WriteFrame(f)
			if err != nil {
				break
			}
		}
	}
}

func (t *TrackBuffer) storeGroup(gb *GroupBuffer) error {
	t.mapMu.Lock()
	defer t.mapMu.Unlock()
	if t.closed.Load() {
		if t.closedErr != nil {
			return t.closedErr
		}
		return ErrClosedTrack
	}
	t.groupMap[gb.GroupSequence()] = gb
	if gb.GroupSequence() > GroupSequence(t.latestGroupSequence.Load()) {
		t.latestGroupSequence.Store(uint64(gb.GroupSequence()))
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

func (t *TrackBuffer) getGroup(seq GroupSequence) (*GroupBuffer, bool) {
	t.mapMu.Lock()
	defer t.mapMu.Unlock()

	if len(t.groupMap) == 0 {
		return nil, false
	}

	gb, ok := t.groupMap[seq]
	if !ok {
		return nil, false
	}
	return gb, true
}

func (t *TrackBuffer) removeGroup(seq GroupSequence) {
	t.mapMu.Lock()
	defer t.mapMu.Unlock()
	delete(t.groupMap, seq)

	// Unannounce if no more groups
	if !t.isReadable() {
		t.announcement.End()
	}
}

func (t *TrackBuffer) isWritable() bool {
	return !t.closed.Load()
}

func (t *TrackBuffer) isReadable() bool {
	return !t.closed.Load() && len(t.groupMap) > 0
}

func (t *TrackBuffer) addNotifyChannel() chan GroupSequence {
	t.chMu.Lock()
	defer t.chMu.Unlock()

	ch := make(chan GroupSequence, 1<<2)
	t.notifyChannels = append(t.notifyChannels, ch)
	return ch
}

// removeNotifyChannel removes the notify channel from the list and closes it.
func (t *TrackBuffer) removeNotifyChannel(ch chan GroupSequence) { // TODO: Use this function
	t.chMu.Lock()
	defer t.chMu.Unlock()

	for i, c := range t.notifyChannels {
		if c == ch {
			t.notifyChannels = append(t.notifyChannels[:i], t.notifyChannels[i+1:]...)
			close(c)
			return
		}
	}
}

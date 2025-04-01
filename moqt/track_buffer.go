package moqt

import (
	"context"
	"sync"
	"time"
)

var DefaultExpires = 10 * time.Second

func BuildTrack(path TrackPath, info Info, expires time.Duration) *TrackBuffer {
	if expires == 0 {
		expires = DefaultExpires
	}

	return &TrackBuffer{
		path:                path,
		latestGroupSequence: info.LatestGroupSequence,
		expires:             expires,
		remoteTrackPriority: info.TrackPriority,
		remoteGroupOrder:    info.GroupOrder,
		groupMap:            make(map[GroupSequence]*GroupBuffer),
		mapMu:               &sync.RWMutex{},
		notifyChannels:      make([]chan GroupSequence, 0, 1), // TODO: Tune the size
		chMu:                &sync.Mutex{},
		announcement:        NewAnnouncement(path),
		mu:                  &sync.RWMutex{},
	}
}

var _ TrackHandler = (*TrackBuffer)(nil)

type TrackBuffer struct {
	path                TrackPath
	latestGroupSequence GroupSequence
	expires             time.Duration

	remoteTrackPriority TrackPriority
	remoteGroupOrder    GroupOrder

	groupMap map[GroupSequence]*GroupBuffer
	mapMu    *sync.RWMutex

	notifyChannels []chan GroupSequence
	chMu           *sync.Mutex

	announcement *Announcement

	closed    bool
	closedErr error

	mu *sync.RWMutex
}

func (t *TrackBuffer) TrackPath() TrackPath {
	return t.path
}

func (t *TrackBuffer) Info() Info {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return Info{
		LatestGroupSequence: t.latestGroupSequence,
		TrackPriority:       t.remoteTrackPriority,
		GroupOrder:          t.remoteGroupOrder,
	}
}

func (t *TrackBuffer) LatestGroupSequence() GroupSequence {
	t.mu.RLock()
	defer t.mu.RUnlock()

	seq := t.latestGroupSequence
	return seq
}

func (t *TrackBuffer) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	count := len(t.groupMap)
	return count
}

func (t *TrackBuffer) NewTrackWriter(priority TrackPriority, order GroupOrder) (TrackWriter, error) {
	if !t.isWritable() {
		return nil, ErrClosedTrack
	}

	return newTrackBufferWriter(t), nil
}

func (t *TrackBuffer) NewTrackReader(priority TrackPriority, order GroupOrder) (TrackReader, error) {
	if !t.isReadable() {
		return nil, ErrEndedTrack
	}

	return newTrackBufferReader(t, order), nil
}

func (t *TrackBuffer) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrClosedTrack
	}

	t.closed = true
	for _, ch := range t.notifyChannels {
		close(ch)
	}
	t.notifyChannels = nil

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

	for _, ch := range t.notifyChannels {
		close(ch)
	}
	t.notifyChannels = nil

	return nil
}

func (t *TrackBuffer) ServeTrack(w TrackWriter, config SubscribeConfig) {
	r, err := t.NewTrackReader(config.TrackPriority, config.GroupOrder)
	if err != nil {
		w.CloseWithError(err)
		return
	}

	ctx := context.Background() // TODO: Cancel in 10 seconds
	for {
		gr, err := r.AcceptGroup(ctx)
		if err != nil {
			return
		}

		// Check if GroupSequence is in range
		if !gr.GroupSequence().IsInRange(config.MinGroupSequence, config.MaxGroupSequence) {
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

func (t *TrackBuffer) ServeAnnouncements(w AnnouncementWriter) {
	announcement := t.announcement
	if !announcement.IsActive() {
		return
	}

	if announcement.TrackPath().Match(w.AnnounceConfig().TrackPattern) {
		w.SendAnnouncements([]*Announcement{t.announcement})
	}
}

func (t *TrackBuffer) GetInfo(path TrackPath) (Info, error) {
	return t.Info(), nil
}

func (t *TrackBuffer) storeGroup(gb *GroupBuffer) error {
	t.mapMu.Lock()
	defer t.mapMu.Unlock()
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

// func (t *TrackBuffer) unannounce() {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	announcements := []*Announcement{newEndedAnnouncement(t.TrackPath())}

// 	for _, a := range t.announced {
// 		a.WriteAnnouncement(announcements)
// 	}

// 	t.announced = nil
// }

func (t *TrackBuffer) isWritable() bool {
	return !t.closed
}

func (t *TrackBuffer) isReadable() bool {
	return !t.closed && len(t.groupMap) > 0
}

func (t *TrackBuffer) addNotifyChannel() chan GroupSequence {
	t.chMu.Lock()
	defer t.chMu.Unlock()

	ch := make(chan GroupSequence, 1)
	t.notifyChannels = append(t.notifyChannels, ch)
	return ch
}

// removeNotifyChannel removes the notify channel from the list and closes it.
func (t *TrackBuffer) removeNotifyChannel(ch chan GroupSequence) { // TODO: Use this function
	defer t.chMu.Unlock()

	for i, c := range t.notifyChannels {
		if c == ch {
			t.notifyChannels = append(t.notifyChannels[:i], t.notifyChannels[i+1:]...)
			close(c)
			return
		}
	}
}

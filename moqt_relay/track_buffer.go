package moqtrelay

import (
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func NewTrackBuffer(subscription moqt.Subscription) *TrackBuffer {
	return &TrackBuffer{
		groupBufs:    make(map[moqt.GroupSequence]GroupBuffer),
		subscription: subscription,
	}
}

type TrackBuffer struct {
	groupBufs    map[moqt.GroupSequence]GroupBuffer
	mu           sync.Mutex
	subscription moqt.Subscription
}

func (t *TrackBuffer) AddGroup(g GroupBuffer) error {
	// Check if the group sequence is in the range
	if t.subscription.MinGroupSequence != 0 && t.subscription.MinGroupSequence > g.GroupSequence() {
		return moqt.ErrInvalidRange
	}
	if t.subscription.MaxGroupSequence != 0 && t.subscription.MaxGroupSequence < g.GroupSequence() {
		return moqt.ErrInvalidRange
	}

	// Check if the group sequence is duplicated
	if _, ok := t.groupBufs[g.GroupSequence()]; ok {
		return moqt.ErrDuplicatedGroup
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Add the sequence
	t.groupBufs[g.GroupSequence()] = g

	return nil
}

func (t *TrackBuffer) RemoveGroup(seq moqt.GroupSequence) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if the group sequence exists
	g, ok := t.groupBufs[seq]
	if !ok {
		return
	}

	err := g.Close()
	if err != nil {
		slog.Error("failed to close the group buffer", slog.String("error", err.Error()))
		return
	}

	// Remove the sequence
	delete(t.groupBufs, seq)
}

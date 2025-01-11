package moqt

import (
	"log/slog"
	"sync"
)

func NewTrackBuffer(subscription Subscription) *TrackBuffer {
	return &TrackBuffer{
		groupBufs:    make(map[GroupSequence]GroupBuffer),
		subscription: subscription,
	}
}

type TrackBuffer struct {
	groupBufs    map[GroupSequence]GroupBuffer
	mu           sync.Mutex
	subscription Subscription
}

func (t *TrackBuffer) AddGroup(g GroupBuffer) error {
	// Check if the group sequence is in the range
	if t.subscription.MinGroupSequence != 0 && t.subscription.MinGroupSequence > g.GroupSequence() {
		return ErrInvalidRange
	}
	if t.subscription.MaxGroupSequence != 0 && t.subscription.MaxGroupSequence < g.GroupSequence() {
		return ErrInvalidRange
	}

	// Check if the group sequence is duplicated
	if _, ok := t.groupBufs[g.GroupSequence()]; ok {
		return ErrDuplicatedGroup
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Add the sequence
	t.groupBufs[g.GroupSequence()] = g

	return nil
}

func (t *TrackBuffer) RemoveGroup(seq GroupSequence) {
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

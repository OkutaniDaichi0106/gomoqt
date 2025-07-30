package moqt

import (
	"context"
	"strings"
	"sync"
)

func NewAnnouncement(ctx context.Context, path BroadcastPath) *Announcement {
	ctx, cancel := context.WithCancel(ctx)

	ann := Announcement{
		path:   path,
		ctx:    ctx,
		cancel: cancel,
	}

	return &ann
}

type Announcement struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	path BroadcastPath

	onEndFuncs []func()
}

func (a *Announcement) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	sb.WriteString(" ")
	sb.WriteString("announce_status: ")
	if a.IsActive() {
		sb.WriteString("active")
	} else {
		sb.WriteString("ended")
	}
	sb.WriteString(", ")
	sb.WriteString("broadcast_path: ")
	sb.WriteString(a.path.String())
	sb.WriteString(" }")
	return sb.String()
}

func (a *Announcement) BroadcastPath() BroadcastPath {
	return a.path
}

func (a *Announcement) AwaitEnd() <-chan struct{} {
	return a.ctx.Done()
}

func (a *Announcement) OnEnd(f func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ctx.Err() != nil {
		f()
		return
	}
	a.onEndFuncs = append(a.onEndFuncs, f)
}

func (a *Announcement) IsActive() bool {
	return a.ctx.Err() == nil
}

func (a *Announcement) End() {
	a.cancel()
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, f := range a.onEndFuncs {
		f()
	}
}

func (a *Announcement) Fork() *Announcement {
	return NewAnnouncement(a.ctx, a.path)
}

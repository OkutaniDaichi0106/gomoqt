package moqt

import (
	"context"
	"strings"
)

func NewAnnouncement(ctx context.Context, path BroadcastPath) *Announcement {
	ctx, cancel := context.WithCancel(ctx)

	ann := Announcement{
		path:   path,
		ctx:    ctx,
		cancel: cancel,
	}

	if ctx.Err() != nil {

	}

	return &ann
}

type Announcement struct {
	ctx    context.Context
	cancel context.CancelFunc

	path BroadcastPath

	info Info
}

func (a *Announcement) String() string {
	var sb strings.Builder
	sb.WriteString("Announcement: {")
	sb.WriteString(" ")
	sb.WriteString("AnnounceStatus: ")
	if a.IsActive() {
		sb.WriteString("ACTIVE")
	} else {
		sb.WriteString("ENDED")
	}
	sb.WriteString(", ")
	sb.WriteString("BroadcastPath: ")
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

func (a *Announcement) IsActive() bool {
	return a.ctx.Err() == nil
}

func (a *Announcement) End() {
	a.cancel()
}

func (a *Announcement) Fork() *Announcement {
	return NewAnnouncement(a.ctx, a.path)
}

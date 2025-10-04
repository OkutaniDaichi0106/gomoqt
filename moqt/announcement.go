package moqt

import (
	"context"
	"runtime"
	"strings"
	"sync"
)

type EndAnnouncementFunc func()

func NewAnnouncement(ctx context.Context, path BroadcastPath) (*Announcement, EndAnnouncementFunc) {
	if !isValidPath(path) {
		panic("[Announcement] invalid track path: " + string(path))
	}

	ctx, cancel := context.WithCancel(ctx)

	ann := Announcement{
		path:   path,
		ctx:    ctx,
		cancel: cancel,
	}

	return &ann, ann.end
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

func (a *Announcement) Context() context.Context {
	return a.ctx
}

func (a *Announcement) OnEnd(f func()) {
	if a.ctx.Err() != nil {
		f()
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.onEndFuncs = append(a.onEndFuncs, f)
}

func (a *Announcement) IsActive() bool {
	return a.ctx.Err() == nil
}

func (a *Announcement) end() {
	a.cancel()
	a.mu.Lock()
	defer a.mu.Unlock()

	workerCount := min(runtime.NumCPU(), len(a.onEndFuncs))
	if workerCount == 0 {
		workerCount = 1
	}

	// buffer jobs to avoid blocking producers when many workers are used
	jobs := make(chan func(), len(a.onEndFuncs))

	var wg sync.WaitGroup

	// spawn workerCount goroutines
	for i := 0; i < workerCount; i++ {
		go func() {
			for f := range jobs {
				f()
				wg.Done()
			}
		}()
	}

	for _, f := range a.onEndFuncs {
		wg.Add(1)
		jobs <- f
	}

	close(jobs)

	wg.Wait()
}

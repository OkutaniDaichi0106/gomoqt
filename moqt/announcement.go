package moqt

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

type EndAnnouncementFunc func()

func NewAnnouncement(ctx context.Context, path BroadcastPath) (*Announcement, EndAnnouncementFunc) {
	if !isValidPath(path) {
		panic("[Announcement] invalid track path: " + string(path))
	}

	ann := Announcement{
		path: path,
		ch:   make(chan struct{}),
	}
	if ctx.Err() != nil {
		ann.active.Store(false)
	} else {
		ann.active.Store(true)
	}
	endFunc := func() { ann.end() }

	context.AfterFunc(ctx, endFunc)

	return &ann, endFunc
}

type Announcement struct {
	mu sync.Mutex
	ch chan struct{}

	path BroadcastPath

	afterHandlers map[*afterHandler]struct{}

	active atomic.Bool
	once   sync.Once
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

func (a *Announcement) Done() <-chan struct{} {
	return a.ch
}

func (a *Announcement) AfterFunc(f func()) (stop func() bool) {
	if !a.active.Load() {
		f()
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	handler := &afterHandler{op: f}
	if a.afterHandlers == nil {
		a.afterHandlers = make(map[*afterHandler]struct{})
	}
	a.afterHandlers[handler] = struct{}{}

	stopFunc := func() bool {
		a.mu.Lock()
		defer a.mu.Unlock()
		if _, exists := a.afterHandlers[handler]; exists {
			delete(a.afterHandlers, handler)
			return true
		}
		return false
	}

	return stopFunc
}

func (a *Announcement) IsActive() bool {
	return a.active.Load()
}

func (a *Announcement) end() {
	// set active to false
	a.active.Store(false)

	a.mu.Lock()
	defer a.mu.Unlock()

	workerCount := min(runtime.NumCPU(), len(a.afterHandlers))
	if workerCount == 0 {
		workerCount = 1
	}

	// buffer jobs to avoid blocking producers when many workers are used
	jobs := make(chan func(), len(a.afterHandlers))

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

	for handler := range a.afterHandlers {
		wg.Add(1)
		jobs <- handler.op
	}

	close(jobs)

	wg.Wait()

	a.once.Do(func() { close(a.ch) })
}

type afterHandler struct {
	op func()
}

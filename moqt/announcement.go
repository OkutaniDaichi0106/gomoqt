package moqt

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// EndAnnouncementFunc is a function that ends an announcement.
type EndAnnouncementFunc func()

// NewAnnouncement constructs a new Announcement for the given broadcast path.
//
// The returned *Announcement remains active until the provided context is canceled or the
// returned EndAnnouncementFunc is called.
// When the announcement ends, its Done() channel is closed once and all callbacks registered via
// AfterFunc are executed exactly once.
// If ctx is already canceled, the Announcement is initially inactive.
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

// Announcement represents the lifecycle of a broadcast.
//
// The key behaviors are:
// - Done() returns a channel that is closed once when the announcement ends.
// - AfterFunc registers a callback to be invoked once when the announcement ends.
// - If the announcement has already ended, the callback is invoked synchronously.
// - The stop function (returned by AfterFunc) removes the callback if it hasn't executed yet and returns true; otherwise it returns false.
//
// Methods are safe to call concurrently.
// AfterFunc and the returned stop function are safe to call concurrently with end().
// end() is idempotent and handlers execute at most once.
type Announcement struct {
	mu sync.Mutex
	ch chan struct{}

	path BroadcastPath

	afterHandlers map[*afterHandler]struct{}

	active atomic.Bool
	once   sync.Once
}

// String returns a string representation of the announcement for debugging.
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

// BroadcastPath returns the broadcast path associated with the announcement.
// The returned path is immutable for the lifetime of the announcement and is
// safe to copy.
func (a *Announcement) BroadcastPath() BroadcastPath {
	return a.path
}

// Done returns a channel that is closed once when the announcement ends.
// Consumers may use this channel to wait until the announcement stops.
func (a *Announcement) Done() <-chan struct{} {
	return a.ch
}

// AfterFunc registers a callback f to be invoked once when the announcement
// ends. If the announcement is active, f is queued and a stop function is
// returned that removes the registration (returns true when removed).
// If the announcement already ended, f is called synchronously and the stop
// function always returns false. AfterFunc and the returned stop function are
// safe to call concurrently with end().
func (a *Announcement) AfterFunc(f func()) (stop func() bool) {
	// Synchronize access to afterHandlers to avoid races with end()
	a.mu.Lock()
	if !a.active.Load() {
		a.mu.Unlock()
		// Announcement already ended — call immediately without registering
		f()
		return func() bool { return false }
	}

	handler := &afterHandler{op: f}
	if a.afterHandlers == nil {
		a.afterHandlers = make(map[*afterHandler]struct{})
	}
	a.afterHandlers[handler] = struct{}{}
	a.mu.Unlock()

	stopFunc := func() bool {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.afterHandlers == nil {
			return false
		}
		if _, exists := a.afterHandlers[handler]; exists {
			delete(a.afterHandlers, handler)
			return true
		}
		return false
	}

	return stopFunc
}

// IsActive reports whether the announcement is currently active.
// Returns false once the announcement has ended.
func (a *Announcement) IsActive() bool {
	return a.active.Load()
}

func (a *Announcement) end() {
	// Ensure end() body only runs once
	a.once.Do(func() {
		// set active to false
		a.active.Store(false)

		// Snapshot handlers under lock, then clear to avoid double-execution and races
		a.mu.Lock()
		handlers := a.afterHandlers
		a.afterHandlers = nil
		a.mu.Unlock()

		// Guard against nil map
		handlerCount := 0
		if handlers != nil {
			handlerCount = len(handlers)
		}

		// Determine worker count
		workerCount := min(runtime.NumCPU(), handlerCount)
		if workerCount == 0 {
			workerCount = 1
		}

		// buffer jobs to avoid blocking producers when many workers are used
		jobs := make(chan func(), handlerCount)

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

		for handler := range handlers {
			wg.Add(1)
			jobs <- handler.op
		}

		close(jobs)

		wg.Wait()

		// Close the Done channel once (since the once.Do wraps this function)
		close(a.ch)
	})
}

type afterHandler struct {
	op func()
}

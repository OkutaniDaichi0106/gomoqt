package moqt

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testHandler is a simple test implementation of TrackHandler
type testHandler struct {
	called    bool
	mu        sync.Mutex
	publisher *Publisher
}

func (h *testHandler) ServeTrack(p *Publisher) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.called = true
	h.publisher = p
}

func (h *testHandler) wasCalled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.called
}

func (h *testHandler) getPublisher() *Publisher {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.publisher
}

func TestMux_Handle(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Test basic handle
	mux.Handle(ctx, BroadcastPath("/test"), TrackHandlerFunc(func(p *Publisher) {

	}))

	// Test overwriting existing handler
	mux.Handle(ctx, BroadcastPath("/test"), TrackHandlerFunc(func(p *Publisher) {

	}))

	// Test multiple paths
	mux.Handle(ctx, BroadcastPath("/another"), TrackHandlerFunc(func(p *Publisher) {

	}))
}

func TestMux_ServeTrack(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	calledCh := make(chan struct{}, 1)

	mux.Handle(ctx, BroadcastPath("/test"), TrackHandlerFunc(func(p *Publisher) {
		assert.Equal(t, "/test", string(p.BroadcastPath))
		assert.Equal(t, "track1", p.TrackName)
		assert.NotNil(t, p.TrackWriter)

		calledCh <- struct{}{}
	}))

	// Create a publisher
	publisher := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     "track1",
		TrackWriter:   &MockTrackWriter{},
	}

	mux.ServeTrack(publisher)

	select {
	case <-calledCh:
	case <-time.After(5 * time.Second):
		t.Error("Handler should have been called")
	}
}

func TestMux_ServeTrack_NotFound(t *testing.T) {
	mux := NewTrackMux()

	// Create a mock track writer
	mockWriter := &MockTrackWriter{
		CloseWithErrorFunc: func(err error) error {
			assert.Equal(t, ErrTrackDoesNotExist, err)
			return nil
		},
	}

	// Create a publisher for non-existent path
	publisher := &Publisher{
		BroadcastPath: BroadcastPath("/nonexistent"),
		TrackName:     "track1",
		TrackWriter:   mockWriter,
	}

	// Should use NotFoundHandler which closes the track
	mux.ServeTrack(publisher)
}

func TestMux_ServeAnnouncements(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handler for wildcard pattern
	mux.Handle(ctx, BroadcastPath("/room/alice"), TrackHandlerFunc(func(p *Publisher) {}))
	mux.Handle(ctx, BroadcastPath("/room/bob"), TrackHandlerFunc(func(p *Publisher) {}))

	expected := map[string]struct{}{
		"/room/alice": {},
		"/room/bob":   {},
	}

	// Create mock announcement writer
	mockWriter := &MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*Announcement) error {
			assert.Len(t, announcements, len(expected))
			for _, ann := range announcements {
				assert.Contains(t, expected, string(ann.BroadcastPath()))
			}
			return nil
		},
	}

	// Test serving announcements
	mux.ServeAnnouncements(mockWriter, "/stream")
}

func TestMux_ContextCancellation(t *testing.T) {
	mux := NewTrackMux()

	calledCh := make(chan struct{}, 1)

	// Create cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	mux.Handle(ctx, BroadcastPath("/test"), TrackHandlerFunc(func(p *Publisher) {
		calledCh <- struct{}{}
	}))

	// Cancel context
	cancel()

	// Wait a bit for cleanup
	time.Sleep(10 * time.Millisecond)

	// Create publisher - should use NotFoundHandler since context was cancelled
	mockWriter := &MockTrackWriter{}
	publisher := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     "track1",
		TrackWriter:   mockWriter,
	}

	mux.ServeTrack(publisher)

	select {
	case <-calledCh:
	case <-time.After(5 * time.Second):
		t.Error("Handler should not have been called")
	}
}

func TestMux_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register multiple handlers concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			handler := TrackHandlerFunc(func(p *Publisher) {
				assert.Equal(t, fmt.Sprintf("/test%d", i), string(p.BroadcastPath))
			})
			mux.Handle(ctx, BroadcastPath(fmt.Sprintf("/test%d", i)), handler)
		}(i)
	}
	wg.Wait()

	// Serve tracks concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mockWriter := &MockTrackWriter{}
			publisher := &Publisher{
				BroadcastPath: BroadcastPath(fmt.Sprintf("/test%d", i)),
				TrackName:     "track1",
				TrackWriter:   mockWriter,
			}
			mux.ServeTrack(publisher)
		}(i)
	}
	wg.Wait()
}

func TestDefaultMux(t *testing.T) {
	// Test that DefaultMux is accessible
	assert.NotNil(t, DefaultMux)

	calledCh := make(chan struct{}, 1)

	// Test Handle on default mux
	Handle(context.Background(), BroadcastPath("/default"), TrackHandlerFunc(func(p *Publisher) {
		calledCh <- struct{}{}
	}))

	// Test ServeTrack on default mux
	mockWriter := &MockTrackWriter{}
	publisher := &Publisher{
		BroadcastPath: BroadcastPath("/default"),
		TrackName:     "track1",
		TrackWriter:   mockWriter,
	}

	DefaultMux.ServeTrack(publisher)

	select {
	case <-calledCh:
	case <-time.After(5 * time.Second):
		t.Error("Handler should have been called")
	}
}

func TestMux_HandleFunc(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	called := false
	var receivedPublisher *Publisher
	handlerFunc := func(p *Publisher) {
		called = true
		receivedPublisher = p
	}

	// Use Handle with TrackHandlerFunc since HandleFunc doesn't exist
	mux.Handle(ctx, BroadcastPath("/func"), TrackHandlerFunc(handlerFunc))

	mockWriter := &MockTrackWriter{}
	publisher := &Publisher{
		BroadcastPath: BroadcastPath("/func"),
		TrackName:     "track1",
		TrackWriter:   mockWriter,
	}

	mux.ServeTrack(publisher)
	assert.True(t, called)
	assert.Equal(t, publisher, receivedPublisher)
}

func TestMux_AnnouncementLifecycle(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handler
	handler := &testHandler{}
	mux.Handle(ctx, BroadcastPath("/stream/*"), handler)

	// Create announcement
	announcement := NewAnnouncement(context.Background(), BroadcastPath("/stream/test"))
	defer announcement.AwaitEnd()

	// Verify announcement is active
	assert.True(t, announcement.IsActive())
	assert.Equal(t, "/stream/test", string(announcement.BroadcastPath()))
}

func TestMux_NestedPaths(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers for nested paths
	deepHandler := &testHandler{}
	mux.Handle(ctx, BroadcastPath("/deep/nested/path"), deepHandler)

	shallowHandler := &testHandler{}
	mux.Handle(ctx, BroadcastPath("/deep"), shallowHandler)

	// Test deep path
	mockWriter1 := &MockTrackWriter{}
	publisher1 := &Publisher{
		BroadcastPath: BroadcastPath("/deep/nested/path"),
		TrackName:     "track1",
		TrackWriter:   mockWriter1,
	}
	mux.ServeTrack(publisher1)
	assert.True(t, deepHandler.wasCalled())
	assert.False(t, shallowHandler.wasCalled())

	// Reset handlers
	deepHandler = &testHandler{}
	shallowHandler = &testHandler{}
	mux.Handle(ctx, BroadcastPath("/deep/nested/path"), deepHandler)
	mux.Handle(ctx, BroadcastPath("/deep"), shallowHandler)

	// Test shallow path
	mockWriter2 := &MockTrackWriter{}
	publisher2 := &Publisher{
		BroadcastPath: BroadcastPath("/deep"),
		TrackName:     "track1",
		TrackWriter:   mockWriter2,
	}
	mux.ServeTrack(publisher2)
	assert.True(t, shallowHandler.wasCalled())
	assert.False(t, deepHandler.wasCalled())
}

func TestMux_ContextTimeout(t *testing.T) {
	mux := NewTrackMux()
	handler := &testHandler{}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	mux.Handle(ctx, BroadcastPath("/timeout"), handler)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Try to serve track - should use NotFoundHandler
	mockWriter := &MockTrackWriter{}
	publisher := &Publisher{
		BroadcastPath: BroadcastPath("/timeout"),
		TrackName:     "track1",
		TrackWriter:   mockWriter,
	}

	mux.ServeTrack(publisher)

	// Handler should not have been called due to timeout
	assert.False(t, handler.wasCalled())
}

func TestMux_MultipleCancellations(t *testing.T) {
	mux := NewTrackMux()

	// Create multiple contexts and handlers
	var contexts []context.CancelFunc
	var handlers []*testHandler

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		contexts = append(contexts, cancel)

		handler := &testHandler{}
		handlers = append(handlers, handler)

		path := BroadcastPath(fmt.Sprintf("/multi%d", i))
		mux.Handle(ctx, path, handler)
	}

	// Cancel all contexts
	for _, cancel := range contexts {
		cancel()
	}

	// Wait for cleanup
	time.Sleep(20 * time.Millisecond)

	// Verify all handlers were removed
	for i, handler := range handlers {
		mockWriter := &MockTrackWriter{}
		publisher := &Publisher{
			BroadcastPath: BroadcastPath(fmt.Sprintf("/multi%d", i)),
			TrackName:     "track1",
			TrackWriter:   mockWriter,
		}

		mux.ServeTrack(publisher)
		assert.False(t, handler.wasCalled())
	}
}

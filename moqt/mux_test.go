package moqt

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, TrackName("track1"), p.TrackName)
		assert.NotNil(t, p.TrackWriter)
		assert.NotNil(t, p.SubscribeStream)

		calledCh <- struct{}{}
	}))

	// Create a publisher
	publisher := &Publisher{
		BroadcastPath:   BroadcastPath("/test"),
		TrackName:       "track1",
		TrackWriter:     &MockTrackWriter{},
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
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
		BroadcastPath:   BroadcastPath("/nonexistent"),
		TrackName:       "track1",
		TrackWriter:     mockWriter,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
	}

	// Should use NotFoundHandler which closes the track
	mux.ServeTrack(publisher)
}

func TestMux_ServeAnnouncements(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers for paths that start with /room
	mux.Handle(ctx, BroadcastPath("/room/alice"), TrackHandlerFunc(func(p *Publisher) {}))
	mux.Handle(ctx, BroadcastPath("/room/bob"), TrackHandlerFunc(func(p *Publisher) {}))

	announced := make([]*Announcement, 0)

	// Create mock announcement writer
	mockWriter := &MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*Announcement) error {
			announced = append(announced, announcements...)
			if len(announced) == 2 {
				return fmt.Errorf("sent all announcements")
			}
			return nil
		},
	}

	// Test serving announcements
	mux.ServeAnnouncements(mockWriter, "/room")
}

func TestMux_ContextCancellation(t *testing.T) {
	mux := NewTrackMux()

	notFoundCalled := make(chan struct{}, 1)
	testHandlerCalled := make(chan struct{}, 1)

	NotFoundHandler = TrackHandlerFunc(func(p *Publisher) {
		notFoundCalled <- struct{}{}
	})

	// Create cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	mux.Handle(ctx, BroadcastPath("/test"), TrackHandlerFunc(func(p *Publisher) {
		testHandlerCalled <- struct{}{}
	}))

	// Cancel context
	cancel()

	// Wait a bit for cleanup
	time.Sleep(10 * time.Millisecond)

	// Create publisher - should use NotFoundHandler since context was cancelled
	mockWriter := &MockTrackWriter{}
	publisher := &Publisher{
		BroadcastPath:   BroadcastPath("/test"),
		TrackName:       "track1",
		TrackWriter:     mockWriter,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
	}

	mux.ServeTrack(publisher)

	select {
	case <-notFoundCalled:
		// Expected
	case <-testHandlerCalled:
		t.Error("Test handler should not have been called after context cancellation")
	case <-time.After(5 * time.Second):
		t.Error("Not found handler should have been called")
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
				BroadcastPath:   BroadcastPath(fmt.Sprintf("/test%d", i)),
				TrackName:       "track1",
				TrackWriter:     mockWriter,
				SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
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
		BroadcastPath:   BroadcastPath("/default"),
		TrackName:       "track1",
		TrackWriter:     mockWriter,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
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
		BroadcastPath:   BroadcastPath("/func"),
		TrackName:       "track1",
		TrackWriter:     mockWriter,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
	}

	mux.ServeTrack(publisher)
	assert.True(t, called)
	assert.Equal(t, publisher, receivedPublisher)
}

func TestMux_AnnouncementLifecycle(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handler
	mux.Handle(ctx, "/room", TrackHandlerFunc(func(p *Publisher) {}))

	// Create announcement
	announcement := NewAnnouncement(context.Background(), BroadcastPath("/room/test"))
	defer announcement.AwaitEnd()

	// Verify announcement is active
	assert.True(t, announcement.IsActive())
	assert.Equal(t, "/room/test", string(announcement.BroadcastPath()))
}

func TestMux_NestedPaths(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Track which handlers are called using channels
	deepCalled := make(chan struct{}, 1)
	shallowCalled := make(chan struct{}, 1)

	// Register handlers for nested paths
	deepHandler := TrackHandlerFunc(func(p *Publisher) {
		deepCalled <- struct{}{}
	})
	mux.Handle(ctx, BroadcastPath("/deep/nested/path"), deepHandler)

	shallowHandler := TrackHandlerFunc(func(p *Publisher) {
		shallowCalled <- struct{}{}
	})
	mux.Handle(ctx, BroadcastPath("/deep"), shallowHandler)
	// Test deep path
	mockWriter1 := &MockTrackWriter{}
	publisher1 := &Publisher{
		BroadcastPath:   BroadcastPath("/deep/nested/path"),
		TrackName:       "track1",
		TrackWriter:     mockWriter1,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
	}
	mux.ServeTrack(publisher1)

	// Check that deep handler was called but shallow wasn't
	select {
	case <-deepCalled:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Deep handler should have been called")
	}

	select {
	case <-shallowCalled:
		t.Error("Shallow handler should not have been called")
	case <-time.After(10 * time.Millisecond):
		// Expected
	}
	// Test shallow path
	mockWriter2 := &MockTrackWriter{}
	publisher2 := &Publisher{
		BroadcastPath:   BroadcastPath("/deep"),
		TrackName:       "track1",
		TrackWriter:     mockWriter2,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
	}
	mux.ServeTrack(publisher2)

	// Check that shallow handler was called
	select {
	case <-shallowCalled:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Shallow handler should have been called")
	}
}

func TestMux_ContextTimeout(t *testing.T) {
	mux := NewTrackMux()

	handlerCalled := make(chan struct{}, 1)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	handler := TrackHandlerFunc(func(p *Publisher) {
		handlerCalled <- struct{}{}
	})
	mux.Handle(ctx, BroadcastPath("/timeout"), handler)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)
	// Try to serve track - should use NotFoundHandler
	mockWriter := &MockTrackWriter{}
	publisher := &Publisher{
		BroadcastPath:   BroadcastPath("/timeout"),
		TrackName:       "track1",
		TrackWriter:     mockWriter,
		SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
	}

	mux.ServeTrack(publisher)

	// Handler should not have been called due to timeout
	select {
	case <-handlerCalled:
		t.Error("Handler should not have been called after timeout")
	case <-time.After(10 * time.Millisecond):
		// Expected - handler should not be called
	}
}

func TestMux_MultipleCancellations(t *testing.T) {
	mux := NewTrackMux()

	// Create multiple contexts and handlers
	var contexts []context.CancelFunc
	var handlerChannels []chan struct{}

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		contexts = append(contexts, cancel)

		handlerCalled := make(chan struct{}, 1)
		handlerChannels = append(handlerChannels, handlerCalled)

		handler := TrackHandlerFunc(func(p *Publisher) {
			handlerCalled <- struct{}{}
		})

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
	for i, handlerCh := range handlerChannels {
		mockWriter := &MockTrackWriter{}
		publisher := &Publisher{
			BroadcastPath:   BroadcastPath(fmt.Sprintf("/multi%d", i)),
			TrackName:       "track1",
			TrackWriter:     mockWriter,
			SubscribeStream: NewMockReceiveSubscribeStream(SubscribeID(1)),
		}

		mux.ServeTrack(publisher)

		// Check that handler was not called
		select {
		case <-handlerCh:
			t.Errorf("Handler %d should not have been called after cancellation", i)
		case <-time.After(10 * time.Millisecond):
			// Expected - handler should not be called
		}
	}
}

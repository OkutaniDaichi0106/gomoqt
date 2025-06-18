package moqt

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMux_Handle(t *testing.T) {
	tests := map[string]struct {
		path        BroadcastPath
		description string
	}{
		"basic_path": {
			path:        BroadcastPath("/test"),
			description: "should register handler for basic path",
		},
		"nested_path": {
			path:        BroadcastPath("/deep/nested/path"),
			description: "should register handler for nested path",
		},
		"root_path": {
			path:        BroadcastPath("/"),
			description: "should register handler for root path",
		},
		"path_with_dots": {
			path:        BroadcastPath("/client.echo"),
			description: "should register handler for path with dots",
		},
		"path_with_empty_segments": {
			path:        BroadcastPath("/test//segment"),
			description: "should register handler for path with empty segments",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			called := false
			handler := TrackHandlerFunc(func(p *Publisher) {
				called = true
				assert.Equal(t, tt.path, p.BroadcastPath, "handler should receive correct path")
			})

			// Register handler with new API
			mux.Handle(ctx, tt.path, handler)

			// Verify handler is registered and callable
			foundHandler := mux.Handler(tt.path)
			assert.NotNil(t, foundHandler, tt.description)

			// Test handler execution
			publisher := createNewTestPublisher(tt.path)
			foundHandler.ServeTrack(publisher)

			assert.True(t, called, "handler should be called when serving track")
		})
	}
}

func TestNewMux_Handle_Overwrite(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	called1, called2 := false, false

	handler1 := TrackHandlerFunc(func(p *Publisher) { called1 = true })
	handler2 := TrackHandlerFunc(func(p *Publisher) { called2 = true })

	// Register first handler
	mux.Handle(ctx, path, handler1)

	// Test first handler
	publisher := createNewTestPublisher(path)
	mux.ServeTrack(publisher)
	assert.True(t, called1, "First handler should be called")
	assert.False(t, called2, "Second handler should not be called yet")

	// Try to overwrite with second handler - should log warning and not overwrite
	called1, called2 = false, false
	mux.Handle(ctx, path, handler2)

	// Test that first handler is still active (overwrite is prevented)
	mux.ServeTrack(publisher)
	assert.True(t, called1, "First handler should still be called after attempted overwrite")
	assert.False(t, called2, "Second handler should not be called due to overwrite prevention")
}

func TestNewMux_Handle_NilContext(t *testing.T) {
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	handler := TrackHandlerFunc(func(p *Publisher) {})

	// Should panic with nil context
	assert.Panics(t, func() {
		mux.Handle(nil, path, handler)
	})

	// Test that handler wasn't registered after panic
	publisher := createNewTestPublisher(path)

	// Should get NotFoundHandler behavior (track writer gets closed)
	mockWriter := publisher.TrackWriter.(*MockTrackWriter)
	mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

	mux.ServeTrack(publisher)

	// Verify NotFoundHandler was called
	mockWriter.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
}

func TestNewMux_Handle_InvalidPath(t *testing.T) {
	tests := []struct {
		name string
		path BroadcastPath
	}{
		{"empty_path", BroadcastPath("")},
		{"no_leading_slash", BroadcastPath("test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()
			handler := TrackHandlerFunc(func(p *Publisher) {})

			// Should panic for invalid paths
			assert.Panics(t, func() {
				mux.Handle(ctx, tt.path, handler)
			}, "should panic for invalid path: %s", tt.path)
		})
	}
}

func TestNewMux_ServeTrack(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()
	path := BroadcastPath("/test")
	trackName := TrackName("track1")

	// Track that handler is called with correct parameters
	calledCh := make(chan *Publisher, 1)
	mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {
		calledCh <- p
	}))

	// Create and serve publisher
	publisher := createNewTestPublisher(path)
	publisher.TrackName = trackName

	mux.ServeTrack(publisher)

	// Verify handler was called with correct publisher
	select {
	case receivedPublisher := <-calledCh:
		assert.Equal(t, path, receivedPublisher.BroadcastPath, "handler should receive correct path")
		assert.Equal(t, trackName, receivedPublisher.TrackName, "handler should receive correct track name")
		assert.NotNil(t, receivedPublisher.TrackWriter, "handler should receive track writer")
		assert.NotNil(t, receivedPublisher.SubscribeStream, "handler should receive subscribe stream")
	case <-time.After(5 * time.Second):
		t.Fatal("Handler should have been called")
	}
}

func TestNewMux_ServeTrack_NotFound(t *testing.T) {
	mux := NewTrackMux()

	// Create a mock track writer
	mockWriter := &MockTrackWriter{}
	// Set up the mock expectation for CloseWithError
	mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

	// Create mock subscribe stream
	mockSubscribeStream := &MockReceiveSubscribeStream{}
	mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
	mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
	mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

	// Create a publisher for non-existent path
	publisher := &Publisher{
		BroadcastPath:   BroadcastPath("/nonexistent"),
		TrackName:       TrackName("track1"),
		TrackWriter:     mockWriter,
		SubscribeStream: mockSubscribeStream,
	}
	// Should use NotFoundHandler which closes the track
	mux.ServeTrack(publisher)

	// Assert that the publisher's TrackWriter was closed with the expected error
	mockWriter.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
	mockWriter.AssertExpectations(t)
}

func TestNewMux_ServeTrack_NilPublisher(t *testing.T) {
	mux := NewTrackMux()

	// Should handle nil publisher gracefully without panic
	assert.NotPanics(t, func() {
		mux.ServeTrack(nil)
	})
}

func TestNewMux_ServeTrack_NilTrackWriter(t *testing.T) {
	mux := NewTrackMux()

	publisher := &Publisher{
		BroadcastPath:   BroadcastPath("/test"),
		TrackName:       TrackName("track1"),
		TrackWriter:     nil, // nil writer
		SubscribeStream: &MockReceiveSubscribeStream{},
	}

	// Should handle nil track writer gracefully without panic
	assert.NotPanics(t, func() {
		mux.ServeTrack(publisher)
	})
}

func TestNewMux_ServeTrack_NilSubscribeStream(t *testing.T) {
	mux := NewTrackMux()

	publisher := &Publisher{
		BroadcastPath:   BroadcastPath("/test"),
		TrackName:       TrackName("track1"),
		TrackWriter:     &MockTrackWriter{},
		SubscribeStream: nil, // nil stream
	}

	// Should handle nil subscribe stream gracefully without panic
	assert.NotPanics(t, func() {
		mux.ServeTrack(publisher)
	})
}

func TestNewMux_ServeTrack_InvalidPath(t *testing.T) {
	mux := NewTrackMux()

	publisher := createNewTestPublisher(BroadcastPath("invalid-path"))

	// Should panic for invalid path
	assert.Panics(t, func() {
		mux.ServeTrack(publisher)
	})
}

func TestNewMux_ServeAnnouncements(t *testing.T) {
	paths := []BroadcastPath{
		BroadcastPath("/room/person1"),
		BroadcastPath("/room/person2"),
		BroadcastPath("/room/person3"),
	}

	mux := NewTrackMux()

	// Register handlers for paths
	for _, path := range paths {
		mux.Handle(context.Background(), path, TrackHandlerFunc(func(p *Publisher) {}))
	}

	// Create mock announcement writer
	announced := make([]*Announcement, 0)
	var mu sync.Mutex
	mockWriter := &MockAnnouncementWriter{}
	mockWriter.On("SendAnnouncement", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		announcement := args.Get(0).(*Announcement)
		mu.Lock()
		announced = append(announced, announcement)
		mu.Unlock()
	})

	// Test serving announcements in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(mockWriter, "/room/")
	}()

	// Give time for ServeAnnouncements to process initial announcements
	time.Sleep(100 * time.Millisecond)

	// Verify that initial announcements were sent
	mu.Lock()
	initialCount := len(announced)
	mu.Unlock()

	assert.Equal(t, 3, initialCount, "Should have received 3 initial announcements")

	// Verify that all paths are announced
	announcedPaths := make(map[string]bool)
	mu.Lock()
	for _, ann := range announced {
		announcedPaths[string(ann.BroadcastPath())] = true
	}
	mu.Unlock()

	for _, path := range paths {
		assert.True(t, announcedPaths[string(path)], "Path %s should be announced", path)
	}

	// Add a new handler and verify it gets announced
	newPath := BroadcastPath("/room/person4")
	mux.Handle(context.Background(), newPath, TrackHandlerFunc(func(p *Publisher) {}))

	// Give time for new announcement to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the new announcement was sent
	mu.Lock()
	finalCount := len(announced)
	mu.Unlock()

	assert.Equal(t, 4, finalCount, "Should have received 4 total announcements after adding new handler")

	// Verify the new path is in the announcements
	mu.Lock()
	found := false
	for _, ann := range announced {
		if ann.BroadcastPath() == newPath {
			found = true
			break
		}
	}
	mu.Unlock()

	assert.True(t, found, "New path %s should be announced", newPath)

	// Verify all mock expectations were met
	mockWriter.AssertExpectations(t)
}

func TestNewMux_ServeAnnouncements_NilWriter(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register a handler
	path := BroadcastPath("/test/path")
	mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))

	// Test nil writer case - should return immediately without panic
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(nil, "/test/")
	}()

	// Should return immediately due to nil writer
	select {
	case <-done:
		// Expected - function should return immediately
	case <-time.After(100 * time.Millisecond):
		t.Error("ServeAnnouncements should have returned immediately with nil writer")
	}
}

func TestNewMux_ServeAnnouncements_InvalidPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"empty_prefix", ""},
		{"no_leading_slash", "test/"},
		{"no_trailing_slash", "/test"},
		{"no_slashes", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := NewTrackMux()
			mockWriter := &MockAnnouncementWriter{}
			mockWriter.On("CloseWithError", InvalidPrefixErrorCode).Return(nil)

			mux.ServeAnnouncements(mockWriter, tt.prefix)

			mockWriter.AssertCalled(t, "CloseWithError", InvalidPrefixErrorCode)
		})
	}
}

func TestNewMux_ServeAnnouncements_EmptyPrefix(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers with different paths
	paths := []BroadcastPath{
		BroadcastPath("/room/a"),
		BroadcastPath("/game/b"),
		BroadcastPath("/chat/c"),
	}

	for _, path := range paths {
		mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))
	}

	// Create mock announcement writer
	announced := make([]*Announcement, 0)
	var mu sync.Mutex
	mockWriter := &MockAnnouncementWriter{}
	mockWriter.On("SendAnnouncement", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		announcement := args.Get(0).(*Announcement)
		mu.Lock()
		announced = append(announced, announcement)
		mu.Unlock()
	})

	// Test serving announcements with root prefix
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(mockWriter, "/")
	}()

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify that all paths are announced (since "/" matches all)
	mu.Lock()
	count := len(announced)
	mu.Unlock()

	assert.Equal(t, 3, count, "Should have received all 3 announcements with root prefix")

	// Verify all expected paths are present
	mu.Lock()
	announcedPaths := make(map[string]bool)
	for _, ann := range announced {
		announcedPaths[string(ann.BroadcastPath())] = true
	}
	mu.Unlock()

	for _, path := range paths {
		assert.True(t, announcedPaths[string(path)], "Path %s should be announced", path)
	}
	// Verify all mock expectations were met
	mockWriter.AssertExpectations(t)
}

func TestNewMux_Handler(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	// Test not found case first - verify behavior not identity
	publisher := createNewTestPublisher(path)
	mockWriter := publisher.TrackWriter.(*MockTrackWriter)
	mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

	mux.ServeTrack(publisher)
	mockWriter.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)

	// Register a handler
	called := false
	expectedHandler := TrackHandlerFunc(func(p *Publisher) { called = true })
	mux.Handle(ctx, path, expectedHandler)

	// Test found case - verify behavior not identity
	publisher2 := createNewTestPublisher(path)
	mux.ServeTrack(publisher2)
	assert.True(t, called, "registered handler should be called")
}

func TestNewMux_Handler_InvalidPath(t *testing.T) {
	mux := NewTrackMux()

	// Should panic for invalid paths
	assert.Panics(t, func() {
		mux.Handler(BroadcastPath(""))
	})

	assert.Panics(t, func() {
		mux.Handler(BroadcastPath("invalid"))
	})
}

func TestNewMux_HandleFunc(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	called := false
	mux.HandleFunc(ctx, path, func(pub *Publisher) {
		called = true
		assert.Equal(t, path, pub.BroadcastPath)
	})

	// Test that the function was registered correctly
	handler := mux.Handler(path)
	assert.NotNil(t, handler, "handler should be registered")

	// Test that the function is called correctly
	publisher := createNewTestPublisher(path)
	handler.ServeTrack(publisher)
	assert.True(t, called, "function should be called")
}

func TestNewMux_Announce_Direct(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement := NewAnnouncement(ctx, path)
	called := false
	handler := TrackHandlerFunc(func(p *Publisher) { called = true })

	// Test direct announce
	mux.Announce(announcement, handler)

	// Verify handler is registered by testing behavior
	publisher := createNewTestPublisher(path)
	mux.ServeTrack(publisher)
	assert.True(t, called, "registered handler should be called")
}

func TestNewMux_Announce_InactiveAnnouncement(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement := NewAnnouncement(ctx, path)
	announcement.End() // Make it inactive

	handler := TrackHandlerFunc(func(p *Publisher) {})

	// Test announce with inactive announcement
	mux.Announce(announcement, handler)

	// Handler should not be registered - test by verifying NotFoundHandler behavior
	publisher := createNewTestPublisher(path)
	mockWriter := publisher.TrackWriter.(*MockTrackWriter)
	mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

	mux.ServeTrack(publisher)
	mockWriter.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
}

func TestNewMux_Announce_DuplicatePath(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement1 := NewAnnouncement(ctx, path)
	announcement2 := NewAnnouncement(ctx, path)

	handler1 := TrackHandlerFunc(func(p *Publisher) {})
	handler2 := TrackHandlerFunc(func(p *Publisher) {})

	// Register first handler
	mux.Announce(announcement1, handler1)

	// Try to register second handler - should log warning and not overwrite
	mux.Announce(announcement2, handler2)

	// First handler should still be active
	foundHandler := mux.Handler(path)
	assert.NotNil(t, foundHandler, "handler should still be registered")
}

func TestNewMux_Clear(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register some handlers
	paths := []BroadcastPath{
		BroadcastPath("/test1"),
		BroadcastPath("/test2"),
		BroadcastPath("/nested/test3"),
	}

	callCounts := make(map[BroadcastPath]bool)
	for _, path := range paths {
		path := path // capture loop variable
		mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {
			callCounts[path] = true
		}))
	}

	// Verify handlers are registered by testing behavior
	for _, path := range paths {
		publisher := createNewTestPublisher(path)
		mux.ServeTrack(publisher)
		assert.True(t, callCounts[path], "handler should be registered for path %s", path)
	}

	// Clear the mux
	mux.Clear()

	// Verify all handlers are removed by testing NotFoundHandler behavior
	for _, path := range paths {
		publisher := createNewTestPublisher(path)
		mockWriter := publisher.TrackWriter.(*MockTrackWriter)
		mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

		mux.ServeTrack(publisher)
		mockWriter.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
	}
}

func TestNewMux_AnnouncementLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	called := false
	handler := TrackHandlerFunc(func(p *Publisher) { called = true })

	// Register handler
	mux.Handle(ctx, path, handler)

	// Verify handler is registered by testing behavior
	publisher := createNewTestPublisher(path)
	mux.ServeTrack(publisher)
	assert.True(t, called, "handler should be registered")

	// Cancel context to end announcement
	cancel()

	// Give time for cleanup
	time.Sleep(50 * time.Millisecond)

	// Handler should be removed - test by verifying NotFoundHandler behavior
	publisher2 := createNewTestPublisher(path)
	mockWriter := publisher2.TrackWriter.(*MockTrackWriter)
	mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

	mux.ServeTrack(publisher2)
	mockWriter.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
}

func TestNewMux_ConcurrentAccess(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	const numGoroutines = 10
	const numPaths = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Test concurrent registration and access
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numPaths; j++ {
				path := BroadcastPath(fmt.Sprintf("/test/%d/%d", id, j))
				handler := TrackHandlerFunc(func(p *Publisher) {})

				// Register handler
				mux.Handle(ctx, path, handler)

				// Access handler
				foundHandler := mux.Handler(path)
				assert.NotNil(t, foundHandler, "handler should be found")

				// Serve track
				publisher := createNewTestPublisher(path)
				mux.ServeTrack(publisher)
			}
		}(i)
	}

	wg.Wait()
}

// Helper function for creating test publishers
func createNewTestPublisher(path BroadcastPath) *Publisher {
	mockWriter := &MockTrackWriter{}
	mockSubscribeStream := &MockReceiveSubscribeStream{}

	// Only set up the expectations that are commonly needed
	mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
	mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
	mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

	return &Publisher{
		BroadcastPath:   path,
		TrackName:       TrackName("test"),
		TrackWriter:     mockWriter,
		SubscribeStream: mockSubscribeStream,
	}
}

// Test DefaultMux functionality
func TestNewMux_DefaultMux(t *testing.T) {
	// Clear DefaultMux first
	DefaultMux.Clear()

	ctx := context.Background()
	path := BroadcastPath("/default/test")

	called := false
	handler := TrackHandlerFunc(func(p *Publisher) {
		called = true
	})

	// Test top-level Handle function
	Handle(ctx, path, handler)

	// Test top-level HandleFunc function
	path2 := BroadcastPath("/default/test2")
	called2 := false
	HandleFunc(ctx, path2, func(pub *Publisher) {
		called2 = true
	})

	// Test handlers work
	publisher := createNewTestPublisher(path)
	DefaultMux.ServeTrack(publisher)
	assert.True(t, called, "handler should be called")

	publisher2 := createNewTestPublisher(path2)
	DefaultMux.ServeTrack(publisher2)
	assert.True(t, called2, "handler2 should be called")

	// Test direct Announce function
	path3 := BroadcastPath("/default/test3")
	announcement := NewAnnouncement(ctx, path3)
	called3 := false
	handler3 := TrackHandlerFunc(func(p *Publisher) { called3 = true })

	Announce(announcement, handler3)

	publisher3 := createNewTestPublisher(path3)
	DefaultMux.ServeTrack(publisher3)
	assert.True(t, called3, "handler3 should be called")

	// Clean up
	DefaultMux.Clear()
}

func TestNewMux_ValidationFunctions(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{"valid_basic", "/test", true},
		{"valid_nested", "/deep/nested/path", true},
		{"valid_root", "/", true},
		{"valid_with_dots", "/client.echo", true},
		{"valid_with_empty_segments", "/test//segment", true},
		{"invalid_empty", "", false},
		{"invalid_no_slash", "test", false},
		{"invalid_middle_only", "test/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPath(BroadcastPath(tt.path))
			assert.Equal(t, tt.valid, result, "isValidPath(%s) should return %t", tt.path, tt.valid)
		})
	}
}

func TestNewMux_PrefixValidation(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		valid  bool
	}{
		{"valid_basic", "/test/", true},
		{"valid_nested", "/deep/nested/", true},
		{"valid_root", "/", true},
		{"invalid_empty", "", false},
		{"invalid_no_leading_slash", "test/", false},
		{"invalid_no_trailing_slash", "/test", false},
		{"invalid_no_slashes", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPrefix(tt.prefix)
			assert.Equal(t, tt.valid, result, "isValidPrefix(%s) should return %t", tt.prefix, tt.valid)
		})
	}
}

func TestNewMux_ServeAnnouncements_PrefixFiltering_Complete(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()
	// Register handlers with specific prefix pattern
	matchingPaths := []BroadcastPath{
		BroadcastPath("/room/alice"),
		BroadcastPath("/room/bob"),
	}
	nonMatchingPaths := []BroadcastPath{
		BroadcastPath("/game/chess"),
		BroadcastPath("/chat/general"),
	}

	// Register all handlers
	allPaths := append(matchingPaths, nonMatchingPaths...)
	for _, path := range allPaths {
		mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))
	}

	// Create mock announcement writer
	announced := make([]*Announcement, 0)
	var mu sync.Mutex
	mockWriter := &MockAnnouncementWriter{}
	mockWriter.On("SendAnnouncement", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		announcement := args.Get(0).(*Announcement)
		mu.Lock()
		announced = append(announced, announcement)
		mu.Unlock()
	})

	// Test with /room/ prefix - should only announce matching paths
	mux.ServeAnnouncements(mockWriter, "/room/")

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	// Collect announced paths
	mu.Lock()
	announcedPaths := make([]BroadcastPath, 0, len(announced))
	for _, ann := range announced {
		announcedPaths = append(announcedPaths, ann.BroadcastPath())
	}
	mu.Unlock()

	// Verify only matching paths are announced
	assert.Equal(t, len(matchingPaths), len(announcedPaths),
		"Should announce only paths matching /room/ prefix")

	// Verify specific paths
	announcedPathsMap := make(map[BroadcastPath]bool)
	for _, path := range announcedPaths {
		announcedPathsMap[path] = true
	}

	for _, expectedPath := range matchingPaths {
		assert.True(t, announcedPathsMap[expectedPath],
			"Expected path %s should be announced", expectedPath)
	}

	// Verify non-matching paths are not announced
	for _, nonMatchingPath := range nonMatchingPaths {
		assert.False(t, announcedPathsMap[nonMatchingPath],
			"Non-matching path %s should not be announced", nonMatchingPath)
	}

	mockWriter.AssertExpectations(t)
}

func TestNewMux_ServeAnnouncements_RootPrefixMatching(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers with different prefixes
	allPaths := []BroadcastPath{
		BroadcastPath("/room/alice"),
		BroadcastPath("/game/chess"),
		BroadcastPath("/chat/general"),
	}

	for _, path := range allPaths {
		mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))
	}

	// Create mock announcement writer
	announced := make([]*Announcement, 0)
	var mu sync.Mutex
	mockWriter := &MockAnnouncementWriter{}
	mockWriter.On("SendAnnouncement", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		announcement := args.Get(0).(*Announcement)
		mu.Lock()
		announced = append(announced, announcement)
		mu.Unlock()
	})

	// Test with root prefix - should announce all paths
	mux.ServeAnnouncements(mockWriter, "/")

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	// Collect announced paths
	mu.Lock()
	announcedPaths := make([]BroadcastPath, 0, len(announced))
	for _, ann := range announced {
		announcedPaths = append(announcedPaths, ann.BroadcastPath())
	}
	mu.Unlock()

	// Verify all paths are announced with root prefix
	assert.Equal(t, len(allPaths), len(announcedPaths),
		"Root prefix should announce all registered paths")

	// Verify all expected paths are present
	announcedPathsMap := make(map[BroadcastPath]bool)
	for _, path := range announcedPaths {
		announcedPathsMap[path] = true
	}

	for _, expectedPath := range allPaths {
		assert.True(t, announcedPathsMap[expectedPath],
			"Path %s should be announced with root prefix", expectedPath)
	}

	mockWriter.AssertExpectations(t)
}

func TestNewMux_ServeAnnouncements_NonMatchingPrefix(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers that don't match the test prefix
	paths := []BroadcastPath{
		BroadcastPath("/room/alice"),
		BroadcastPath("/chat/general"),
	}

	for _, path := range paths {
		mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))
	}

	// Create mock announcement writer
	announced := make([]*Announcement, 0)
	var mu sync.Mutex
	mockWriter := &MockAnnouncementWriter{}
	mockWriter.On("SendAnnouncement", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		announcement := args.Get(0).(*Announcement)
		mu.Lock()
		announced = append(announced, announcement)
		mu.Unlock()
	})

	// Test with non-matching prefix
	mux.ServeAnnouncements(mockWriter, "/game/")

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	// Verify no announcements are sent
	mu.Lock()
	announcedCount := len(announced)
	mu.Unlock()

	assert.Equal(t, 0, announcedCount,
		"Should not announce any paths when prefix doesn't match")

	// Note: We don't call AssertExpectations here because no calls are expected
}

func TestNewMux_ServeAnnouncements_BroadcastServerIssue(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handler for "/index" like broadcast server does
	mux.Handle(ctx, BroadcastPath("/index"), TrackHandlerFunc(func(p *Publisher) {}))

	// Create mock announcement writer to capture sent announcements
	announced := make([]*Announcement, 0)
	var mu sync.Mutex
	mockWriter := &MockAnnouncementWriter{}
	mockWriter.On("SendAnnouncement", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		announcement := args.Get(0).(*Announcement)
		mu.Lock()
		announced = append(announced, announcement)
		mu.Unlock()
	})

	// Test with "/" prefix like client opens announce stream
	mux.ServeAnnouncements(mockWriter, "/")

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify that the "/index" announcement was sent
	mu.Lock()
	count := len(announced)
	mu.Unlock()

	assert.Equal(t, 1, count, "Should have sent 1 announcement for /index")
	if count > 0 {
		mu.Lock()
		ann := announced[0]
		mu.Unlock()
		assert.Equal(t, BroadcastPath("/index"), ann.BroadcastPath(), "Should announce /index path")
	}

	// Verify all mock expectations were met
	mockWriter.AssertExpectations(t)
}

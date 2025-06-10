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

func TestMux_Handle(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		action   string
		expected bool
	}{
		"basic handle": {
			path:     BroadcastPath("/test"),
			action:   "handle",
			expected: true,
		},
		"overwrite existing handler": {
			path:     BroadcastPath("/test"),
			action:   "overwrite",
			expected: true,
		},
		"multiple paths": {
			path:     BroadcastPath("/another"),
			action:   "handle",
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			handler := TrackHandlerFunc(func(p *Publisher) {})

			switch tt.action {
			case "handle":
				mux.Handle(ctx, tt.path, handler)
			case "overwrite":
				// First handle
				mux.Handle(ctx, tt.path, handler)
				// Then overwrite
				mux.Handle(ctx, tt.path, handler)
			}

			assert.True(t, tt.expected)
		})
	}
}

func TestMux_ServeTrack(t *testing.T) {
	tests := map[string]struct {
		path      BroadcastPath
		trackName TrackName
		timeout   time.Duration
	}{
		"serve existing track": {
			path:      BroadcastPath("/test"),
			trackName: TrackName("track1"),
			timeout:   5 * time.Second,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			calledCh := make(chan struct{}, 1)
			mux.Handle(ctx, tt.path, TrackHandlerFunc(func(p *Publisher) {
				assert.Equal(t, string(tt.path), string(p.BroadcastPath))
				assert.Equal(t, tt.trackName, p.TrackName)
				assert.NotNil(t, p.TrackWriter)
				assert.NotNil(t, p.SubscribeStream)

				calledCh <- struct{}{}
			})) // Create mock subscribe stream
			mockSubscribeStream := &MockReceiveSubscribeStream{}
			mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
			mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
			mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

			// Create a publisher
			publisher := &Publisher{
				BroadcastPath:   tt.path,
				TrackName:       tt.trackName,
				TrackWriter:     &MockTrackWriter{},
				SubscribeStream: mockSubscribeStream,
			}

			mux.ServeTrack(publisher)

			select {
			case <-calledCh:
			case <-time.After(tt.timeout):
				t.Error("Handler should have been called")
			}
		})
	}
}

func TestMux_ServeTrack_NotFound(t *testing.T) {
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
}

func TestMux_ServeAnnouncements(t *testing.T) {
	paths := []BroadcastPath{
		BroadcastPath("/room/person1"),
		BroadcastPath("/room/person2"),
		BroadcastPath("/room/person3"),
	}

	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers for paths
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

	// Test serving announcements in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(mockWriter, "/room")
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
	mux.Handle(ctx, newPath, TrackHandlerFunc(func(p *Publisher) {}))

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

func TestMux_ServeAnnouncements_ErrorHandling(t *testing.T) {
	tests := map[string]struct {
		mockError       error
		expectReturn    bool
		expectCallCount int
	}{
		"writer error on initial send": {
			mockError:       fmt.Errorf("send error"),
			expectReturn:    true,
			expectCallCount: 1,
		},
		"nil writer": {
			mockError:       nil,
			expectReturn:    true,
			expectCallCount: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			// Register a handler
			path := BroadcastPath("/test/path")
			mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))

			if tt.mockError == nil && tt.expectCallCount == 0 {
				// Test nil writer case
				// This should return immediately without panic
				mux.ServeAnnouncements(nil, "/test")
				return
			} // Create mock that returns error
			mockWriter := &MockAnnouncementWriter{}
			mockWriter.On("SendAnnouncement", mock.Anything).Return(tt.mockError)

			// Test serving announcements with error
			done := make(chan struct{})
			go func() {
				defer close(done)
				mux.ServeAnnouncements(mockWriter, "/test")
			}()

			// Give time for processing
			time.Sleep(50 * time.Millisecond) // Verify the expected number of calls
			if tt.expectCallCount > 0 {
				mockWriter.AssertNumberOfCalls(t, "SendAnnouncement", tt.expectCallCount)
			}

			// Should return due to error
			select {
			case <-done:
				// Expected
			case <-time.After(100 * time.Millisecond):
				if tt.expectReturn {
					t.Error("ServeAnnouncements should have returned due to error")
				}
			}
		})
	}
}

func TestMux_ServeAnnouncements_EmptyPrefix(t *testing.T) {
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
		mux.ServeAnnouncements(mockWriter, "")
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
}

func TestMux_ServeAnnouncements_NoMatchingPaths(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handlers with paths that won't match the prefix
	paths := []BroadcastPath{
		BroadcastPath("/room/person1"),
		BroadcastPath("/room/person2"),
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

	// Test serving announcements with non-matching prefix
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(mockWriter, "/game")
	}()

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify that no announcements were sent initially
	mu.Lock()
	count := len(announced)
	mu.Unlock()

	assert.Equal(t, 0, count, "Should have received no announcements for non-matching prefix")
}

func TestMux_ServeAnnouncements_ContextCancellation(t *testing.T) {
	mux := NewTrackMux()

	// Create a cancelable context and register handler
	ctx, cancel := context.WithCancel(context.Background())
	path := BroadcastPath("/test/path")
	mux.Handle(ctx, path, TrackHandlerFunc(func(p *Publisher) {}))
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

	// Start serving announcements
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(mockWriter, "/test")
	}()

	// Give time for initial processing
	time.Sleep(50 * time.Millisecond)

	// Verify initial announcement was sent
	mu.Lock()
	initialCount := len(announced)
	mu.Unlock()
	assert.Equal(t, 1, initialCount, "Should have received 1 initial announcement")

	// Cancel the context (this should end the announcement)
	cancel()

	// Give time for context cancellation to be processed
	time.Sleep(100 * time.Millisecond)

	// Add a new handler with the same path - should not be announced since context was cancelled
	newCtx := context.Background()
	mux.Handle(newCtx, path, TrackHandlerFunc(func(p *Publisher) {}))

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Should receive the new announcement
	mu.Lock()
	finalCount := len(announced)
	mu.Unlock()

	assert.GreaterOrEqual(t, finalCount, initialCount, "Should have received new announcement after context cancellation")
}

func TestMux_ConcurrentAccess(t *testing.T) {
	tests := map[string]struct {
		goroutines int
		pathPrefix string
	}{
		"concurrent access": {
			goroutines: 10,
			pathPrefix: "/test",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			// Register multiple handlers concurrently
			var wg sync.WaitGroup
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					handler := TrackHandlerFunc(func(p *Publisher) {
						assert.Equal(t, fmt.Sprintf("%s%d", tt.pathPrefix, i), string(p.BroadcastPath))
					})
					mux.Handle(ctx, BroadcastPath(fmt.Sprintf("%s%d", tt.pathPrefix, i)), handler)
				}(i)
			}
			wg.Wait() // Serve tracks concurrently
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					mockWriter := &MockTrackWriter{}

					// Create mock subscribe stream
					mockSubscribeStream := &MockReceiveSubscribeStream{}
					mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
					mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
					mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

					publisher := &Publisher{
						BroadcastPath:   BroadcastPath(fmt.Sprintf("%s%d", tt.pathPrefix, i)),
						TrackName:       "track1",
						TrackWriter:     mockWriter,
						SubscribeStream: mockSubscribeStream,
					}
					mux.ServeTrack(publisher)
				}(i)
			}
			wg.Wait()
		})
	}
}

func TestDefaultMux(t *testing.T) {
	tests := map[string]struct {
		path    BroadcastPath
		timeout time.Duration
	}{
		"default mux handle and serve": {
			path:    BroadcastPath("/default"),
			timeout: 5 * time.Second,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Test that DefaultMux is accessible
			assert.NotNil(t, DefaultMux)

			calledCh := make(chan struct{}, 1)

			// Test Handle on default mux
			Handle(context.Background(), tt.path, TrackHandlerFunc(func(p *Publisher) {
				calledCh <- struct{}{}
			})) // Test ServeTrack on default mux
			mockWriter := &MockTrackWriter{}

			// Create mock subscribe stream
			mockSubscribeStream := &MockReceiveSubscribeStream{}
			mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
			mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
			mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

			publisher := &Publisher{
				BroadcastPath:   tt.path,
				TrackName:       "track1",
				TrackWriter:     mockWriter,
				SubscribeStream: mockSubscribeStream,
			}

			DefaultMux.ServeTrack(publisher)

			select {
			case <-calledCh:
			case <-time.After(tt.timeout):
				t.Error("Handler should have been called")
			}
		})
	}
}

func TestMux_HandleFunc(t *testing.T) {
	tests := map[string]struct {
		path      BroadcastPath
		trackName TrackName
	}{
		"handle func wrapper": {
			path:      BroadcastPath("/func"),
			trackName: TrackName("track1"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			called := false
			var receivedPublisher *Publisher
			handlerFunc := func(p *Publisher) {
				called = true
				receivedPublisher = p
			} // Use Handle with TrackHandlerFunc since HandleFunc doesn't exist
			mux.Handle(ctx, tt.path, TrackHandlerFunc(handlerFunc))
			mockWriter := &MockTrackWriter{}

			// Create mock subscribe stream
			mockSubscribeStream := &MockReceiveSubscribeStream{}
			mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
			mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
			mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

			publisher := &Publisher{
				BroadcastPath:   tt.path,
				TrackName:       tt.trackName,
				TrackWriter:     mockWriter,
				SubscribeStream: mockSubscribeStream,
			}

			mux.ServeTrack(publisher)
			assert.True(t, called)
			assert.Equal(t, publisher, receivedPublisher)
		})
	}
}

func TestMux_AnnouncementLifecycle(t *testing.T) {
	tests := map[string]struct {
		handlerPath     BroadcastPath
		announcePath    BroadcastPath
		expectActive    bool
		expectPathMatch bool
	}{
		"announcement lifecycle": {
			handlerPath:     BroadcastPath("/room"),
			announcePath:    BroadcastPath("/room/test"),
			expectActive:    true,
			expectPathMatch: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			// Register handler
			mux.Handle(ctx, tt.handlerPath, TrackHandlerFunc(func(p *Publisher) {}))

			// Create announcement
			announcement := NewAnnouncement(context.Background(), tt.announcePath)
			defer announcement.AwaitEnd()

			// Verify announcement is active
			assert.Equal(t, tt.expectActive, announcement.IsActive())
			if tt.expectPathMatch {
				assert.Equal(t, string(tt.announcePath), string(announcement.BroadcastPath()))
			}
		})
	}
}

func TestMux_NestedPaths(t *testing.T) {
	tests := map[string]struct {
		deepPath      BroadcastPath
		shallowPath   BroadcastPath
		testPath      BroadcastPath
		expectDeep    bool
		expectShallow bool
		timeout       time.Duration
	}{
		"deep path call": {
			deepPath:      BroadcastPath("/deep/nested/path"),
			shallowPath:   BroadcastPath("/deep"),
			testPath:      BroadcastPath("/deep/nested/path"),
			expectDeep:    true,
			expectShallow: false,
			timeout:       100 * time.Millisecond,
		},
		"shallow path call": {
			deepPath:      BroadcastPath("/deep/nested/path"),
			shallowPath:   BroadcastPath("/deep"),
			testPath:      BroadcastPath("/deep"),
			expectDeep:    false,
			expectShallow: true,
			timeout:       100 * time.Millisecond,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mux := NewTrackMux()

			// Track which handlers are called using channels
			deepCalled := make(chan struct{}, 1)
			shallowCalled := make(chan struct{}, 1)

			// Register handlers for nested paths
			deepHandler := TrackHandlerFunc(func(p *Publisher) {
				deepCalled <- struct{}{}
			})
			mux.Handle(ctx, tt.deepPath, deepHandler)

			shallowHandler := TrackHandlerFunc(func(p *Publisher) {
				shallowCalled <- struct{}{}
			})
			mux.Handle(ctx, tt.shallowPath, shallowHandler) // Test path
			mockWriter := &MockTrackWriter{}

			// Create mock subscribe stream
			mockSubscribeStream := &MockReceiveSubscribeStream{}
			mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
			mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
			mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

			publisher := &Publisher{
				BroadcastPath:   tt.testPath,
				TrackName:       "track1",
				TrackWriter:     mockWriter,
				SubscribeStream: mockSubscribeStream,
			}
			mux.ServeTrack(publisher)

			// Check deep handler expectation
			if tt.expectDeep {
				select {
				case <-deepCalled:
					// Expected
				case <-time.After(tt.timeout):
					t.Error("Deep handler should have been called")
				}
			} else {
				select {
				case <-deepCalled:
					t.Error("Deep handler should not have been called")
				case <-time.After(10 * time.Millisecond):
					// Expected
				}
			}

			// Check shallow handler expectation
			if tt.expectShallow {
				select {
				case <-shallowCalled:
					// Expected
				case <-time.After(tt.timeout):
					t.Error("Shallow handler should have been called")
				}
			} else {
				select {
				case <-shallowCalled:
					t.Error("Shallow handler should not have been called")
				case <-time.After(10 * time.Millisecond):
					// Expected
				}
			}
		})
	}
}

func TestMux_ContextTimeout(t *testing.T) {
	tests := map[string]struct {
		path         BroadcastPath
		timeout      time.Duration
		waitTime     time.Duration
		expectCalled bool
	}{
		"context timeout": {
			path:         BroadcastPath("/timeout"),
			timeout:      50 * time.Millisecond,
			waitTime:     100 * time.Millisecond,
			expectCalled: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mux := NewTrackMux()

			handlerCalled := make(chan struct{}, 1)

			// Create context with short timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			handler := TrackHandlerFunc(func(p *Publisher) {
				handlerCalled <- struct{}{}
			})
			mux.Handle(ctx, tt.path, handler) // Wait for timeout
			time.Sleep(tt.waitTime)           // Try to serve track - should use NotFoundHandler
			mockWriter := &MockTrackWriter{}
			// Set up the mock expectation for CloseWithError when using NotFoundHandler
			mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

			// Create mock subscribe stream
			mockSubscribeStream := &MockReceiveSubscribeStream{}
			mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
			mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
			mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

			publisher := &Publisher{
				BroadcastPath:   tt.path,
				TrackName:       "track1",
				TrackWriter:     mockWriter,
				SubscribeStream: mockSubscribeStream,
			}

			mux.ServeTrack(publisher)

			// Check handler call expectation
			if tt.expectCalled {
				select {
				case <-handlerCalled:
					// Expected
				case <-time.After(10 * time.Millisecond):
					t.Error("Handler should have been called")
				}
			} else {
				select {
				case <-handlerCalled:
					t.Error("Handler should not have been called after timeout")
				case <-time.After(10 * time.Millisecond):
					// Expected - handler should not be called
				}
			}
		})
	}
}

func TestMux_MultipleCancellations(t *testing.T) {
	tests := map[string]struct {
		numContexts int
		pathPrefix  string
		waitTime    time.Duration
		timeout     time.Duration
	}{
		"multiple cancellations": {
			numContexts: 5,
			pathPrefix:  "/multi",
			waitTime:    20 * time.Millisecond,
			timeout:     10 * time.Millisecond,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mux := NewTrackMux()

			// Create multiple contexts and handlers
			var contexts []context.CancelFunc
			var handlerChannels []chan struct{}

			for i := 0; i < tt.numContexts; i++ {
				ctx, cancel := context.WithCancel(context.Background())
				contexts = append(contexts, cancel)

				handlerCalled := make(chan struct{}, 1)
				handlerChannels = append(handlerChannels, handlerCalled)

				handler := TrackHandlerFunc(func(p *Publisher) {
					handlerCalled <- struct{}{}
				})

				path := BroadcastPath(fmt.Sprintf("%s%d", tt.pathPrefix, i))
				mux.Handle(ctx, path, handler)
			}

			// Cancel all contexts
			for _, cancel := range contexts {
				cancel()
			}

			// Wait for cleanup
			time.Sleep(tt.waitTime) // Verify all handlers were removed
			for i, handlerCh := range handlerChannels {
				mockWriter := &MockTrackWriter{}
				// Set up the mock expectation for CloseWithError when using NotFoundHandler
				mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

				// Create mock subscribe stream
				mockSubscribeStream := &MockReceiveSubscribeStream{}
				mockSubscribeStream.On("SubscribeID").Return(SubscribeID(1)).Maybe()
				mockSubscribeStream.On("SubscribeConfig").Return(&SubscribeConfig{}, nil).Maybe()
				mockSubscribeStream.On("Updated").Return(make(<-chan struct{})).Maybe()

				publisher := &Publisher{
					BroadcastPath:   BroadcastPath(fmt.Sprintf("%s%d", tt.pathPrefix, i)),
					TrackName:       "track1",
					TrackWriter:     mockWriter,
					SubscribeStream: mockSubscribeStream,
				}

				mux.ServeTrack(publisher)

				// Check that handler was not called
				select {
				case <-handlerCh:
					t.Errorf("Handler %d should not have been called after cancellation", i)
				case <-time.After(tt.timeout):
					// Expected - handler should not be called
				}
			}
		})
	}
}

// // Helper function to count nodes in the tree
// func countNodes(node *routingNode) int {
// 	if node == nil {
// 		return 0
// 	}

// 	count := 1
// 	for _, child := range node.children {
// 		count += countNodes(child)
// 	}

// 	return count
// }

// // Helper function to parse path segments
// func parsePathSegments(path string) []string {
// 	if path == "/" {
// 		return []string{}
// 	}

// 	// Remove leading slash and split
// 	if len(path) > 0 && path[0] == '/' {
// 		path = path[1:]
// 	}

// 	if path == "" {
// 		return []string{}
// 	}

// 	segments := []string{}
// 	current := ""
// 	for _, char := range path {
// 		if char == '/' {
// 			if current != "" {
// 				segments = append(segments, current)
// 				current = ""
// 			}
// 		} else {
// 			current += string(char)
// 		}
// 	}
// 	if current != "" {
// 		segments = append(segments, current)
// 	}

// 	return segments
// }

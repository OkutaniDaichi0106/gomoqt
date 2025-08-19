package moqt

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test NewTrackMux function
func TestNewTrackMux(t *testing.T) {
	mux := NewTrackMux()
	assert.NotNil(t, mux, "NewTrackMux should return non-nil mux")
	assert.NotNil(t, &mux.trackHandlerIndex, "mux trackHandlerIndex should be initialized")
	assert.NotNil(t, &mux.announcementTree, "mux announcementTree should be initialized")
	assert.Equal(t, 0, len(mux.trackHandlerIndex), "trackHandlerIndex should be empty initially")
}

// Test Mux.Publish method
func TestMux_Publish(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	called := false
	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {
		called = true
	})

	// Register handler
	mux.Publish(ctx, path, handler)

	// Verify handler is registered by calling ServeTrack
	tw := &TrackWriter{BroadcastPath: path}
	mux.serveTrack(tw)

	assert.True(t, called, "handler should be called")
}

// Test Mux.PublishFunc method
func TestMux_PublishFunc(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	called := false
	mux.PublishFunc(ctx, path, func(ctx context.Context, tw *TrackWriter) {
		called = true
		assert.Equal(t, path, tw.BroadcastPath)
	})

	// Verify handler is registered and called
	tw := &TrackWriter{BroadcastPath: path}
	mux.serveTrack(tw)

	assert.True(t, called, "handler function should be called")
}

// Test path validation
func TestMux_Publish_InvalidPath(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {})

	tests := []struct {
		name string
		path BroadcastPath
	}{
		{"empty_path", BroadcastPath("")},
		{"no_leading_slash", BroadcastPath("test")},
		{"relative_path", BroadcastPath("./test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				mux.Publish(ctx, tt.path, handler)
			}, "should panic for invalid path: %s", tt.path)
		})
	}
}

// Test ServeTrack with registered handler
func TestMux_ServeTrack(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test/path")

	receivedTW := make(chan *TrackWriter, 1)
	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {
		receivedTW <- tw
	})

	mux.Publish(ctx, path, handler)

	// Create track writer and serve
	tw := &TrackWriter{
		BroadcastPath: path,
		TrackName:     TrackName("test_track"),
	}
	mux.serveTrack(tw)

	// Verify handler was called with correct track writer
	select {
	case received := <-receivedTW:
		assert.Equal(t, tw, received, "handler should receive the same track writer")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("handler should have been called")
	}
}

// Test ServeTrack with NotFoundHandler (no registered handler)
func TestMux_ServeTrack_NotFound(t *testing.T) {
	mux := NewTrackMux()

	tw := &TrackWriter{
		BroadcastPath: BroadcastPath("/nonexistent"),
		TrackName:     TrackName("test_track"),
	}

	// Should use NotFoundHandler - this shouldn't panic
	assert.NotPanics(t, func() {
		mux.serveTrack(tw)
	}, "ServeTrack with unregistered path should not panic")
}

// Test ServeTrack with nil TrackWriter
func TestMux_ServeTrack_NilTrackWriter(t *testing.T) {
	mux := NewTrackMux()

	assert.NotPanics(t, func() {
		mux.serveTrack(nil)
	}, "ServeTrack with nil should not panic")
}

// Test Publishr method
func TestMux_Publishr(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	// Initially should return NotFoundHandler
	a, handler := mux.TrackHandler(path)
	assert.Nil(t, a, "Announcement should be nil for unregistered path")
	// Compare handler functions by reflect.Value since func == func is invalid
	assert.Equal(t, reflect.ValueOf(NotFoundTrackHandler), reflect.ValueOf(handler), "Handler should be NotFoundTrackHandler for unregistered path")

	// Register handler
	expectedHandler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {})
	mux.Publish(ctx, path, expectedHandler)

	// Should return registered handler
	a, handler = mux.TrackHandler(path)
	assert.NotNil(t, a, "Announcement should not be nil for registered path")
	assert.Equal(t, reflect.ValueOf(expectedHandler), reflect.ValueOf(handler), "Handler should be the registered handler")
}

// Test nested paths
func TestMux_NestedPaths(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	paths := []BroadcastPath{
		BroadcastPath("/"),
		BroadcastPath("/api"),
		BroadcastPath("/api/v1"),
		BroadcastPath("/api/v1/users"),
		BroadcastPath("/api/v2"),
	}

	calledPaths := make(map[BroadcastPath]bool)
	mu := sync.Mutex{}

	// Register handlers for all paths
	for _, path := range paths {
		p := path // capture loop variable
		mux.PublishFunc(ctx, p, func(ctx context.Context, tw *TrackWriter) {
			mu.Lock()
			calledPaths[p] = true
			mu.Unlock()
		})
	}

	// Test each path
	for _, path := range paths {
		tw := &TrackWriter{BroadcastPath: path}
		mux.serveTrack(tw)

		mu.Lock()
		called := calledPaths[path]
		mu.Unlock()

		assert.True(t, called, "handler for path %s should be called", path)
	}
}

// Test concurrent access
func TestMux_ConcurrentAccess(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	const numGoroutines = 50
	const pathsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent registration and serving
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < pathsPerGoroutine; j++ {
				path := BroadcastPath(fmt.Sprintf("/test/%d/%d", id, j))

				// Register handler
				mux.PublishFunc(ctx, path, func(ctx context.Context, tw *TrackWriter) {
					// Handler called
				})

				// Serve track
				tw := &TrackWriter{BroadcastPath: path}
				mux.serveTrack(tw)

				// Small delay to increase chance of race conditions
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()
}

// Test path validation function
func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name  string
		path  BroadcastPath
		valid bool
	}{
		{"root", BroadcastPath("/"), true},
		{"simple", BroadcastPath("/test"), true},
		{"nested", BroadcastPath("/api/v1/users"), true},
		{"with_dots", BroadcastPath("/api/user.profile"), true},
		{"with_underscores", BroadcastPath("/api/user_profile"), true},
		{"with_hyphens", BroadcastPath("/api/user-profile"), true},
		{"empty", BroadcastPath(""), false},
		{"no_leading_slash", BroadcastPath("test"), false},
		{"only_dots", BroadcastPath("./test"), false},
		{"double_dots", BroadcastPath("/../test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPath(tt.path)
			assert.Equal(t, tt.valid, result, "isValidPath(%s) should return %v", tt.path, tt.valid)
		})
	}
}

// Test prefix validation function
func TestIsValidPrefix(t *testing.T) {
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
			assert.Equal(t, tt.valid, result, "isValidPrefix(%s) should return %v", tt.prefix, tt.valid)
		})
	}
}

// Test default mux functions
func TestDefaultMux(t *testing.T) {
	// Clean up any previous state
	DefaultMux = NewTrackMux()
	ctx := context.Background()

	path := BroadcastPath("/default/test")

	called := false
	Publish(ctx, path, TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {
		called = true
	}))

	tw := &TrackWriter{BroadcastPath: path}
	DefaultMux.serveTrack(tw)

	assert.True(t, called, "default handler should be called")
}

func TestDefaultMux_PublishFunc(t *testing.T) {
	// Clean up any previous state
	DefaultMux = NewTrackMux()
	ctx := context.Background()

	path := BroadcastPath("/default/test2")

	called := false
	PublishFunc(ctx, path, func(ctx context.Context, tw *TrackWriter) {
		called = true
	})

	tw := &TrackWriter{BroadcastPath: path}
	DefaultMux.serveTrack(tw)

	assert.True(t, called, "default handler function should be called")
}

// Test edge cases
func TestMux_EdgeCases(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	// Test with special characters in path
	specialPath := BroadcastPath("/测试/テスト/🚀/test")
	called := false

	assert.NotPanics(t, func() {
		mux.PublishFunc(ctx, specialPath, func(ctx context.Context, tw *TrackWriter) {
			called = true
		})
	}, "should handle special characters in path")

	tw := &TrackWriter{BroadcastPath: specialPath}
	mux.serveTrack(tw)

	assert.True(t, called, "handler with special characters should be called")
}

// Test overwriting handlers
func TestMux_OverwriteHandler(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	called1 := false
	called2 := false

	handler1 := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) { called1 = true })
	handler2 := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) { called2 = true })

	// Register first handler
	mux.Publish(ctx, path, handler1)

	// Test first handler
	tw := &TrackWriter{BroadcastPath: path}
	mux.serveTrack(tw)
	assert.True(t, called1, "first handler should be called")
	assert.False(t, called2, "second handler should not be called yet")

	// Overwrite with second handler
	called1, called2 = false, false
	mux.Publish(ctx, path, handler2)

	// Test that second handler is now active
	mux.serveTrack(tw)
	assert.False(t, called1, "first handler should not be called after overwrite")
	assert.True(t, called2, "second handler should be called after overwrite")
}

// Test Mux.Clear method
func TestMux_Clear(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()

	// Register some handlers
	paths := []BroadcastPath{
		BroadcastPath("/test1"),
		BroadcastPath("/test2"),
		BroadcastPath("/nested/test3"),
	}

	for _, path := range paths {
		mux.PublishFunc(ctx, path, func(ctx context.Context, tw *TrackWriter) {})
	}

	// Verify handlers are registered
	for _, path := range paths {
		a, handler := mux.TrackHandler(path)
		assert.NotNil(t, a, "Announcement should exist for path %s before Clear", path)
		assert.NotNil(t, handler, "Handler should exist for path %s before Clear", path)
	}

	// Clear the mux
	mux.Clear()

	// Verify all handlers are removed
	for _, path := range paths {
		a, handler := mux.TrackHandler(path)
		assert.Nil(t, a, "Announcement should be nil for path %s after Clear", path)
		// Compare function pointers instead of direct equality
		assert.Equal(t, reflect.ValueOf(NotFoundTrackHandler).Pointer(),
			reflect.ValueOf(handler).Pointer(), "Handler should be NotFoundTrackHandler for path %s after Clear", path)
	}

	// Clear should be idempotent
	assert.NotPanics(t, func() { mux.Clear() }, "Clear should be idempotent")

	// Should be able to register new handlers after clear
	newPath := BroadcastPath("/afterclear")
	called := false
	mux.PublishFunc(ctx, newPath, func(ctx context.Context, tw *TrackWriter) { called = true })

	tw := &TrackWriter{BroadcastPath: newPath}
	mux.serveTrack(tw)
	assert.True(t, called, "handler should work after re-registering post-Clear")
}

// Test Mux.Announce method
func TestMux_Announce(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement, end := NewAnnouncement(ctx, path)
	defer end() // Ensure cleanup

	called := false
	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) { called = true })

	// Test direct announce
	mux.Announce(announcement, handler)

	// Verify handler is registered
	tw := &TrackWriter{BroadcastPath: path}
	mux.serveTrack(tw)
	assert.True(t, called, "announced handler should be called")
}

// Test with inactive announcement
func TestMux_Announce_Inactive(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement, end := NewAnnouncement(ctx, path)
	end() // Make it inactive immediately

	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {})

	// Should not register handler for inactive announcement
	mux.Announce(announcement, handler)

	// Handler should not be registered
	a, foundHandler := mux.TrackHandler(path)
	assert.Nil(t, a, "Announcement should be nil for inactive announcement")
	handlerPtr := reflect.ValueOf(foundHandler).Pointer()
	notFoundPtr := reflect.ValueOf(NotFoundTrackHandler).Pointer()
	assert.Equal(t, notFoundPtr, handlerPtr, "Handler should be NotFoundTrackHandler for inactive announcement")
}

func TestNotFound(t *testing.T) {
	tests := map[string]struct {
		trackWriter *TrackWriter
		expectPanic bool
	}{
		"nil track writer": {
			trackWriter: nil,
			expectPanic: false,
		},
		"track writer with nil TrackWriter": {
			trackWriter: newTrackWriter(BroadcastPath("/test"), TrackName("test"), nil, nil, nil),
			expectPanic: false,
		},
		"track writer with mock TrackWriter": {
			trackWriter: newTrackWriter(BroadcastPath("/test"), TrackName("test"),
				newReceiveSubscribeStream(SubscribeID(1), func() quic.Stream {
					mockStream := &MockQUICStream{}
					mockStream.On("Context").Return(context.Background())
					mockStream.On("Read", mock.Anything).Return(0, io.EOF)
					mockStream.On("CancelWrite", mock.Anything).Return()
					mockStream.On("CancelRead", mock.Anything).Return()
					return mockStream
				}(), &TrackConfig{}),
				func() (quic.SendStream, error) {
					return &MockQUICSendStream{}, nil
				}, func() {}),
			expectPanic: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Should not panic in any case
			NotFound(context.Background(), tt.trackWriter)
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	tests := map[string]struct {
		trackWriter *TrackWriter
		expectPanic bool
	}{
		"nil track writer": {
			trackWriter: nil,
			expectPanic: false,
		},
		"track writer with nil TrackWriter": {
			trackWriter: newTrackWriter(BroadcastPath("/test"), TrackName("test"), nil, nil, nil),
			expectPanic: false,
		},
		"track writer with mock TrackWriter": {
			trackWriter: newTrackWriter(BroadcastPath("/test"), TrackName("test"),
				newReceiveSubscribeStream(SubscribeID(1), func() quic.Stream {
					mockStream := &MockQUICStream{}
					mockStream.On("Context").Return(context.Background())
					mockStream.On("Read", mock.Anything).Return(0, io.EOF)
					mockStream.On("CancelWrite", mock.Anything).Return()
					mockStream.On("CancelRead", mock.Anything).Return()
					return mockStream
				}(), &TrackConfig{}),
				func() (quic.SendStream, error) {
					return &MockQUICSendStream{}, nil
				}, func() {}),
			expectPanic: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			trackWriter := tt.trackWriter

			// Should not panic in any case
			NotFoundTrackHandler.ServeTrack(context.Background(), trackWriter)
		})
	}
}

func TestTrackHandlerFunc(t *testing.T) {
	called := false
	var receivedTrackWriter *TrackWriter

	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {
		called = true
		receivedTrackWriter = tw
	})

	// Create a proper TrackWriter with a valid receiveSubscribeStream
	testTrackWriter := newTrackWriter(BroadcastPath("/test"), TrackName("test"),
		newReceiveSubscribeStream(SubscribeID(1), func() quic.Stream {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			return mockStream
		}(), &TrackConfig{}),
		func() (quic.SendStream, error) {
			return &MockQUICSendStream{}, nil
		}, func() {})

	handler.ServeTrack(context.Background(), testTrackWriter)

	assert.True(t, called, "handler function was not called")
	assert.Equal(t, testTrackWriter, receivedTrackWriter, "handler did not receive the correct track writer")
}

func TestTrackHandlerFuncServeTrack(t *testing.T) {
	callCount := 0

	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {
		callCount++
	})

	// Create a proper TrackWriter with a valid receiveSubscribeStream
	trackWriter := newTrackWriter(BroadcastPath("/test"), TrackName("test"),
		newReceiveSubscribeStream(SubscribeID(1), func() quic.Stream {
			mockStream := &MockQUICStream{}
			mockStream.On("Context").Return(context.Background())
			mockStream.On("Read", mock.Anything).Return(0, io.EOF)
			return mockStream
		}(), &TrackConfig{}),
		func() (quic.SendStream, error) {
			return &MockQUICSendStream{}, nil
		}, func() {})

	handler.ServeTrack(context.Background(), trackWriter)
	handler.ServeTrack(context.Background(), trackWriter)
	handler.ServeTrack(context.Background(), trackWriter)

	assert.Equal(t, 3, callCount, "expected handler to be called 3 times")
}

// Additional robust tests for TrackMux
func TestMux_AnnouncementDeliveryToMultipleListeners(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/multi")

	announcement, end := NewAnnouncement(ctx, path)
	defer end()

	received := make(chan *Announcement, 2)
	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {})

	// Register two listeners for the same path
	mux.Announce(announcement, handler)
	mux.Announce(announcement, handler)

	// Simulate announcement delivery
	go func() {
		received <- announcement
	}()
	go func() {
		received <- announcement
	}()

	// Both listeners should receive the announcement
	count := 0
	for i := 0; i < 2; i++ {
		select {
		case ann := <-received:
			assert.Equal(t, announcement, ann, "Listener should receive the correct announcement")
			count++
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Listener did not receive announcement")
		}
	}
	assert.Equal(t, 2, count, "Both listeners should receive the announcement")
}

func TestMux_AnnouncementTreeCleanup(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/cleanup")

	announcement, end := NewAnnouncement(ctx, path)
	defer end()

	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {})
	mux.Announce(announcement, handler)

	// Simulate listener disconnect
	end()

	// After cleanup, handler should not be registered
	a, h := mux.TrackHandler(path)
	assert.Nil(t, a, "Announcement should be nil after cleanup")
	assert.Equal(t, reflect.ValueOf(NotFoundTrackHandler).Pointer(), reflect.ValueOf(h).Pointer(), "Handler should be NotFoundTrackHandler after cleanup")
}

func TestMux_AnnouncementDeliveryErrorPropagation(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/errorprop")

	announcement, end := NewAnnouncement(ctx, path)
	defer end()

	// Handler that simulates SendAnnouncement error
	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {
		// Simulate error: do nothing, just ensure no panic
	})
	mux.Announce(announcement, handler)

	// No panic should occur
	tw := &TrackWriter{BroadcastPath: path}
	assert.NotPanics(t, func() {
		mux.serveTrack(tw)
	}, "ServeTrack should handle SendAnnouncement error gracefully")
}

func TestMux_AnnounceWithNilAnnouncementOrHandler(t *testing.T) {
	mux := NewTrackMux()
	// Announce with nil Announcement
	assert.NotPanics(t, func() {
		mux.Announce(nil, nil)
	}, "Announce with nil Announcement/handler should not panic")
}

func TestMux_SimultaneousAnnounceAndPublish(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/race")

	var wg sync.WaitGroup
	wg.Add(2)

	handler := TrackHandlerFunc(func(ctx context.Context, tw *TrackWriter) {})
	announcement, end := NewAnnouncement(ctx, path)
	defer end()

	go func() {
		defer wg.Done()
		mux.Announce(announcement, handler)
	}()
	go func() {
		defer wg.Done()
		mux.Publish(ctx, path, handler)
	}()

	wg.Wait()
	// After both, handler should be registered
	_, h := mux.TrackHandler(path)
	assert.NotNil(t, h, "Handler should be registered after simultaneous Announce and Publish")
}

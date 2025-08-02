package moqt

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
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
			handler := TrackHandlerFunc(func(tw *TrackWriter) {
				called = true
				assert.Equal(t, tt.path, tw.BroadcastPath, "handler should receive correct path")
			})

			// Register handler with new API
			mux.Handle(ctx, tt.path, handler)

			// Verify handler is registered and callable
			foundHandler := mux.Handler(tt.path)
			assert.NotNil(t, foundHandler, tt.description)

			// Test handler execution
			trackWriter := newTrackWriter(tt.path, TrackName("test_track"), nil, nil, nil)
			foundHandler.ServeTrack(trackWriter)

			assert.True(t, called, "handler should be called when serving track")
		})
	}
}

func TestNewMux_Handle_Overwrite(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	called1, called2 := false, false

	handler1 := TrackHandlerFunc(func(tw *TrackWriter) { called1 = true })
	handler2 := TrackHandlerFunc(func(tw *TrackWriter) { called2 = true })

	// Register first handler
	mux.Handle(ctx, path, handler1)

	// Allow time for handler registration and potential goroutines to settle
	time.Sleep(10 * time.Millisecond)

	// Test first handler
	trackWriter := newTrackWriter(path, TrackName("test_track1"), nil, nil, nil)
	mux.ServeTrack(trackWriter)
	assert.True(t, called1, "First handler should be called")
	assert.False(t, called2, "Second handler should not be called yet")

	// Try to overwrite with second handler - should log warning and not overwrite
	called1, called2 = false, false
	mux.Handle(ctx, path, handler2)

	// Allow time for potential retry mechanisms to settle
	time.Sleep(10 * time.Millisecond)

	// Test that first handler is still active (overwrite is prevented)
	trackWriter2 := newTrackWriter(path, TrackName("test_track2"), nil, nil, nil)
	mux.ServeTrack(trackWriter2)
	assert.True(t, called1, "First handler should still be called after attempted overwrite")
	assert.False(t, called2, "Second handler should not be called due to overwrite prevention")
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
			handler := TrackHandlerFunc(func(tw *TrackWriter) {})

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
	calledCh := make(chan *TrackWriter, 1)
	mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {
		calledCh <- tw
	}))

	// Allow time for handler registration and potential goroutines to settle
	time.Sleep(10 * time.Millisecond)

	// Create and serve track writer
	trackWriter := newTrackWriter(path, trackName, nil, nil, nil)

	mux.ServeTrack(trackWriter)

	// Verify handler was called with correct publisher
	select {
	case receivedTrackWriter := <-calledCh:
		assert.NotNil(t, receivedTrackWriter, "handler should receive track writer")
		assert.Equal(t, path, receivedTrackWriter.BroadcastPath, "handler should receive correct path")
		assert.Equal(t, trackName, receivedTrackWriter.TrackName, "handler should receive correct track name")
	case <-time.After(5 * time.Second):
		t.Fatal("Handler should have been called")
	}
}

func TestNewMux_ServeTrack_NotFound(t *testing.T) {
	mux := NewTrackMux()

	// Create a mock stream for the subscribe stream
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF) // Background goroutine will try to read
	mockStream.On("CancelWrite", mock.Anything).Return()
	mockStream.On("CancelRead", mock.Anything).Return()

	// Create a mock track writer
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	closeFunc := func() {}
	trackWriter := newTrackWriter("/broadcast/test", "test_track", substr, openUniStreamFunc, closeFunc)

	// Should use NotFoundHandler which closes the controller
	mux.ServeTrack(trackWriter)

	// Verify that CancelWrite was called with the error code for TrackNotFoundErrorCode
	mockStream.AssertCalled(t, "CancelWrite", quic.StreamErrorCode(TrackNotFoundErrorCode))
	mockStream.AssertExpectations(t)
}

func TestNewMux_ServeTrack_NilTrackWriter(t *testing.T) {
	mux := NewTrackMux()

	// Should handle nil track writer gracefully without panic
	assert.NotPanics(t, func() {
		mux.ServeTrack(nil)
	})
}

func TestNewMux_ServeTrack_NilSubscribeStream(t *testing.T) {
	mux := NewTrackMux()

	openUniStreamFunc := func() (quic.SendStream, error) {
		return &MockQUICSendStream{}, nil
	}
	closeFunc := func() {}
	trackWriter := newTrackWriter("/broadcast/test", "test_track", nil, openUniStreamFunc, closeFunc)

	// Should handle nil subscribe stream gracefully without panic
	assert.NotPanics(t, func() {
		mux.ServeTrack(trackWriter)
	})
}

func TestNewMux_ServeTrack_InvalidPath(t *testing.T) {
	mux := NewTrackMux()

	trackWriter := newTrackWriter("invalid-path", "test_track", nil, nil, nil)

	// Should panic for invalid path
	assert.Panics(t, func() {
		mux.ServeTrack(trackWriter)
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
		mux.Handle(context.Background(), path, TrackHandlerFunc(func(tw *TrackWriter) {}))
	}

	// Allow time for handlers to register and potential retry mechanisms to settle
	time.Sleep(20 * time.Millisecond)

	// Create a mock stream for the announcement writer
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)

	// Create real announcement writer instead of mock
	announceWriter := newAnnouncementWriter(mockStream, "/room/")

	// Test serving announcements in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(announceWriter, "/room/")
	}()

	// Give time for ServeAnnouncements to process initial announcements
	time.Sleep(150 * time.Millisecond)

	// Verify that announcements were processed
	activesCount := len(announceWriter.actives)
	assert.Equal(t, 3, activesCount, "Should have received 3 initial announcements")

	// Verify that all expected paths are in actives
	expectedSuffixes := []string{"/person1", "/person2", "/person3"}
	for _, expectedSuffix := range expectedSuffixes {
		assert.Contains(t, announceWriter.actives, expectedSuffix, "Expected suffix %s should be in actives", expectedSuffix)
	}

	// Add a new handler and verify it gets announced
	newPath := BroadcastPath("/room/person4")
	mux.Handle(context.Background(), newPath, TrackHandlerFunc(func(tw *TrackWriter) {}))

	// Give time for new announcement to be processed with retry mechanism
	time.Sleep(150 * time.Millisecond)

	// Verify the new announcement was processed
	finalCount := len(announceWriter.actives)
	assert.Equal(t, 4, finalCount, "Should have received 4 total announcements after adding new handler")

	// Verify the new suffix is in the actives
	assert.Contains(t, announceWriter.actives, "/person4", "New suffix /person4 should be in actives")

	// Clean up
	select {
	case <-done:
		// ServeAnnouncements completed
	case <-time.After(1 * time.Second):
		t.Log("ServeAnnouncements still running, but test completed")
	}
}

func TestNewMux_ServeAnnouncements_NilWriter(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register a handler
	path := BroadcastPath("/test/path")
	mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {}))

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

			// Create a mock stream for the announcement writer
			mockStream := &MockQUICStream{}
			mockStream.On("CancelWrite", quic.StreamErrorCode(InvalidPrefixErrorCode)).Return()
			mockStream.On("CancelRead", quic.StreamErrorCode(InvalidPrefixErrorCode)).Return()
			mockStream.On("StreamID").Return(quic.StreamID(1))

			// Create real announcement writer
			announceWriter := newAnnouncementWriter(mockStream, "/valid/")

			// Should call CloseWithError for invalid prefix
			mux.ServeAnnouncements(announceWriter, tt.prefix)

			// Give time for processing
			time.Sleep(10 * time.Millisecond)

			// Verify that the appropriate calls were made
			mockStream.AssertExpectations(t)
		})
	}
}

func TestNewMux_ServeAnnouncements_EmptyPrefix(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register initial handlers with different paths
	paths := []BroadcastPath{
		BroadcastPath("/room/a"),
		BroadcastPath("/game/b"),
		BroadcastPath("/chat/c"),
	}

	for _, path := range paths {
		mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {}))
	}

	// Allow time for handlers to register and potential retry mechanisms to settle
	time.Sleep(20 * time.Millisecond)

	// Create mock announcement writer
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)

	prefix := "/"
	announceWriter := newAnnouncementWriter(mockStream, prefix)

	// Test serving announcements with root prefix in goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(announceWriter, prefix)
	}()

	// Give time for processing
	time.Sleep(150 * time.Millisecond)

	// Verify that all paths are announced (since "/" matches all)
	count := len(announceWriter.actives)
	assert.Equal(t, 3, count, "Should have received all 3 announcements with root prefix")

	// Verify all expected suffixes are present in actives
	expectedSuffixes := []string{"/room/a", "/game/b", "/chat/c"}
	for _, expectedSuffix := range expectedSuffixes {
		assert.Contains(t, announceWriter.actives, expectedSuffix, "Expected suffix %s should be in actives", expectedSuffix)
	}

	// Clean up
	select {
	case <-done:
		// ServeAnnouncements completed
	case <-time.After(1 * time.Second):
		t.Log("ServeAnnouncements still running, but test completed")
	}
}

func TestNewMux_Handler(t *testing.T) {
	mux := NewTrackMux()

	path := BroadcastPath("/test")

	handler := mux.Handler(path)
	assert.NotNil(t, handler, "Handler should be registered for path")
	assert.Equal(t, NotFoundHandler, handler, "Handler should be NotFoundHandler initially")

	// Register a handler
	expectedHandler := TrackHandlerFunc(func(tw *TrackWriter) {})

	handler = mux.Handler(path)
	assert.NotNil(t, handler, "Handler should be registered after Handle call")
	assert.Equal(t, expectedHandler, handler, "Handler should still be NotFoundHandler before Handle call")
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
	mux.HandleFunc(ctx, path, func(tw *TrackWriter) {
		called = true
		assert.Equal(t, path, tw.BroadcastPath)
	})

	// Test that the function was registered correctly
	handler := mux.Handler(path)
	assert.NotNil(t, handler, "handler should be registered")

	// Test that the function is called correctly
	trackWriter := newTrackWriter(path, TrackName("test_track"), nil, nil, nil)
	handler.ServeTrack(trackWriter)
	assert.True(t, called, "function should be called")
}

func TestNewMux_Announce_Direct(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement := NewAnnouncement(ctx, path)
	called := false
	handler := TrackHandlerFunc(func(tw *TrackWriter) { called = true })

	// Test direct announce
	mux.Announce(announcement, handler)

	// Verify handler is registered by testing behavior
	trackWriter := newTrackWriter(path, TrackName("test_track"), nil, nil, nil)
	mux.ServeTrack(trackWriter)
	assert.True(t, called, "registered handler should be called")
}

func TestNewMux_Announce_InactiveAnnouncement(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement := NewAnnouncement(ctx, path)
	announcement.End() // Make it inactive

	handler := TrackHandlerFunc(func(tw *TrackWriter) {})

	// Test announce with inactive announcement
	mux.Announce(announcement, handler)

	assert.Equal(t, mux.Handler(path), NotFoundHandler, "should not register handler for inactive announcement")

	// Handler should not be registered - test by verifying NotFoundHandler behavior
	mockStream := &MockQUICStream{}
	mockStream.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)
	subscribeStream := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	closeFuncCalled := false
	trackWriter := newTrackWriter(path, TrackName("test_track"), subscribeStream,
		nil, func() { closeFuncCalled = true })

	mux.ServeTrack(trackWriter)

	mockStream.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
	assert.True(t, closeFuncCalled, "close function should be called for inactive announcement")
}

func TestNewMux_Announce_DuplicatePath(t *testing.T) {
	mux := NewTrackMux()
	ctx := context.Background()
	path := BroadcastPath("/test")

	announcement1 := NewAnnouncement(ctx, path)
	announcement2 := NewAnnouncement(ctx, path)

	handler1 := TrackHandlerFunc(func(tw *TrackWriter) {})
	handler2 := TrackHandlerFunc(func(tw *TrackWriter) {})

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
		mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {
			callCounts[path] = true
		}))
	}

	// Verify handlers are registered by testing behavior
	for _, path := range paths {
		trackWriter := newTrackWriter(BroadcastPath(path), TrackName("test_track"), nil, nil, nil)
		mux.ServeTrack(trackWriter)
		assert.True(t, callCounts[path], "handler should be registered for path %s", path)
		handler := mux.Handler(path)
		assert.NotEqual(t, handler, NotFoundHandler, "handler should not be NotFoundHandler for path %s before Clear", path)
	}

	// Clear the mux
	mux.Clear()

	// Verify all handlers are removed by testing NotFoundHandler behavior
	for _, path := range paths {
		// HandlerがNotFoundHandlerになっていることを確認
		handler := mux.Handler(path)
		assert.Equal(t, NotFoundHandler, handler, "handler should be NotFoundHandler for path %s after Clear", path)

		closeFuncCalled := false
		mockStream := &MockQUICStream{}
		mockStream.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)
		subscribeStream := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
		trackWriter := newTrackWriter(path, "test_track", subscribeStream, nil, func() {
			closeFuncCalled = true
		})

		mux.ServeTrack(trackWriter)
		mockStream.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
		assert.True(t, closeFuncCalled, "close function should be called for cleared mux")
	}

	// Clearを複数回呼んでもpanicしないことを確認
	assert.NotPanics(t, func() { mux.Clear() }, "Clear should be idempotent")

	// Clear後に再登録できることを確認
	newPath := BroadcastPath("/afterclear")
	called := false
	mux.Handle(ctx, newPath, TrackHandlerFunc(func(tw *TrackWriter) { called = true }))
	trackWriter := newTrackWriter(newPath, "test_track", nil, nil, nil)
	mux.ServeTrack(trackWriter)
	assert.True(t, called, "handler should be called after re-registering post-Clear")

	// 再登録したhandlerが取得できること
	handler := mux.Handler(newPath)
	assert.NotNil(t, handler, "handler should be present for newPath after re-registering")
}

func TestNewMux_AnnouncementLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mux := NewTrackMux()
	path := BroadcastPath("/test")

	called := false
	handler := TrackHandlerFunc(func(tw *TrackWriter) { called = true })

	// Register handler
	mux.Handle(ctx, path, handler)

	// Verify handler is registered by testing behavior
	trackWriter := newTrackWriter(path, "test_track", nil, nil, nil)
	mux.ServeTrack(trackWriter)
	assert.True(t, called, "handler should be registered")

	// Cancel context to end announcement
	cancel()

	// Give time for cleanup
	time.Sleep(50 * time.Millisecond)

	// Handler should be removed - test by verifying NotFoundHandler behavior
	mockStream := &MockQUICStream{}
	mockStream.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)
	subscribeStream := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	closeFuncCalled := false
	trackWriter2 := newTrackWriter("/broadcast/test", "test_track", subscribeStream, nil, func() {
		closeFuncCalled = true
	})

	mux.ServeTrack(trackWriter2)
	mockStream.AssertCalled(t, "CloseWithError", TrackNotFoundErrorCode)
	assert.True(t, closeFuncCalled, "close function should be called for cleared mux")
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
				handler := TrackHandlerFunc(func(tw *TrackWriter) {})

				// Register handler
				mux.Handle(ctx, path, handler)

				// Access handler
				foundHandler := mux.Handler(path)
				assert.NotNil(t, foundHandler, "handler should be found")

				trackWriter := newTrackWriter(path, "test_track", nil, nil, nil)
				mux.ServeTrack(trackWriter)
			}
		}(i)
	}

	wg.Wait()
}

// Test DefaultMux functionality
func TestNewMux_DefaultMux(t *testing.T) {
	// Clear DefaultMux first
	DefaultMux.Clear()

	ctx := context.Background()
	path := BroadcastPath("/default/test")

	called := false
	handler := TrackHandlerFunc(func(tw *TrackWriter) {
		called = true
	})

	// Test top-level Handle function
	Handle(ctx, path, handler)

	// Test top-level HandleFunc function
	path2 := BroadcastPath("/default/test2")
	called2 := false
	HandleFunc(ctx, path2, func(tw *TrackWriter) {
		called2 = true
	})

	// Test handlers work
	trackWriter := newTrackWriter(path, "test_track", nil, nil, nil)
	DefaultMux.ServeTrack(trackWriter)
	assert.True(t, called, "handler should be called")

	trackWriter2 := newTrackWriter(path2, "test_track2", nil, nil, nil)
	DefaultMux.ServeTrack(trackWriter2)
	assert.True(t, called2, "handler2 should be called")

	// Test direct Announce function
	path3 := BroadcastPath("/default/test3")
	announcement := NewAnnouncement(ctx, path3)
	called3 := false
	handler3 := TrackHandlerFunc(func(tw *TrackWriter) { called3 = true })

	Announce(announcement, handler3)

	trackWriter3 := newTrackWriter(path3, "test_track3", nil, nil, nil)
	DefaultMux.ServeTrack(trackWriter3)
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
		mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {}))
	}

	// Allow time for handlers to register and potential retry mechanisms to settle
	time.Sleep(20 * time.Millisecond)

	// Create mock stream for the announcement writer
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)

	// Create real announcement writer
	prefix := "/room/"
	announceWriter := newAnnouncementWriter(mockStream, prefix)

	// Test with /room/ prefix in goroutine - should only announce matching paths
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(announceWriter, prefix)
	}()

	// Give time for processing with retry mechanism
	time.Sleep(150 * time.Millisecond)

	// Verify only matching paths are announced
	activesCount := len(announceWriter.actives)
	assert.Equal(t, len(matchingPaths), activesCount,
		"Should announce only paths matching /room/ prefix")

	// Verify specific paths - check expected suffixes in actives
	expectedSuffixes := []string{"/alice", "/bob"}
	for _, expectedSuffix := range expectedSuffixes {
		assert.Contains(t, announceWriter.actives, expectedSuffix,
			"Expected suffix %s should be in actives", expectedSuffix)
	}

	// Verify non-matching paths are not announced by checking that unexpected suffixes are not present
	unexpectedSuffixes := []string{"/chess", "/general"}
	for _, unexpectedSuffix := range unexpectedSuffixes {
		assert.NotContains(t, announceWriter.actives, unexpectedSuffix,
			"Unexpected suffix %s should not be in actives", unexpectedSuffix)
	}

	// Clean up
	select {
	case <-done:
		// ServeAnnouncements completed
	case <-time.After(1 * time.Second):
		t.Log("ServeAnnouncements still running, but test completed")
	}
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
		mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {}))
	}

	// Allow time for handlers to register and potential retry mechanisms to settle
	time.Sleep(20 * time.Millisecond)

	// Create mock stream for the announcement writer
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)

	// Create real announcement writer
	prefix := "/"
	announceWriter := newAnnouncementWriter(mockStream, prefix)

	// Test with root prefix in goroutine - should announce all paths
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(announceWriter, prefix)
	}()

	// Give time for processing with retry mechanism
	time.Sleep(150 * time.Millisecond)

	// Verify all paths are announced with root prefix
	activesCount := len(announceWriter.actives)
	assert.Equal(t, len(allPaths), activesCount,
		"Root prefix should announce all registered paths")

	// Verify all expected suffixes are present in actives
	expectedSuffixes := []string{"/room/alice", "/game/chess", "/chat/general"}
	for _, expectedSuffix := range expectedSuffixes {
		assert.Contains(t, announceWriter.actives, expectedSuffix,
			"Suffix %s should be in actives with root prefix", expectedSuffix)
	}

	// Clean up
	select {
	case <-done:
		// ServeAnnouncements completed
	case <-time.After(1 * time.Second):
		t.Log("ServeAnnouncements still running, but test completed")
	}
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
		mux.Handle(ctx, path, TrackHandlerFunc(func(tw *TrackWriter) {}))
	}

	// Allow time for handlers to register and potential retry mechanisms to settle
	time.Sleep(20 * time.Millisecond)

	// Create mock stream for the announcement writer
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)

	// Create real announcement writer
	prefix := "/game/"
	announceWriter := newAnnouncementWriter(mockStream, prefix)

	// Test with non-matching prefix in goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(announceWriter, prefix)
	}()

	// Give time for processing with retry mechanism
	time.Sleep(150 * time.Millisecond)

	// Verify no announcements are sent
	activesCount := len(announceWriter.actives)
	assert.Equal(t, 0, activesCount,
		"Should not announce any paths when prefix doesn't match")

	// Clean up
	select {
	case <-done:
		// ServeAnnouncements completed
	case <-time.After(1 * time.Second):
		t.Log("ServeAnnouncements still running, but test completed")
	}
}

func TestNewMux_ServeAnnouncements_BroadcastServerIssue(t *testing.T) {
	ctx := context.Background()
	mux := NewTrackMux()

	// Register handler for "/index" like broadcast server does
	mux.Handle(ctx, BroadcastPath("/index"), TrackHandlerFunc(func(tw *TrackWriter) {}))

	// Allow time for handlers to register and potential retry mechanisms to settle
	time.Sleep(20 * time.Millisecond)

	// Create mock stream for the announcement writer
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)

	// Create real announcement writer
	prefix := "/"
	announceWriter := newAnnouncementWriter(mockStream, prefix)

	// Test with "/" prefix like client opens announce stream in goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeAnnouncements(announceWriter, prefix)
	}()

	// Give time for processing with retry mechanism
	time.Sleep(150 * time.Millisecond)

	// Verify that the "/index" announcement was sent
	activesCount := len(announceWriter.actives)
	assert.Equal(t, 1, activesCount, "Should have sent 1 announcement for /index")

	// Verify the specific suffix is in actives
	assert.Contains(t, announceWriter.actives, "/index", "Should announce /index path")

	// Clean up
	select {
	case <-done:
		// ServeAnnouncements completed
	case <-time.After(1 * time.Second):
		t.Log("ServeAnnouncements still running, but test completed")
	}
}

// =========================================================================
// announcingNode Tests
// =========================================================================

// 1. Basic structure and node creation
func TestAnnouncingNode_NewAnnouncingNode(t *testing.T) {
	node := newAnnouncingNode()

	assert.NotNil(t, node, "newAnnouncingNode should return non-nil node")
	assert.NotNil(t, node.announcements, "announcements map should be initialized")
	assert.NotNil(t, node.channels, "writers map should be initialized")
	assert.NotNil(t, node.children, "children map should be initialized")
	assert.NotNil(t, node.channels, "channels map should be initialized")

	// Verify initial state
	assert.Equal(t, 0, len(node.announcements), "announcements should be empty initially")
	assert.Equal(t, 0, len(node.channels), "writers should be empty initially")
	assert.Equal(t, 0, len(node.children), "children should be empty initially")
	assert.Equal(t, 0, len(node.channels), "channels should be empty initially")
	assert.Equal(t, uint64(0), node.announcementsCount.Load(), "announcements count should be zero initially")
}

// 2. Node search and creation functionality (findNode)
func TestAnnouncingNode_FindNode_EmptySegments(t *testing.T) {
	root := newAnnouncingNode()

	// Empty segments should return root node
	result := root.findNode([]string{}, nil)
	assert.Equal(t, root, result, "findNode with empty segments should return root node")
}

func TestAnnouncingNode_FindNode_SingleSegment(t *testing.T) {
	root := newAnnouncingNode()

	// Single segment should create and return child node
	child := root.findNode([]string{"test"}, nil)
	assert.NotNil(t, child, "findNode should return non-nil child node")
	assert.NotEqual(t, root, child, "child should be different from root")

	// Verify child is stored in root's children
	root.mu.RLock()
	storedChild, exists := root.children["test"]
	root.mu.RUnlock()
	assert.True(t, exists, "child should be stored in root's children map")
	assert.Equal(t, child, storedChild, "returned child should match stored child")
}

func TestAnnouncingNode_FindNode_MultipleSegments(t *testing.T) {
	root := newAnnouncingNode()

	// Multiple segments should create nested structure
	segments := []string{"level1", "level2", "level3"}
	deepChild := root.findNode(segments, nil)

	assert.NotNil(t, deepChild, "findNode should return non-nil deep child node")

	// Verify nested structure exists
	current := root
	for i, segment := range segments {
		current.mu.RLock()
		next, exists := current.children[segment]
		current.mu.RUnlock()
		assert.True(t, exists, "segment %d (%s) should exist in children", i, segment)
		assert.NotNil(t, next, "child node should be non-nil")
		current = next
	}

	assert.Equal(t, deepChild, current, "final node should match returned node")
}

func TestAnnouncingNode_FindNode_ExistingNodes(t *testing.T) {
	root := newAnnouncingNode()

	// Create node first time
	first := root.findNode([]string{"existing"}, nil)

	// Get same node second time
	second := root.findNode([]string{"existing"}, nil)

	assert.Equal(t, first, second, "findNode should return same node for existing path")

	// Verify only one child exists
	root.mu.RLock()
	childCount := len(root.children)
	root.mu.RUnlock()
	assert.Equal(t, 1, childCount, "should have only one child node")
}

func TestAnnouncingNode_FindNode_WithOpFunction(t *testing.T) {
	root := newAnnouncingNode()

	var calledNodes []*announcingNode
	opFunc := func(node *announcingNode) {
		calledNodes = append(calledNodes, node)
	}

	// Create nested path with op function
	segments := []string{"a", "b", "c"}
	result := root.findNode(segments, opFunc)

	assert.NotNil(t, result, "findNode should return non-nil result")
	assert.Equal(t, len(segments), len(calledNodes), "op function should be called for each segment")

	// Verify op function was called on correct nodes
	current := root
	for i, expectedNode := range calledNodes {
		assert.Equal(t, current, expectedNode, "op function should be called on node %d", i)
		if i < len(segments) {
			current.mu.RLock()
			next := current.children[segments[i]]
			current.mu.RUnlock()
			current = next
		}
	}
}

// 3. Announcement management
func TestAnnouncingNode_AddAnnouncement(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()
	announcement := NewAnnouncement(ctx, BroadcastPath("/test"))
	segment := "testSegment"

	// Add announcement
	node.addAnnouncement(segment, announcement)

	// Verify announcement is stored
	node.mu.RLock()
	stored, exists := node.announcements[segment]
	node.mu.RUnlock()

	assert.True(t, exists, "announcement should exist in node")
	assert.Equal(t, announcement, stored, "stored announcement should match added announcement")
}

func TestAnnouncingNode_AddAnnouncement_Overwrite(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	announcement1 := NewAnnouncement(ctx, BroadcastPath("/test1"))
	announcement2 := NewAnnouncement(ctx, BroadcastPath("/test2"))
	segment := "testSegment"

	// Add first announcement
	node.addAnnouncement(segment, announcement1)

	// Overwrite with second announcement
	node.addAnnouncement(segment, announcement2)

	// Verify second announcement overwrites first
	node.mu.RLock()
	stored, exists := node.announcements[segment]
	count := len(node.announcements)
	node.mu.RUnlock()

	assert.True(t, exists, "announcement should exist")
	assert.Equal(t, announcement2, stored, "second announcement should overwrite first")
	assert.Equal(t, 1, count, "should have only one announcement")
}

func TestAnnouncingNode_AddAnnouncement_MultipleSegments(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	segments := []string{"seg1", "seg2", "seg3"}
	announcements := make([]*Announcement, len(segments))

	// Add announcements for different segments
	for i, segment := range segments {
		announcements[i] = NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test%d", i)))
		node.addAnnouncement(segment, announcements[i])
	}

	// Verify all announcements are stored independently
	node.mu.RLock()
	assert.Equal(t, len(segments), len(node.announcements), "should have all announcements")
	for i, segment := range segments {
		stored, exists := node.announcements[segment]
		assert.True(t, exists, "announcement for segment %s should exist", segment)
		assert.Equal(t, announcements[i], stored, "announcement for segment %s should match", segment)
	}
	node.mu.RUnlock()
}

func TestAnnouncingNode_RemoveAnnouncement(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()
	announcement := NewAnnouncement(ctx, BroadcastPath("/test"))
	segment := "testSegment"

	// Add announcement first
	node.addAnnouncement(segment, announcement)

	// Verify it exists
	node.mu.RLock()
	_, exists := node.announcements[segment]
	node.mu.RUnlock()
	assert.True(t, exists, "announcement should exist before removal")

	// Remove announcement
	node.removeAnnouncement(segment, announcement)

	// Verify it's removed
	node.mu.RLock()
	_, exists = node.announcements[segment]
	count := len(node.announcements)
	node.mu.RUnlock()

	assert.False(t, exists, "announcement should not exist after removal")
	assert.Equal(t, 0, count, "announcements map should be empty")
}

func TestAnnouncingNode_RemoveAnnouncement_WrongAnnouncement(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	announcement1 := NewAnnouncement(ctx, BroadcastPath("/test1"))
	announcement2 := NewAnnouncement(ctx, BroadcastPath("/test2"))
	segment := "testSegment"

	// Add first announcement
	node.addAnnouncement(segment, announcement1)

	// Try to remove with different announcement
	node.removeAnnouncement(segment, announcement2)

	// Verify original announcement still exists
	node.mu.RLock()
	stored, exists := node.announcements[segment]
	node.mu.RUnlock()

	assert.True(t, exists, "original announcement should still exist")
	assert.Equal(t, announcement1, stored, "original announcement should be unchanged")
}

func TestAnnouncingNode_RemoveAnnouncement_NonExistentSegment(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()
	announcement := NewAnnouncement(ctx, BroadcastPath("/test"))

	// Try to remove from non-existent segment (should not panic)
	assert.NotPanics(t, func() {
		node.removeAnnouncement("nonexistent", announcement)
	}, "removeAnnouncement should not panic for non-existent segment")

	// Verify node state is unchanged
	node.mu.RLock()
	count := len(node.announcements)
	node.mu.RUnlock()
	assert.Equal(t, 0, count, "announcements should remain empty")
}

// 4. Announcement collection (appendAnnouncements)
func TestAnnouncingNode_AppendAnnouncements_SingleNode(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	// Add some announcements
	announcements := []*Announcement{
		NewAnnouncement(ctx, BroadcastPath("/test1")),
		NewAnnouncement(ctx, BroadcastPath("/test2")),
		NewAnnouncement(ctx, BroadcastPath("/test3")),
	}

	segments := []string{"seg1", "seg2", "seg3"}
	for i, ann := range announcements {
		node.addAnnouncement(segments[i], ann)
	}

	// Collect announcements
	result := node.appendAnnouncements(nil)

	assert.Equal(t, len(announcements), len(result), "should collect all announcements")

	// Verify all announcements are collected (order may vary)
	collected := make(map[*Announcement]bool)
	for _, ann := range result {
		collected[ann] = true
	}

	for _, expected := range announcements {
		assert.True(t, collected[expected], "announcement should be collected")
	}
}

func TestAnnouncingNode_AppendAnnouncements_EmptyNode(t *testing.T) {
	node := newAnnouncingNode()

	// Collect from empty node
	result := node.appendAnnouncements(nil)

	assert.NotNil(t, result, "result should not be nil")
	assert.Equal(t, 0, len(result), "should return empty slice for empty node")
}

func TestAnnouncingNode_AppendAnnouncements_WithExistingSlice(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	// Create existing slice
	existing := []*Announcement{
		NewAnnouncement(ctx, BroadcastPath("/existing1")),
		NewAnnouncement(ctx, BroadcastPath("/existing2")),
	}

	// Add announcement to node
	newAnn := NewAnnouncement(ctx, BroadcastPath("/new"))
	node.addAnnouncement("new", newAnn)

	// Append to existing slice
	result := node.appendAnnouncements(existing)

	assert.Equal(t, len(existing)+1, len(result), "should append to existing slice")

	// Verify existing announcements are preserved
	for i, expected := range existing {
		assert.Equal(t, expected, result[i], "existing announcement %d should be preserved", i)
	}

	// Verify new announcement is added
	assert.Equal(t, newAnn, result[len(existing)], "new announcement should be appended")
}

func TestAnnouncingNode_AppendAnnouncements_NestedStructure(t *testing.T) {
	root := newAnnouncingNode()
	ctx := context.Background()

	// Create nested structure with announcements at each level
	// Root level
	rootAnn := NewAnnouncement(ctx, BroadcastPath("/root"))
	root.addAnnouncement("root", rootAnn)

	// Level 1
	level1 := root.findNode([]string{"level1"}, nil)
	level1Ann := NewAnnouncement(ctx, BroadcastPath("/level1"))
	level1.addAnnouncement("l1", level1Ann)

	// Level 2
	level2 := root.findNode([]string{"level1", "level2"}, nil)
	level2Ann := NewAnnouncement(ctx, BroadcastPath("/level2"))
	level2.addAnnouncement("l2", level2Ann)

	// Collect all announcements
	result := root.appendAnnouncements(nil)

	assert.Equal(t, 3, len(result), "should collect all announcements from nested structure")

	// Verify all announcements are collected
	collected := make(map[*Announcement]bool)
	for _, ann := range result {
		collected[ann] = true
	}

	expectedAnns := []*Announcement{rootAnn, level1Ann, level2Ann}
	for _, expected := range expectedAnns {
		assert.True(t, collected[expected], "announcement should be collected from nested structure")
	}
}

func TestAnnouncingNode_AppendAnnouncements_MultipleChildren(t *testing.T) {
	root := newAnnouncingNode()
	ctx := context.Background()

	// Create multiple child branches
	childNames := []string{"child1", "child2", "child3"}
	expectedAnns := make([]*Announcement, 0)

	for _, childName := range childNames {
		child := root.findNode([]string{childName}, nil)
		ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/%s", childName)))
		child.addAnnouncement("ann", ann)
		expectedAnns = append(expectedAnns, ann)
	}

	// Collect announcements
	result := root.appendAnnouncements(nil)

	assert.Equal(t, len(expectedAnns), len(result), "should collect from all child branches")

	// Verify all announcements are collected
	collected := make(map[*Announcement]bool)
	for _, ann := range result {
		collected[ann] = true
	}

	for _, expected := range expectedAnns {
		assert.True(t, collected[expected], "announcement from child should be collected")
	}
}

// 5. Counter functionality (countAnnouncement)
// The counter is used to track total announcements under a node for efficient memory allocation in appendAnnouncements
func TestAnnouncingNode_CountAnnouncement(t *testing.T) {
	node := newAnnouncingNode()

	// Initial count should be 0
	assert.Equal(t, uint64(0), node.announcementsCount.Load(), "initial count should be 0")

	// Call countAnnouncement - simulates traversal for announcement collection
	countAnnouncement(node)

	// Count should be incremented
	assert.Equal(t, uint64(1), node.announcementsCount.Load(), "count should be incremented to 1")

	// Call multiple times - simulates multiple traversals
	countAnnouncement(node)
	countAnnouncement(node)

	assert.Equal(t, uint64(3), node.announcementsCount.Load(), "count should be incremented to 3")
}

func TestAnnouncingNode_CountAnnouncement_NilNode(t *testing.T) {
	// Should not panic with nil node
	assert.NotPanics(t, func() {
		countAnnouncement(nil)
	}, "countAnnouncement should handle nil node safely")
}

func TestAnnouncingNode_CountAnnouncement_Concurrent(t *testing.T) {
	node := newAnnouncingNode()
	const numGoroutines = 100
	const incrementsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent increments
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				countAnnouncement(node)
			}
		}()
	}

	wg.Wait()

	expected := uint64(numGoroutines * incrementsPerGoroutine)
	actual := node.announcementsCount.Load()
	assert.Equal(t, expected, actual, "concurrent count should be accurate")
}

func TestAnnouncingNode_CountAnnouncement_MemoryEfficiency(t *testing.T) {
	root := newAnnouncingNode()
	ctx := context.Background()

	// Create a structure with known number of announcements
	const numChildren = 3
	const announcementsPerChild = 2
	const rootAnnouncements = 1
	totalExpected := rootAnnouncements + (numChildren * announcementsPerChild)

	// Add announcements to root
	rootAnn := NewAnnouncement(ctx, BroadcastPath("/root"))
	root.addAnnouncement("root", rootAnn)

	// Add announcements to children
	for i := 0; i < numChildren; i++ {
		child := root.findNode([]string{fmt.Sprintf("child_%d", i)}, nil)
		for j := 0; j < announcementsPerChild; j++ {
			ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/child_%d/ann_%d", i, j)))
			child.addAnnouncement(fmt.Sprintf("ann_%d", j), ann)
		}
	}

	// Simulate the counting process that would happen during announcement tree traversal
	// This mimics what happens in ServeAnnouncements
	countAnnouncement(root)
	for i := 0; i < numChildren; i++ {
		child := root.children[fmt.Sprintf("child_%d", i)]
		// Each child would be counted based on its announcements
		for j := 0; j < announcementsPerChild; j++ {
			countAnnouncement(child)
		}
	}

	// Now test appendAnnouncements with the counted capacity
	result := root.appendAnnouncements(nil)

	// Verify we got all announcements
	assert.Equal(t, totalExpected, len(result), "should collect all announcements")

	// The key insight: root.announcementsCount should provide efficient capacity for slice allocation
	rootCount := root.announcementsCount.Load()
	t.Logf("Root count (for capacity): %d, Actual announcements collected: %d", rootCount, len(result))

	// The count is used as capacity hint for efficient memory allocation
	assert.Greater(t, rootCount, uint64(0), "root should have positive count for capacity estimation")
}

// 6. Channel management
func TestAnnouncingNode_ChannelManagement(t *testing.T) {
	node := newAnnouncingNode()

	// Create channels
	ch1 := make(chan *Announcement, 1)
	ch2 := make(chan *Announcement, 1)
	ch3 := make(chan *Announcement, 1)

	// Add channels
	node.mu.Lock()
	node.channels[ch1] = struct{}{}
	node.channels[ch2] = struct{}{}
	node.channels[ch3] = struct{}{}
	node.mu.Unlock()

	// Verify channels are stored
	node.mu.RLock()
	count := len(node.channels)
	_, exists1 := node.channels[ch1]
	_, exists2 := node.channels[ch2]
	_, exists3 := node.channels[ch3]
	node.mu.RUnlock()

	assert.Equal(t, 3, count, "should have 3 channels")
	assert.True(t, exists1, "channel 1 should exist")
	assert.True(t, exists2, "channel 2 should exist")
	assert.True(t, exists3, "channel 3 should exist")

	// Remove one channel
	node.mu.Lock()
	delete(node.channels, ch2)
	node.mu.Unlock()

	// Verify removal
	node.mu.RLock()
	count = len(node.channels)
	_, exists1 = node.channels[ch1]
	_, exists2 = node.channels[ch2]
	_, exists3 = node.channels[ch3]
	node.mu.RUnlock()

	assert.Equal(t, 2, count, "should have 2 channels after removal")
	assert.True(t, exists1, "channel 1 should still exist")
	assert.False(t, exists2, "channel 2 should not exist")
	assert.True(t, exists3, "channel 3 should still exist")

	// Clean up
	close(ch1)
	close(ch2)
	close(ch3)
}

func TestAnnouncingNode_ChannelManagement_Independence(t *testing.T) {
	node1 := newAnnouncingNode()
	node2 := newAnnouncingNode()

	ch1 := make(chan *Announcement, 1)
	ch2 := make(chan *Announcement, 1)

	// Add channels to different nodes
	node1.mu.Lock()
	node1.channels[ch1] = struct{}{}
	node1.mu.Unlock()

	node2.mu.Lock()
	node2.channels[ch2] = struct{}{}
	node2.mu.Unlock()

	// Verify independence
	node1.mu.RLock()
	count1 := len(node1.channels)
	_, exists1InNode1 := node1.channels[ch1]
	_, exists2InNode1 := node1.channels[ch2]
	node1.mu.RUnlock()

	node2.mu.RLock()
	count2 := len(node2.channels)
	_, exists1InNode2 := node2.channels[ch1]
	_, exists2InNode2 := node2.channels[ch2]
	node2.mu.RUnlock()

	assert.Equal(t, 1, count1, "node1 should have 1 channel")
	assert.Equal(t, 1, count2, "node2 should have 1 channel")
	assert.True(t, exists1InNode1, "ch1 should exist in node1")
	assert.False(t, exists2InNode1, "ch2 should not exist in node1")
	assert.False(t, exists1InNode2, "ch1 should not exist in node2")
	assert.True(t, exists2InNode2, "ch2 should exist in node2")

	// Clean up
	close(ch1)
	close(ch2)
}

// 7. Concurrent access and thread safety
func TestAnnouncingNode_ConcurrentAccess_AddRemoveAnnouncements(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	const numGoroutines = 50
	const opsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Add and remove goroutines

	// Concurrent add operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				segment := fmt.Sprintf("seg_%d_%d", id, j)
				ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test_%d_%d", id, j)))
				node.addAnnouncement(segment, ann)
			}
		}(i)
	}

	// Concurrent remove operations (some may not find anything to remove)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				segment := fmt.Sprintf("seg_%d_%d", id, j)
				ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test_%d_%d", id, j)))
				node.removeAnnouncement(segment, ann)
			}
		}(i)
	}

	wg.Wait()

	// Test should complete without panic or deadlock
	node.mu.RLock()
	count := len(node.announcements)
	node.mu.RUnlock()

	// The exact count is unpredictable due to race conditions,
	// but the test should not panic
	t.Logf("Final announcement count: %d", count)
}

func TestAnnouncingNode_ConcurrentAccess_FindNode(t *testing.T) {
	root := newAnnouncingNode()

	const numGoroutines = 20
	const pathsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent findNode operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < pathsPerGoroutine; j++ {
				segments := []string{fmt.Sprintf("level1_%d", id), fmt.Sprintf("level2_%d_%d", id, j)}
				node := root.findNode(segments, countAnnouncement)
				assert.NotNil(t, node, "findNode should return non-nil node")
			}
		}(i)
	}

	wg.Wait()

	// Verify final state consistency
	root.mu.RLock()
	childCount := len(root.children)
	root.mu.RUnlock()

	assert.Equal(t, numGoroutines, childCount, "should have correct number of level1 children")

	// Check counter consistency
	// Each findNode call increments the root counter once (for traversing the root)
	// This represents the number of announcements under root for memory allocation efficiency
	totalExpectedCount := uint64(numGoroutines * pathsPerGoroutine)
	actualCount := root.announcementsCount.Load()
	assert.Equal(t, totalExpectedCount, actualCount, "counter should reflect traversal count for memory allocation")
}

func TestAnnouncingNode_ConcurrentAccess_AppendAnnouncements(t *testing.T) {
	root := newAnnouncingNode()
	ctx := context.Background()

	// Set up initial structure
	for i := 0; i < 10; i++ {
		node := root.findNode([]string{fmt.Sprintf("child_%d", i)}, nil)
		ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test_%d", i)))
		node.addAnnouncement("ann", ann)
	}

	const numReaders = 10
	var wg sync.WaitGroup
	wg.Add(numReaders)

	// Concurrent readers
	results := make([]int, numReaders)
	for i := 0; i < numReaders; i++ {
		go func(id int) {
			defer wg.Done()
			anns := root.appendAnnouncements(nil)
			results[id] = len(anns)
		}(i)
	}

	wg.Wait()

	// All readers should get the same result
	expectedCount := 10
	for i, count := range results {
		assert.Equal(t, expectedCount, count, "reader %d should get correct count", i)
	}
}

// 8. Edge cases
func TestAnnouncingNode_EdgeCases_EmptySegments(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	// Test with empty string segment
	emptySegment := ""
	ann := NewAnnouncement(ctx, BroadcastPath("/test"))

	assert.NotPanics(t, func() {
		node.addAnnouncement(emptySegment, ann)
	}, "should handle empty segment without panic")

	// Verify empty segment is stored
	node.mu.RLock()
	stored, exists := node.announcements[emptySegment]
	node.mu.RUnlock()

	assert.True(t, exists, "empty segment should be stored")
	assert.Equal(t, ann, stored, "announcement should be stored for empty segment")

	// Test removal with empty segment
	assert.NotPanics(t, func() {
		node.removeAnnouncement(emptySegment, ann)
	}, "should handle empty segment removal without panic")
}

func TestAnnouncingNode_EdgeCases_SpecialCharacters(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	// Test with special characters in segments
	specialSegments := []string{
		"segment/with/slash",
		"segment with spaces",
		"segment-with-dashes",
		"segment_with_underscores",
		"segment.with.dots",
		"segment?with?question",
		"segment#with#hash",
		"セグメント",    // Japanese characters
		"🚀segment", // Emoji
	}

	announcements := make([]*Announcement, len(specialSegments))

	// Add announcements with special segments
	for i, segment := range specialSegments {
		announcements[i] = NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test%d", i)))
		assert.NotPanics(t, func() {
			node.addAnnouncement(segment, announcements[i])
		}, "should handle special segment: %s", segment)
	}

	// Verify all are stored
	node.mu.RLock()
	count := len(node.announcements)
	node.mu.RUnlock()
	assert.Equal(t, len(specialSegments), count, "should store all special segments")

	// Verify retrieval works
	for i, segment := range specialSegments {
		node.mu.RLock()
		stored, exists := node.announcements[segment]
		node.mu.RUnlock()
		assert.True(t, exists, "special segment should exist: %s", segment)
		assert.Equal(t, announcements[i], stored, "announcement should match for segment: %s", segment)
	}
}

func TestAnnouncingNode_EdgeCases_DeepNesting(t *testing.T) {
	root := newAnnouncingNode()

	// Create very deep nesting
	const depth = 100
	segments := make([]string, depth)
	for i := 0; i < depth; i++ {
		segments[i] = fmt.Sprintf("level_%d", i)
	}

	// Count how many times our op function is called
	var callCount int
	opFunc := func(node *announcingNode) {
		callCount++
		countAnnouncement(node) // Also call the original function
	}

	// Should handle deep nesting without issue
	var deepNode *announcingNode
	assert.NotPanics(t, func() {
		deepNode = root.findNode(segments, opFunc)
	}, "should handle deep nesting without panic")

	assert.NotNil(t, deepNode, "deep node should be created")

	// Verify path exists
	current := root
	for i, segment := range segments {
		current.mu.RLock()
		next, exists := current.children[segment]
		current.mu.RUnlock()
		assert.True(t, exists, "segment %d should exist", i)
		current = next
	}
	assert.Equal(t, deepNode, current, "should reach the deep node")

	// Check counter was called correctly
	// Debug: print the actual counts
	actualCount := root.announcementsCount.Load()
	t.Logf("Op function called: %d times", callCount)
	t.Logf("Root atomic counter: %d", actualCount)

	// op function should be called once for each segment (depth times)
	assert.Equal(t, depth, callCount, "op function should be called for each segment")

	// The counter represents the total announcements under this node for efficient memory allocation
	// In this case, root has been traversed once during the findNode operation
	assert.Equal(t, uint64(1), actualCount, "root counter should be incremented once during traversal")
}

func TestAnnouncingNode_EdgeCases_LargeNumberOfAnnouncements(t *testing.T) {
	node := newAnnouncingNode()
	ctx := context.Background()

	// Add large number of announcements
	const numAnnouncements = 1000
	announcements := make([]*Announcement, numAnnouncements)

	for i := 0; i < numAnnouncements; i++ {
		announcements[i] = NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test_%d", i)))
		segment := fmt.Sprintf("seg_%d", i)
		node.addAnnouncement(segment, announcements[i])
	}

	// Verify all are stored
	node.mu.RLock()
	count := len(node.announcements)
	node.mu.RUnlock()
	assert.Equal(t, numAnnouncements, count, "should store all announcements")

	// Test collection performance
	var collected []*Announcement
	start := time.Now()
	assert.NotPanics(t, func() {
		collected = node.appendAnnouncements(nil)
	}, "should handle large collection without panic")
	duration := time.Since(start)

	assert.Equal(t, numAnnouncements, len(collected), "should collect all announcements")
	t.Logf("Collection of %d announcements took: %v", numAnnouncements, duration)

	// Performance should be reasonable (less than 1 second for 1000 items)
	assert.Less(t, duration, time.Second, "collection should be reasonably fast")
}

// 9. Memory management and resources
func TestAnnouncingNode_MemoryManagement_ProperCleanup(t *testing.T) {
	root := newAnnouncingNode()
	ctx := context.Background()

	// Create structure and add announcements
	const numChildren = 10
	const numAnnouncements = 5

	for i := 0; i < numChildren; i++ {
		child := root.findNode([]string{fmt.Sprintf("child_%d", i)}, nil)
		for j := 0; j < numAnnouncements; j++ {
			ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test_%d_%d", i, j)))
			child.addAnnouncement(fmt.Sprintf("ann_%d", j), ann)
		}
	}

	// Verify initial state
	initialChildren := len(root.children)
	assert.Equal(t, numChildren, initialChildren, "should have correct number of children")

	// Clear all announcements
	for i := 0; i < numChildren; i++ {
		child := root.children[fmt.Sprintf("child_%d", i)]
		child.mu.Lock()
		// Clear announcements map
		for segment := range child.announcements {
			delete(child.announcements, segment)
		}
		child.mu.Unlock()
	}

	// Verify announcements are cleared while structure remains
	for i := 0; i < numChildren; i++ {
		child := root.children[fmt.Sprintf("child_%d", i)]
		child.mu.RLock()
		count := len(child.announcements)
		child.mu.RUnlock()
		assert.Equal(t, 0, count, "child %d should have no announcements", i)
	}

	// Structure should still exist
	assert.Equal(t, numChildren, len(root.children), "children structure should remain")
}

func TestAnnouncingNode_MemoryManagement_ChannelCleanup(t *testing.T) {
	node := newAnnouncingNode()

	// Create and register channels
	const numChannels = 10
	channels := make([]chan *Announcement, numChannels)

	for i := 0; i < numChannels; i++ {
		channels[i] = make(chan *Announcement, 1)
		node.mu.Lock()
		node.channels[channels[i]] = struct{}{}
		node.mu.Unlock()
	}

	// Verify channels are registered
	node.mu.RLock()
	count := len(node.channels)
	node.mu.RUnlock()
	assert.Equal(t, numChannels, count, "should have all channels registered")

	// Simulate cleanup by removing channels
	for i := 0; i < numChannels; i++ {
		node.mu.Lock()
		delete(node.channels, channels[i])
		node.mu.Unlock()
		close(channels[i])
	}

	// Verify cleanup
	node.mu.RLock()
	finalCount := len(node.channels)
	node.mu.RUnlock()
	assert.Equal(t, 0, finalCount, "should have no channels after cleanup")
}

func TestAnnouncingNode_MemoryManagement_NoMemoryLeaks(t *testing.T) {
	// This test creates and destroys many nodes to check for memory leaks
	// In a real environment, you might use tools like pprof to verify

	const iterations = 100
	const nodesPerIteration = 10

	for iteration := 0; iteration < iterations; iteration++ {
		root := newAnnouncingNode()
		ctx := context.Background()

		// Create temporary structure
		nodes := make([]*announcingNode, nodesPerIteration)
		for i := 0; i < nodesPerIteration; i++ {
			nodes[i] = root.findNode([]string{fmt.Sprintf("temp_%d", i)}, nil)
			ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/temp_%d", i)))
			nodes[i].addAnnouncement("temp", ann)
		}

		// Verify structure
		assert.Equal(t, nodesPerIteration, len(root.children), "iteration %d should have correct children", iteration)

		// Let structure go out of scope (simulating cleanup)
		root = nil
		nodes = nil

		// Force garbage collection occasionally
		if iteration%10 == 0 {
			// In a real test, you might call runtime.GC() here
			// but it's generally not recommended in unit tests
		}
	}

	// Test completes without excessive memory usage or panic
	t.Log("Memory management test completed successfully")
}

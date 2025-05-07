// Package moqt_test provides tests for the moqt package.
package moqt_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	"github.com/stretchr/testify/assert"
)

// Define a custom type for context keys
type contextKey string

// TestTrackMuxBasicRouting tests basic routing functionality
func TestTrackMuxBasicRouting(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Create handlers and register them to paths
	audioHandler := &moqt.MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			slog.Info("Getting audio info of audio track",
				"track_path", path,
			)
			return moqt.Info{}, nil
		},
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			slog.Info("Serving audio track",
				"track_path", w.TrackPath(),
				"config", config,
			)
		},
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			slog.Info("Serving audio announcements",
				"config", config,
			)
		},
	}

	// Create another handler for video
	videoHandler := &moqt.MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			slog.Info("Getting video info")
			return moqt.Info{
				TrackPriority:       0,
				LatestGroupSequence: 0,
				GroupOrder:          0,
			}, nil
		},
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			slog.Info("Serving video track",
				"track_path", w.TrackPath(),
				"config", config.String(),
			)
		},
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			slog.Info("Serving video announcements",
				"config", config.String(),
			)
		},
	}

	// Register handlers using context
	ctx := context.Background()
	mux.Handle(ctx, "/tracks/audio", audioHandler)
	mux.Handle(ctx, "/tracks/video", videoHandler)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Verify request to audio track
		audioWriter := &moqt.MockTrackWriter{
			PathValue: "/tracks/audio",
			OpenGroupFunc: func(seq moqt.GroupSequence) (moqt.GroupWriter, error) {
				slog.Info("Opening group for audio track",
					"group_sequence", seq,
				)
				return nil, nil
			},
			CloseFunc: func() error {
				slog.Info("Closing audio track")
				return nil
			},
			CloseWithErrorFunc: func(err error) error {
				slog.Info("Closing audio track with error",
					"error", err,
				)
				return nil
			},
		}
		mux.ServeTrack(audioWriter, &moqt.SubscribeConfig{})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Verify request to video track
		videoWriter := &moqt.MockTrackWriter{
			PathValue: "/tracks/video",
			OpenGroupFunc: func(seq moqt.GroupSequence) (moqt.GroupWriter, error) {
				slog.Info("Opening group for video track",
					"group_sequence", seq,
				)
				return nil, nil
			},
			CloseFunc: func() error {
				slog.Info("Closing video track")
				return nil
			},
			CloseWithErrorFunc: func(err error) error {
				slog.Info("Closing video track with error",
					"error", err,
				)
				return nil
			},
		}
		mux.ServeTrack(videoWriter, &moqt.SubscribeConfig{})
	}()
	notFoundCalled := false
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled = true
		},
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Verify request to non-existent path
		unknownWriter := &moqt.MockTrackWriter{
			PathValue: "/tracks/unknown",
			OpenGroupFunc: func(seq moqt.GroupSequence) (moqt.GroupWriter, error) {
				slog.Info("Opening group for unknown track",
					"group_sequence", seq,
				)
				return nil, nil
			},
			CloseFunc: func() error {
				slog.Info("Closing unknown track")
				return nil
			},
			CloseWithErrorFunc: func(err error) error {
				slog.Info("Closing unknown track with error",
					"error", err,
				)
				return nil
			},
		}

		mux.ServeTrack(unknownWriter, &moqt.SubscribeConfig{})
	}()

	wg.Wait()

	if !notFoundCalled {
		t.Errorf("NotFoundHandler was not called for non-existent path")
	}
}

// TestGetInfo tests the GetInfo functionality
func TestGetInfo(t *testing.T) {
	mux := moqt.NewTrackMux()

	expectedInfo := moqt.Info{
		TrackPriority:       10,
		LatestGroupSequence: 100,
		GroupOrder:          1,
	}

	audioHandler := &moqt.MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			return expectedInfo, nil
		},
	}

	// Add context
	mux.Handle(context.Background(), "/tracks/audio", audioHandler)

	// Test getting information for an existing track
	info, err := mux.GetInfo("/tracks/audio")
	if err != nil {
		t.Errorf("GetInfo returned error: %v", err)
	}

	if info.TrackPriority != expectedInfo.TrackPriority {
		t.Errorf("Expected TrackPriority %d, got %d", expectedInfo.TrackPriority, info.TrackPriority)
	}

	// Test getting information for a non-existent track
	_, err = mux.GetInfo("/tracks/unknown")
	if err == nil || !errors.Is(err, moqt.ErrTrackDoesNotExist) {
		t.Errorf("Expected ErrTrackDoesNotExist for non-existent track, got: %v", err)
	}
}

// TestAnnouncements tests the announcement system
func TestAnnouncements(t *testing.T) {
	mux := moqt.NewTrackMux()

	// First, register announcement subscribers
	audioAnnounced := 0
	audioAnnouncer := &moqt.MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*moqt.Announcement) error {
			slog.Info("Sending announcements to audio announcer",
				"announcements", announcements,
			)
			audioAnnounced += len(announcements)
			return nil
		},
	}
	audioAnnounceConfig := &moqt.AnnounceConfig{
		TrackPattern: "/tracks/audio/*", // Single segments under audio
	}

	videoAnnounced := 0
	videoAnnouncer := &moqt.MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*moqt.Announcement) error {
			slog.Info("Sending announcements to audio announcer",
				"announcements", announcements,
			)
			videoAnnounced += len(announcements)
			return nil
		},
	}
	videoAnnounceConfig := &moqt.AnnounceConfig{
		TrackPattern: "/tracks/video/**", // All segments under video
	}

	allTracksAnnounced := 0
	allTracksAnnouncer := &moqt.MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*moqt.Announcement) error {
			slog.Info("Sending announcements to audio announcer",
				"announcements", announcements,
			)
			allTracksAnnounced += len(announcements)
			return nil
		},
	}
	allTracksAnnounceConfig := &moqt.AnnounceConfig{
		TrackPattern: "/**", // All tracks
	}

	go mux.ServeAnnouncements(audioAnnouncer, audioAnnounceConfig)
	go mux.ServeAnnouncements(videoAnnouncer, videoAnnounceConfig)
	go mux.ServeAnnouncements(allTracksAnnouncer, allTracksAnnounceConfig)

	// Verify that announcements occur when handlers are registered
	audioHandler := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/audio/main")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	videoHandler := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/video/main")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	nestedVideoHandler := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/video/streams/hd")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	otherHandler := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/other/track")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	// Add context
	ctx := context.Background()
	mux.Handle(ctx, "/tracks/audio/main", audioHandler)
	mux.Handle(ctx, "/tracks/video/main", videoHandler)
	mux.Handle(ctx, "/tracks/video/streams/hd", nestedVideoHandler)
	mux.Handle(ctx, "/other/track", otherHandler)

	// Verify that each announcer received tracks matching their pattern
	if audioAnnounced != 1 {
		t.Errorf("Audio announcer should receive 1 notification, got %d", audioAnnounced)
	}

	if videoAnnounced != 2 {
		t.Errorf("Video announcer should receive 2 notifications, got %d", videoAnnounced)
	}

	if allTracksAnnounced != 4 {
		t.Errorf("All tracks announcer should receive 4 notifications, got %d", allTracksAnnounced)
	}
}

// TestHandlerOverwrite tests handler overwriting
func TestTrackMux_HandlerOverwrite(t *testing.T) {
	mux := moqt.NewTrackMux()

	handler1 := &moqt.MockTrackHandler{}
	handler2 := &moqt.MockTrackHandler{}

	// Register handler using context
	ctx := context.Background()
	mux.Handle(ctx, "/tracks/test", handler1)

	// Register a different handler to the same path (a warning should be logged)
	mux.Handle(ctx, "/tracks/test", handler2)

	// Verify that handler2 is used
	writer := &moqt.MockTrackWriter{PathValue: "/tracks/test"}
	mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	// Add check to verify that context for handler1 was cancelled
}

// TestConcurrentAccess tests the safety of concurrent access
func TestTrackMux_ConcurrentAccess(t *testing.T) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Set up multiple handlers
	for i := 0; i < 10; i++ {
		path := moqt.TrackPath("/tracks/path" + string(rune(i+'0')))
		handler := &moqt.MockTrackHandler{
			ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
				// Do nothing
			},
			ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
				// Do nothing
			},
			GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
				return moqt.Info{}, nil
			},
		}
		mux.Handle(ctx, path, handler)
	}

	var wg sync.WaitGroup

	// Read concurrently with 10 goroutines
	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := moqt.TrackPath("/tracks/path" + string(rune(idx+'0')))
			writer := &moqt.MockTrackWriter{PathValue: path}
			mux.ServeTrack(writer, &moqt.SubscribeConfig{})
			mux.GetInfo(path)
		}(i)
	}

	// Write concurrently with 5 more goroutines
	for i := 10; i < 15; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := moqt.TrackPath("/tracks/path" + string(rune(idx+'0')))
			handler := &moqt.MockTrackHandler{
				ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
					// Do nothing
				},
				ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
					// Do nothing
				},
				GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
					return moqt.Info{}, nil
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			mux.Handle(ctx, path, handler)
		}(i)
	}

	wg.Wait()
	// If we get here, concurrent access is working without deadlocks
}

// TestDefaultMux tests the functionality of the global DefaultMux
func TestDefaultMux(t *testing.T) {
	// Temporarily save the original DefaultMux for reset
	origDefaultMux := moqt.DefaultMux
	defer func() {
		moqt.DefaultMux = origDefaultMux
	}()

	moqt.DefaultMux = moqt.NewTrackMux()

	// Register a handler to the default Mux
	handler := &moqt.MockTrackHandler{}
	ctx := context.Background()
	moqt.Handle(ctx, "/default/test", handler)

	// Verify that the handler is called correctly
	writer := &moqt.MockTrackWriter{PathValue: "/default/test"}
	moqt.ServeTrack(writer, &moqt.SubscribeConfig{})

	// Also test GetInfo
	info, err := moqt.GetInfo("/default/test")
	if err != nil {
		t.Errorf("GetInfo via DefaultMux returned error: %v", err)
	}

	// Verify the content of the Info
	if info.TrackPriority != 1 {
		t.Errorf("Expected default TrackPriority 1, got %d", info.TrackPriority)
	}
}

// TestMuxHandler tests that TrackMux correctly implements the TrackHandler interface
func TestTrackMux_HandledByInnerMux(t *testing.T) {
	// Verify that TrackMux implements the TrackHandler interface
	var _ moqt.Handler = (*moqt.TrackMux)(nil)

	// Register another TrackMux as a handler
	outerMux := moqt.NewTrackMux()
	innerMux := moqt.NewTrackMux()
	ctx := context.Background()

	// Register a handler to the inner Mux
	innerHandler := &moqt.MockTrackHandler{}
	innerMux.Handle(ctx, "/inner/track", innerHandler)

	// Register the inner Mux as a handler for the outer Mux
	outerMux.Handle(ctx, "/outer", innerMux)

	// Test request to the inner handler
	writer := &moqt.MockTrackWriter{PathValue: "/outer/inner/track"}
	outerMux.ServeTrack(writer, &moqt.SubscribeConfig{})
}

// TestNestedRouting tests routing for deeply nested paths
func TestTrackMux_NestedRouting(t *testing.T) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Register a handler to a deeply nested path
	deepHandler := &moqt.MockTrackHandler{}
	mux.Handle(ctx, "/a/b/c/d/e/f", deepHandler)

	// Test request to the exact path
	writer := &moqt.MockTrackWriter{PathValue: "/a/b/c/d/e/f"}
	go mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	// Test request to a partial path (NotFoundHandler should be called)
	partialWriter := &moqt.MockTrackWriter{PathValue: "/a/b/c"}
	notFoundCalled := false
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled = true
		},
	}

	mux.ServeTrack(partialWriter, &moqt.SubscribeConfig{})

	if !notFoundCalled {
		t.Errorf("NotFoundHandler was not called for partial path match")
	}
}

// TestServeAnnouncementWithDefaultMux tests announcements using DefaultMux
func TestTrackMux_ServeAnnouncement(t *testing.T) {
	mux := moqt.NewTrackMux()

	expectedAnnouncements := []*moqt.Announcement{
		moqt.NewAnnouncement("/global/track"),
	}

	// Register an announcement subscriber
	announcedCh := make(chan []*moqt.Announcement, 1)
	defer close(announcedCh)

	announcer := &moqt.MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*moqt.Announcement) error {
			slog.Info("Sending announcements to announcer",
				"announcements", announcements,
			)
			announcedCh <- announcements
			return nil
		},
	}

	go mux.ServeAnnouncements(announcer, &moqt.AnnounceConfig{TrackPattern: "/**"})

	// Register a handler and verify that announcements occur
	handler := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			slog.Info("Sending announcements from handler",
				"announcements", expectedAnnouncements,
			)
			w.SendAnnouncements(expectedAnnouncements)
		},
	}

	mux.Handle(context.Background(), "/global/track", handler)
	select {
	case announced := <-announcedCh:
		assert.NotNil(t, announced)
		assert.Equal(t, expectedAnnouncements, announced)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for announcements")
	}
}

// TestGetInfoImplementation tests the GetInfo implementation in detail
func TestTrackMux_GetInfo(t *testing.T) {
	mux := moqt.NewTrackMux()
	ctx := context.Background()

	// Register multiple handlers with different information
	audioInfo := moqt.Info{
		TrackPriority:       10,
		LatestGroupSequence: 100,
		GroupOrder:          1,
	}

	videoInfo := moqt.Info{
		TrackPriority:       20,
		LatestGroupSequence: 200,
		GroupOrder:          2,
	}

	audioHandler := &moqt.MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			return audioInfo, nil
		},
	}

	videoHandler := &moqt.MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			return videoInfo, nil
		},
	}

	mux.Handle(ctx, "/media/audio", audioHandler)
	mux.Handle(ctx, "/media/video", videoHandler)

	// Get and verify information for the audio track
	info, err := mux.GetInfo("/media/audio")
	if err != nil {
		t.Errorf("GetInfo returned error for audio track: %v", err)
	}

	if info.TrackPriority != audioInfo.TrackPriority {
		t.Errorf("Expected TrackPriority %d, got %d", audioInfo.TrackPriority, info.TrackPriority)
	}

	if info.LatestGroupSequence != audioInfo.LatestGroupSequence {
		t.Errorf("Expected LatestGroupSequence %d, got %d", audioInfo.LatestGroupSequence, info.LatestGroupSequence)
	}

	if info.GroupOrder != audioInfo.GroupOrder {
		t.Errorf("Expected GroupOrder %d, got %d", audioInfo.GroupOrder, info.GroupOrder)
	}

	// Get and verify information for the video track
	info, err = mux.GetInfo("/media/video")
	if err != nil {
		t.Errorf("GetInfo returned error for video track: %v", err)
	}

	if info.TrackPriority != videoInfo.TrackPriority {
		t.Errorf("Expected TrackPriority %d, got %d", videoInfo.TrackPriority, info.TrackPriority)
	}
}

// TestContextLifecycleManagement tests context-based handler lifecycle management
// including both timeout-based and manual cancellation scenarios
func TestTrackMux_ContextLifecycleManagement(t *testing.T) {
	t.Run("automatic timeout-based cancellation", func(t *testing.T) {
		mux := moqt.NewTrackMux()

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel() // To prevent leaks in case of early returns

		handler := &moqt.MockTrackHandler{}
		mux.Handle(ctx, "/temporary/resource", handler)

		// Verify handler was registered
		writer := &moqt.MockTrackWriter{PathValue: "/temporary/resource"}
		go mux.ServeTrack(writer, &moqt.SubscribeConfig{})

		// Wait for context timeout
		time.Sleep(100 * time.Millisecond)

		// Verify handler was removed
		notFoundCalled := false
		origNotFoundHandler := moqt.NotFoundHandler
		defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

		moqt.NotFoundHandler = &moqt.MockTrackHandler{
			ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
				notFoundCalled = true
			},
		}

		// Wait for cleanup to complete
		mux.ServeTrack(writer, &moqt.SubscribeConfig{})

		if !notFoundCalled {
			t.Errorf("Handler should have been removed after context timeout")
		}
	})

	t.Run("manual context cancellation", func(t *testing.T) {
		mux := moqt.NewTrackMux()

		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())

		handler := &moqt.MockTrackHandler{}
		mux.Handle(ctx, "/cancellable/resource", handler)

		// Verify handler was registered
		writer := &moqt.MockTrackWriter{PathValue: "/cancellable/resource"}
		mux.ServeTrack(writer, &moqt.SubscribeConfig{})

		// Manually cancel the context
		cancel()

		// Wait for cancellation processing to complete
		time.Sleep(10 * time.Millisecond)

		// Verify handler was removed
		notFoundCalled := false
		origNotFoundHandler := moqt.NotFoundHandler
		defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

		moqt.NotFoundHandler = &moqt.MockTrackHandler{
			ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
				notFoundCalled = true
			},
		}

		mux.ServeTrack(writer, &moqt.SubscribeConfig{})

		if !notFoundCalled {
			t.Errorf("Handler should have been removed after manual context cancellation")
		}
	})

	t.Run("dual mode handling", func(t *testing.T) {
		mux := moqt.NewTrackMux()

		// Set up handler for path
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		handler := &moqt.MockTrackHandler{}
		path := moqt.TrackPath("/dual/mode/path")

		mux.Handle(ctx, path, handler)

		// Verify handler is correctly registered and can serve requests
		writer := &moqt.MockTrackWriter{PathValue: path}
		called := false
		handler.ServeTrackFunc = func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			called = true
		}

		mux.ServeTrack(writer, &moqt.SubscribeConfig{})

		if !called {
			t.Error("Handler should have been called")
		}

		// Cancel context and verify handler is removed
		cancel()
		time.Sleep(10 * time.Millisecond) // Wait for cancellation processing

		notFoundCalled := false
		origNotFoundHandler := moqt.NotFoundHandler
		defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

		moqt.NotFoundHandler = &moqt.MockTrackHandler{
			ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
				notFoundCalled = true
			},
		}

		mux.ServeTrack(writer, &moqt.SubscribeConfig{})

		if !notFoundCalled {
			t.Errorf("NotFoundHandler should have been called after context cancellation")
		}
	})
}

// TestHandlerOverwriteWithContext tests handler overwriting with context management
func TestTrackMux_HandlerOverwriteWithContext(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Prepare first handler and its context
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	handler1 := &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			// Do nothing specific for handler1
		},
	}

	// Prepare second handler and its context
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	handler2 := &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			// Do nothing specific for handler2
		},
	}

	// Register first handler
	mux.Handle(ctx1, "/overwrite/test", handler1)

	// Verify handler1 was registered
	writer := &moqt.MockTrackWriter{PathValue: "/overwrite/test"}
	mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	// Register a different handler to the same path (overwrite)
	mux.Handle(ctx2, "/overwrite/test", handler2)

	// Check if first context was cancelled (wait needed)
	time.Sleep(10 * time.Millisecond)

	// Cancel context2
	cancel2()

	// Wait for cancellation processing to complete
	time.Sleep(10 * time.Millisecond)

	// Verify handler was removed
	notFoundCalled := false
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled = true
		},
	}

	mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	if !notFoundCalled {
		t.Errorf("Handler should have been removed after context cancellation")
	}
}

// TestHierarchicalContextManagement tests hierarchical context management
func TestTrackMux_HierarchicalContextManagement(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Create parent context
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	// Create child contexts (derived from parent)
	childCtx1 := context.WithValue(parentCtx, contextKey("key"), "child1")
	childCtx2 := context.WithValue(parentCtx, contextKey("key"), "child2")

	// Register multiple handlers
	handler1 := &moqt.MockTrackHandler{}
	handler2 := &moqt.MockTrackHandler{}
	handler3 := &moqt.MockTrackHandler{}

	mux.Handle(childCtx1, "/hierarchy/resource1", handler1)
	mux.Handle(childCtx2, "/hierarchy/resource2", handler2)
	mux.Handle(parentCtx, "/hierarchy/parent", handler3)

	// Verify all handlers were registered
	writer1 := &moqt.MockTrackWriter{PathValue: "/hierarchy/resource1"}
	writer2 := &moqt.MockTrackWriter{PathValue: "/hierarchy/resource2"}
	writer3 := &moqt.MockTrackWriter{PathValue: "/hierarchy/parent"}

	go mux.ServeTrack(writer1, &moqt.SubscribeConfig{})
	go mux.ServeTrack(writer2, &moqt.SubscribeConfig{})
	go mux.ServeTrack(writer3, &moqt.SubscribeConfig{})

	// Cancel parent context (all children auto-cancel)
	parentCancel()

	// Wait for cancellation processing to complete
	time.Sleep(10 * time.Millisecond)

	// Verify all handlers were removed
	notFoundCalled := 0
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled++
		},
	}

	mux.ServeTrack(writer1, &moqt.SubscribeConfig{})
	mux.ServeTrack(writer2, &moqt.SubscribeConfig{})
	mux.ServeTrack(writer3, &moqt.SubscribeConfig{})

	if notFoundCalled != 3 {
		t.Errorf("Expected all 3 handlers to be removed after parent context cancellation, but %d were removed", notFoundCalled)
	}
}

// TestNodeCleanupAfterCancellation tests that nodes are properly cleaned up after context cancellation
func TestTrackMux_NodeCleanupAfterCancellation(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Register handler for a deeply nested path
	ctx, cancel := context.WithCancel(context.Background())
	handler := &moqt.MockTrackHandler{}

	// Register the deeply nested path
	mux.Handle(ctx, "/deep/path/to/resource", handler)

	// Verify handler was registered
	writer := &moqt.MockTrackWriter{PathValue: "/deep/path/to/resource"}
	go mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	// Cancel the context
	cancel()

	// Wait for cleanup to complete
	time.Sleep(10 * time.Millisecond)

	// Verify handler was removed
	notFoundCalled := false
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled = true
		},
	}

	mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	if !notFoundCalled {
		t.Errorf("Handler should have been removed after context cancellation")
	}

	// Verify intermediate path is also inaccessible (confirming complete node cleanup)
	intermediateWriter := &moqt.MockTrackWriter{PathValue: "/deep/path"}
	intermediateNotFoundCalled := false

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			intermediateNotFoundCalled = true
		},
	}

	mux.ServeTrack(intermediateWriter, &moqt.SubscribeConfig{})

	if !intermediateNotFoundCalled {
		t.Errorf("Intermediate node should have been removed after context cancellation")
	}
}

// TestConcurrentContextCancellation tests concurrent cancellation of multiple contexts
func TestTrackMux_ConcurrentContextCancellation(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Test concurrent registration and cancellation of multiple handlers
	const handlerCount = 50
	var wg sync.WaitGroup

	for i := range handlerCount {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			path := moqt.TrackPath(fmt.Sprintf("/concurrent/path%d", idx))
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20+idx)*time.Millisecond)
			defer cancel()

			handler := &moqt.MockTrackHandler{}
			mux.Handle(ctx, path, handler)

			// Verify handler was registered
			writer := &moqt.MockTrackWriter{PathValue: path}
			mux.ServeTrack(writer, &moqt.SubscribeConfig{})
		}(i)
	}

	wg.Wait()

	// Wait for all timeouts to occur
	time.Sleep(100 * time.Millisecond)

	// Verify all handlers were removed
	removedCount := 0
	var checkWg sync.WaitGroup

	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	var mu sync.Mutex // Mutex for thread-safe counter access

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			mu.Lock()
			removedCount++
			mu.Unlock()
		},
	}

	for i := range handlerCount {
		checkWg.Add(1)
		go func(idx int) {
			defer checkWg.Done()
			path := moqt.TrackPath(fmt.Sprintf("/concurrent/path%d", idx))
			writer := &moqt.MockTrackWriter{PathValue: path}
			mux.ServeTrack(writer, &moqt.SubscribeConfig{})
		}(i)
	}

	checkWg.Wait()

	if removedCount != handlerCount {
		t.Errorf("Expected %d handlers to be removed after context cancellation, but got %d", handlerCount, removedCount)
	}
}

// TestMultipleHandlersWithSingleContext tests multiple handlers with a single context
func TestTrackMux_MultipleHandlersWithSingleContext(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Share a single context among multiple handlers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create multiple handlers
	handler1 := &moqt.MockTrackHandler{}
	handler2 := &moqt.MockTrackHandler{}
	handler3 := &moqt.MockTrackHandler{}

	// Register multiple paths using the same context
	mux.Handle(ctx, "/multi/path1", handler1)
	mux.Handle(ctx, "/multi/path2", handler2)
	mux.Handle(ctx, "/multi/path3", handler3)

	// Verify all paths are working
	writer1 := &moqt.MockTrackWriter{PathValue: "/multi/path1"}
	writer2 := &moqt.MockTrackWriter{PathValue: "/multi/path2"}
	writer3 := &moqt.MockTrackWriter{PathValue: "/multi/path3"}

	called1, called2, called3 := false, false, false

	handler1.ServeTrackFunc = func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
		called1 = true
	}
	handler2.ServeTrackFunc = func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
		called2 = true
	}
	handler3.ServeTrackFunc = func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
		called3 = true
	}

	mux.ServeTrack(writer1, &moqt.SubscribeConfig{})
	mux.ServeTrack(writer2, &moqt.SubscribeConfig{})
	mux.ServeTrack(writer3, &moqt.SubscribeConfig{})

	if !called1 || !called2 || !called3 {
		t.Errorf("All handlers should have been called: %v, %v, %v", called1, called2, called3)
	}

	// Cancel context
	cancel()
	time.Sleep(10 * time.Millisecond)

	// Verify all handlers were removed
	notFoundCalled := 0
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &moqt.MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled++
		},
	}

	mux.ServeTrack(writer1, &moqt.SubscribeConfig{})
	mux.ServeTrack(writer2, &moqt.SubscribeConfig{})
	mux.ServeTrack(writer3, &moqt.SubscribeConfig{})

	if notFoundCalled != 3 {
		t.Errorf("All handlers should have been removed, but got %d removals", notFoundCalled)
	}
}

// TestSingleWildcardAnnounce tests announcements for single wildcard patterns
func TestTrackMux_SingleWildcardAnnounce(t *testing.T) {
	mux := moqt.NewTrackMux()

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Register an announcement subscriber
	singleWildcardAnnounced := 0
	singleWildcardAnnouncer := &moqt.MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*moqt.Announcement) error {
			mu.Lock()
			defer mu.Unlock()

			defer wg.Done()

			singleWildcardAnnounced += len(announcements)

			return nil
		},
	}
	singleWildcardConfig := &moqt.AnnounceConfig{
		TrackPattern: "/tracks/audio/*", // Single segments under audio
	}

	// Start serving announcements in a separate goroutine
	go mux.ServeAnnouncements(singleWildcardAnnouncer, singleWildcardConfig)

	// Register handlers and verify that announcements occur
	handler1 := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/audio/main")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	handler2 := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/audio/sub")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	wg.Add(1) // Increment WaitGroup for one handler
	mux.Handle(context.Background(), "/tracks/audio/main", handler1)

	wg.Add(1) // Increment WaitGroup for one handler
	mux.Handle(context.Background(), "/tracks/audio/sub", handler2)

	// Wait for all announcements to be processed
	wg.Wait()

	// Verify that the announcer received the notification
	if singleWildcardAnnounced != 2 {
		t.Errorf("Single wildcard announcer should receive 2 notifications, got %d", singleWildcardAnnounced)
	}
}

// TestDoubleWildcardAnnounce verifies that the double wildcard pattern (**) correctly
// matches and delivers announcements for all paths under the specified prefix.
func TestTrackMux_DoubleWildcardAnnounce(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Create a context that can be cancelled when the test completes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Clean up resources when the test is done

	wg := &sync.WaitGroup{}
	var mu sync.Mutex // Mutex to protect access to doubleWildcardAnnounced
	doubleWildcardAnnounced := 0

	// Mock announcement writer for receiving announcements
	doubleWildcardAnnouncer := &moqt.MockAnnouncementWriter{
		SendAnnouncementsFunc: func(announcements []*moqt.Announcement) error {
			slog.Info("received announcements",
				"count", len(announcements),
			)

			mu.Lock()
			defer mu.Unlock()

			doubleWildcardAnnounced += len(announcements)

			wg.Done()
			return nil
		},
	}

	// Use a double wildcard pattern
	doubleWildcardConfig := &moqt.AnnounceConfig{
		TrackPattern: "/tracks/video/**", // Matches all paths under the video directory
	}

	// Start the announcement service
	go mux.ServeAnnouncements(doubleWildcardAnnouncer, doubleWildcardConfig)

	// Set up test handlers
	handler1 := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			// Announce a top-level video track
			ann := moqt.NewAnnouncement("/tracks/video/main")
			wg.Add(1) // Increment counter before calling SendAnnouncements
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	handler2 := &moqt.MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			// Announce a nested video track
			ann := moqt.NewAnnouncement("/tracks/video/streams/hd")
			wg.Add(1) // Increment counter before calling SendAnnouncements
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	// Register handlers
	mux.Handle(ctx, "/tracks/video/main", handler1)
	mux.Handle(ctx, "/tracks/video/streams/hd", handler2)

	wg.Wait()

	if doubleWildcardAnnounced != 2 {
		t.Errorf("double wildcard announcer should receive exactly 2 notifications, got %d", doubleWildcardAnnounced)
	}
}

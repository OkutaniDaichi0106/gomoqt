package moqt_test

import (
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

// TestTrackMuxBasicRouting tests basic routing functionality
func TestTrackMuxBasicRouting(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Create handlers and register them to paths
	audioHandler := &MockTrackHandler{
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
	videoHandler := &MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			slog.Info("Getting video info")
			return moqt.Info{
				TrackPriority:       0,
				LatestGroupSequence: 0,
				GroupOrder:          0,
			}, nil
		},
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			slog.Info("Serving video track")
		},
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			slog.Info("Serving video announcements")
		},
	}

	mux.Handle("/tracks/audio", audioHandler)
	mux.Handle("/tracks/video", videoHandler)

	// Verify request to audio track
	audioWriter := &MockTrackWriter{PathValue: "/tracks/audio"}
	mux.ServeTrack(audioWriter, &moqt.SubscribeConfig{})

	// Verify request to video track
	videoWriter := &MockTrackWriter{PathValue: "/tracks/video"}
	mux.ServeTrack(videoWriter, &moqt.SubscribeConfig{})

	// Verify request to non-existent path
	unknownWriter := &MockTrackWriter{PathValue: "/tracks/unknown"}
	notFoundCalled := false
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &MockTrackHandler{
		ServeTrackFunc: func(w moqt.TrackWriter, config *moqt.SubscribeConfig) {
			notFoundCalled = true
		},
	}
	mux.ServeTrack(unknownWriter, &moqt.SubscribeConfig{})

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

	audioHandler := &MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			return expectedInfo, nil
		},
	}

	mux.Handle("/tracks/audio", audioHandler)

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
	audioAnnouncer := &MockAnnouncementWriter{
		ConfigValue: moqt.AnnounceConfig{
			TrackPattern: "/tracks/audio/*", // Single segment under audio
		},
	}

	videoAnnouncer := &MockAnnouncementWriter{
		ConfigValue: moqt.AnnounceConfig{
			TrackPattern: "/tracks/video/**", // Multiple segments under video
		},
	}

	allTracksAnnouncer := &MockAnnouncementWriter{
		ConfigValue: moqt.AnnounceConfig{
			TrackPattern: "/**", // All tracks
		},
	}

	mux.ServeAnnouncements(audioAnnouncer, &moqt.AnnounceConfig{})
	mux.ServeAnnouncements(videoAnnouncer, &moqt.AnnounceConfig{})
	mux.ServeAnnouncements(allTracksAnnouncer, &moqt.AnnounceConfig{})

	// Verify that announcements occur when handlers are registered
	audioHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/audio/main")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	videoHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/video/main")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	nestedVideoHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/tracks/video/streams/hd")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	otherHandler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/other/track")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	mux.Handle("/tracks/audio/main", audioHandler)
	mux.Handle("/tracks/video/main", videoHandler)
	mux.Handle("/tracks/video/streams/hd", nestedVideoHandler)
	mux.Handle("/other/track", otherHandler)

	// Verify that each announcer received tracks matching their pattern
	if audioAnnouncer.Notifications != 1 {
		t.Errorf("Audio announcer should receive 1 notification, got %d", audioAnnouncer.Notifications)
	}

	if videoAnnouncer.Notifications != 2 {
		t.Errorf("Video announcer should receive 2 notifications, got %d", videoAnnouncer.Notifications)
	}

	if allTracksAnnouncer.Notifications != 4 {
		t.Errorf("All tracks announcer should receive 4 notifications, got %d", allTracksAnnouncer.Notifications)
	}
}

// TestHandlerOverwrite tests handler overwriting
func TestHandlerOverwrite(t *testing.T) {
	mux := moqt.NewTrackMux()

	handler1 := &MockTrackHandler{}
	handler2 := &MockTrackHandler{}

	mux.Handle("/tracks/test", handler1)

	// Register a different handler to the same path (a warning should be logged)
	mux.Handle("/tracks/test", handler2)

	// Verify that handler2 is used
	writer := &MockTrackWriter{PathValue: "/tracks/test"}
	mux.ServeTrack(writer, &moqt.SubscribeConfig{})

}

// TestConcurrentAccess tests the safety of concurrent access
func TestConcurrentAccess(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Set up multiple handlers
	for i := 0; i < 10; i++ {
		path := moqt.TrackPath("/tracks/path" + string(rune(i+'0')))
		handler := &MockTrackHandler{
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
		mux.Handle(path, handler)
	}

	var wg sync.WaitGroup

	// Read concurrently with 10 goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := moqt.TrackPath("/tracks/path" + string(rune(idx+'0')))
			writer := &MockTrackWriter{PathValue: path}
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
			handler := &MockTrackHandler{
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
			mux.Handle(path, handler)
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
	handler := &MockTrackHandler{}
	moqt.Handle("/default/test", handler)

	// Verify that the handler is called correctly
	writer := &MockTrackWriter{PathValue: "/default/test"}
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

// TestWildcardRouting tests wildcard path matching
func TestWildcardRouting(t *testing.T) {
	mux := moqt.NewTrackMux()

	singleHandler := &MockTrackHandler{}
	doubleHandler := &MockTrackHandler{}

	// Announcer subscribing to single wildcard (*)
	singleWildcardAnnouncer := &MockAnnouncementWriter{
		ConfigValue: moqt.AnnounceConfig{
			TrackPattern: "/wildcard/*",
		},
	}

	// Announcer subscribing to double wildcard (**)
	doubleWildcardAnnouncer := &MockAnnouncementWriter{
		ConfigValue: moqt.AnnounceConfig{
			TrackPattern: "/deep/**",
		},
	}

	mux.ServeAnnouncements(singleWildcardAnnouncer, &moqt.AnnounceConfig{})
	mux.ServeAnnouncements(doubleWildcardAnnouncer, &moqt.AnnounceConfig{})

	// Register handlers corresponding to each pattern
	mux.Handle("/wildcard/one", singleHandler)
	mux.Handle("/deep/one/two/three", doubleHandler)

	// Test single wildcard (*)
	if singleWildcardAnnouncer.Notifications != 1 {
		t.Errorf("Single wildcard announcer should receive 1 notification, got %d", singleWildcardAnnouncer.Notifications)
	}

	// Test double wildcard (**)
	if doubleWildcardAnnouncer.Notifications != 1 {
		t.Errorf("Double wildcard announcer should receive 1 notification, got %d", doubleWildcardAnnouncer.Notifications)
	}
}

// TestMuxHandler tests that TrackMux correctly implements the TrackHandler interface
func TestMuxHandler(t *testing.T) {
	// Verify that TrackMux implements the TrackHandler interface
	var _ moqt.TrackHandler = (*moqt.TrackMux)(nil)

	// Register another TrackMux as a handler
	outerMux := moqt.NewTrackMux()
	innerMux := moqt.NewTrackMux()

	// Register a handler to the inner Mux
	innerHandler := &MockTrackHandler{}
	innerMux.Handle("/inner/track", innerHandler)

	// Register the inner Mux as a handler for the outer Mux
	outerMux.Handle("/outer", innerMux)

	// Test request to the inner handler
	writer := &MockTrackWriter{PathValue: "/outer/inner/track"}
	outerMux.ServeTrack(writer, &moqt.SubscribeConfig{})

}

// TestNestedRouting tests routing for deeply nested paths
func TestNestedRouting(t *testing.T) {
	mux := moqt.NewTrackMux()

	// Register a handler to a deeply nested path
	deepHandler := &MockTrackHandler{}
	mux.Handle("/a/b/c/d/e/f", deepHandler)

	// Test request to the exact path
	writer := &MockTrackWriter{PathValue: "/a/b/c/d/e/f"}
	mux.ServeTrack(writer, &moqt.SubscribeConfig{})

	// Test request to a partial path (NotFoundHandler should be called)
	partialWriter := &MockTrackWriter{PathValue: "/a/b/c"}
	notFoundCalled := false
	origNotFoundHandler := moqt.NotFoundHandler
	defer func() { moqt.NotFoundHandler = origNotFoundHandler }()

	moqt.NotFoundHandler = &MockTrackHandler{
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
func TestServeAnnouncementWithDefaultMux(t *testing.T) {
	// Save the original DefaultMux for later restoration
	origDefaultMux := moqt.DefaultMux
	defer func() {
		moqt.DefaultMux = origDefaultMux
	}()

	moqt.DefaultMux = moqt.NewTrackMux()

	// Register an announcement subscriber
	announcer := &MockAnnouncementWriter{
		ConfigValue: moqt.AnnounceConfig{
			TrackPattern: "/**", // All tracks
		},
	}

	moqt.ServeAnnouncements(announcer, &moqt.AnnounceConfig{})

	// Register a handler and verify that announcements occur
	handler := &MockTrackHandler{
		ServeAnnouncementsFunc: func(w moqt.AnnouncementWriter, config *moqt.AnnounceConfig) {
			ann := moqt.NewAnnouncement("/global/track")
			w.SendAnnouncements([]*moqt.Announcement{ann})
		},
	}

	moqt.Handle("/global/track", handler)

	// Verify that the announcer received the notification
	if announcer.Notifications != 1 {
		t.Errorf("Announcer should receive 1 notification via DefaultMux, got %d", announcer.Notifications)
	}
}

// TestGetInfoImplementation tests the GetInfo implementation in detail
func TestGetInfoImplementation(t *testing.T) {
	mux := moqt.NewTrackMux()

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

	audioHandler := &MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			return audioInfo, nil
		},
	}

	videoHandler := &MockTrackHandler{
		GetInfoFunc: func(path moqt.TrackPath) (moqt.Info, error) {
			return videoInfo, nil
		},
	}

	mux.Handle("/media/audio", audioHandler)
	mux.Handle("/media/video", videoHandler)

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

// BenchmarkPathMatching benchmarks the performance of path matching
func BenchmarkPathMatching(b *testing.B) {
	mux := moqt.NewTrackMux()

	// Register many handlers
	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			path := moqt.TrackPath("/section" + string(rune(i+'0')) + "/subsection" + string(rune(j+'0')))
			mux.Handle(path, &MockTrackHandler{})
		}
	}

	// Test a deeply nested path
	writer := &MockTrackWriter{PathValue: "/section5/subsection7"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.ServeTrack(writer, &moqt.SubscribeConfig{})
	}
}

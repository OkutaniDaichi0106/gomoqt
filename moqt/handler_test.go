package moqt

import (
	"testing"
)

func TestNotFound(t *testing.T) {
	tests := []struct {
		name string
		pub  *Publisher
	}{
		{
			name: "nil publisher",
			pub:  nil,
		},
		{
			name: "publisher with nil TrackWriter",
			pub: &Publisher{
				BroadcastPath: BroadcastPath("/test"),
				TrackName:     TrackName("test"),
				TrackWriter:   nil,
			},
		},
		{
			name: "publisher with mock TrackWriter",
			pub: &Publisher{
				BroadcastPath: BroadcastPath("/test"),
				TrackName:     TrackName("test"),
				TrackWriter:   &MockTrackWriter{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			NotFound(tt.pub)
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	pub := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
		TrackWriter:   &MockTrackWriter{},
	}

	// Should not panic
	NotFoundHandler.ServeTrack(pub)
}

func TestTrackHandlerFunc(t *testing.T) {
	called := false
	var receivedPub *Publisher

	handler := TrackHandlerFunc(func(pub *Publisher) {
		called = true
		receivedPub = pub
	})

	testPub := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
	}

	handler.ServeTrack(testPub)

	if !called {
		t.Error("handler function was not called")
	}

	if receivedPub != testPub {
		t.Error("handler did not receive the correct publisher")
	}
}

func TestTrackHandlerFuncServeTrack(t *testing.T) {
	callCount := 0

	handler := TrackHandlerFunc(func(pub *Publisher) {
		callCount++
	})

	pub := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
	}

	// Call multiple times to ensure it works correctly
	handler.ServeTrack(pub)
	handler.ServeTrack(pub)
	handler.ServeTrack(pub)

	if callCount != 3 {
		t.Errorf("expected handler to be called 3 times, got %d", callCount)
	}
}

func TestNotFoundWithMockTrackWriter(t *testing.T) {
	mockWriter := &MockTrackWriter{}
	pub := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
		TrackWriter:   mockWriter,
	}

	NotFound(pub)
}

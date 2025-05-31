package moqt

import (
	"testing"
)

func TestSubscriber(t *testing.T) {
	tests := []struct {
		name            string
		broadcastPath   BroadcastPath
		trackName       TrackName
		trackReader     TrackReader
		subscribeStream *sendSubscribeStream
	}{
		{
			name:            "basic subscriber",
			broadcastPath:   BroadcastPath("/live/stream"),
			trackName:       TrackName("video"),
			trackReader:     &MockTrackReader{},
			subscribeStream: nil, // Can be nil for this test
		},
		{
			name:            "empty paths",
			broadcastPath:   BroadcastPath(""),
			trackName:       TrackName(""),
			trackReader:     nil,
			subscribeStream: nil,
		},
		{
			name:            "complex paths",
			broadcastPath:   BroadcastPath("/live/stream/user123/session456"),
			trackName:       TrackName("audio-track-high-quality"),
			trackReader:     &MockTrackReader{},
			subscribeStream: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriber := Subscriber{
				BroadcastPath:   tt.broadcastPath,
				TrackName:       tt.trackName,
				TrackReader:     tt.trackReader,
				SubscribeStream: tt.subscribeStream,
			}

			if subscriber.BroadcastPath != tt.broadcastPath {
				t.Errorf("BroadcastPath = %v, want %v", subscriber.BroadcastPath, tt.broadcastPath)
			}

			if subscriber.TrackName != tt.trackName {
				t.Errorf("TrackName = %v, want %v", subscriber.TrackName, tt.trackName)
			}

			if subscriber.TrackReader != tt.trackReader {
				t.Errorf("TrackReader = %v, want %v", subscriber.TrackReader, tt.trackReader)
			}

			if subscriber.SubscribeStream != tt.subscribeStream {
				t.Errorf("SubscribeStream = %v, want %v", subscriber.SubscribeStream, tt.subscribeStream)
			}
		})
	}
}

func TestSubscriberZeroValue(t *testing.T) {
	var subscriber Subscriber

	if subscriber.BroadcastPath != "" {
		t.Errorf("zero value BroadcastPath = %v, want empty", subscriber.BroadcastPath)
	}

	if subscriber.TrackName != "" {
		t.Errorf("zero value TrackName = %v, want empty", subscriber.TrackName)
	}

	if subscriber.TrackReader != nil {
		t.Errorf("zero value TrackReader = %v, want nil", subscriber.TrackReader)
	}

	if subscriber.SubscribeStream != nil {
		t.Errorf("zero value SubscribeStream = %v, want nil", subscriber.SubscribeStream)
	}
}

func TestSubscriberComparison(t *testing.T) {
	reader1 := &MockTrackReader{}
	reader2 := &MockTrackReader{}

	subscriber1 := Subscriber{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("track1"),
		TrackReader:   reader1,
	}

	subscriber2 := Subscriber{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("track1"),
		TrackReader:   reader1, // Same reader
	}

	subscriber3 := Subscriber{
		BroadcastPath: BroadcastPath("/different"),
		TrackName:     TrackName("track1"),
		TrackReader:   reader2, // Different reader
	}

	// Test same content with same reader
	if subscriber1.BroadcastPath != subscriber2.BroadcastPath ||
		subscriber1.TrackName != subscriber2.TrackName ||
		subscriber1.TrackReader != subscriber2.TrackReader {
		t.Error("subscribers with same content should have equal fields")
	}

	// Test different content
	if subscriber1.BroadcastPath == subscriber3.BroadcastPath &&
		subscriber1.TrackReader == subscriber3.TrackReader {
		t.Error("subscribers with different content should not be equal")
	}
}

func TestSubscriberFieldTypes(t *testing.T) {
	subscriber := Subscriber{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
	}

	// Test that fields have the correct types
	var _ BroadcastPath = subscriber.BroadcastPath
	var _ TrackName = subscriber.TrackName
	var _ TrackReader = subscriber.TrackReader              // Can be nil
	var _ *sendSubscribeStream = subscriber.SubscribeStream // Can be nil
}

func TestSubscriberWithMockReader(t *testing.T) {
	reader := &MockTrackReader{}
	subscriber := Subscriber{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
		TrackReader:   reader,
	}

	if subscriber.TrackReader == nil {
		t.Error("TrackReader should not be nil")
	}

	// Test that we can call methods on the reader
	err := subscriber.TrackReader.Close()
	if err != nil {
		t.Errorf("unexpected error closing reader: %v", err)
	}

	if !reader.AssertCalled(t, "Close") {
		t.Error("expected reader to be closed")
	}
}

package moqt

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	tests := map[string]struct {
		trackWriter *TrackWriter
	}{
		"nil track writer": {
			trackWriter: nil,
		},
		"track writer with nil TrackWriter": {
			trackWriter: newTrackWriter(BroadcastPath("/test"), TrackName("test"), nil, nil, nil),
		}, "track writer with mock TrackWriter": {
			trackWriter: newTrackWriter(BroadcastPath("/test"), TrackName("test"),
				newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{}),
				func() (quic.SendStream, error) {
					return &MockQUICSendStream{}, nil
				}, func() {}),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Should not panic
			NotFound(tt.trackWriter)
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	trackWriter := newTrackWriter(BroadcastPath("/test"), TrackName("test"), nil, nil, nil)

	// Should not panic
	NotFoundHandler.ServeTrack(trackWriter)
}

func TestTrackHandlerFunc(t *testing.T) {
	called := false
	var receivedTrackWriter *TrackWriter

	handler := TrackHandlerFunc(func(tw *TrackWriter) {
		called = true
		receivedTrackWriter = tw
	})

	testTrackWriter := newTrackWriter(BroadcastPath("/test"), TrackName("test"), nil, nil, nil)
	handler.ServeTrack(testTrackWriter)

	assert.True(t, called, "handler function was not called")
	assert.Equal(t, testTrackWriter, receivedTrackWriter, "handler did not receive the correct track writer")
}

func TestTrackHandlerFuncServeTrack(t *testing.T) {
	callCount := 0

	handler := TrackHandlerFunc(func(tw *TrackWriter) {
		callCount++
	})

	trackWriter := newTrackWriter(BroadcastPath("/test"), TrackName("test"), nil, nil, nil)
	// Call multiple times to ensure it works correctly
	handler.ServeTrack(trackWriter)
	handler.ServeTrack(trackWriter)
	handler.ServeTrack(trackWriter)

	assert.Equal(t, 3, callCount, "expected handler to be called 3 times")
}

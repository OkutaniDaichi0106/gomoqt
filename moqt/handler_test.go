package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	tests := map[string]struct {
		pub *Publisher
	}{
		"nil publisher": {
			pub: nil,
		},
		"publisher with nil TrackWriter": {
			pub: &Publisher{
				BroadcastPath: BroadcastPath("/test"),
				TrackName:     TrackName("test"),
				TrackWriter:   nil,
			},
		}, "publisher with mock TrackWriter": {
			pub: &Publisher{
				BroadcastPath: BroadcastPath("/test"),
				TrackName:     TrackName("test"),
				TrackWriter: func() *MockTrackWriter {
					mock := &MockTrackWriter{}
					mock.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)
					return mock
				}(),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Should not panic
			NotFound(tt.pub)
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	mockWriter := &MockTrackWriter{}
	mockWriter.On("CloseWithError", TrackNotFoundErrorCode).Return(nil)

	pub := &Publisher{
		BroadcastPath: BroadcastPath("/test"),
		TrackName:     TrackName("test"),
		TrackWriter:   mockWriter,
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

	assert.True(t, called, "handler function was not called")
	assert.Equal(t, testPub, receivedPub, "handler did not receive the correct publisher")
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

	assert.Equal(t, 3, callCount, "expected handler to be called 3 times")
}

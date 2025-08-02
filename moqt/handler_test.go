package moqt

import (
	"context"
	"io"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			expectPanic: true,
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
			if tt.expectPanic {
				assert.Panics(t, func() {
					NotFound(tt.trackWriter)
				})
			} else {
				// Should not panic
				NotFound(tt.trackWriter)
			}
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
			expectPanic: true,
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

			if tt.expectPanic {
				// Should panic because receiveSubscribeStream is nil
				assert.Panics(t, func() {
					NotFoundHandler.ServeTrack(trackWriter)
				})
			} else {
				// Should not panic
				NotFoundHandler.ServeTrack(trackWriter)
			}
		})
	}
}

func TestTrackHandlerFunc(t *testing.T) {
	called := false
	var receivedTrackWriter *TrackWriter

	handler := TrackHandlerFunc(func(tw *TrackWriter) {
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

	handler.ServeTrack(testTrackWriter)

	assert.True(t, called, "handler function was not called")
	assert.Equal(t, testTrackWriter, receivedTrackWriter, "handler did not receive the correct track writer")
}

func TestTrackHandlerFuncServeTrack(t *testing.T) {
	callCount := 0

	handler := TrackHandlerFunc(func(tw *TrackWriter) {
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

	// Call multiple times to ensure it works correctly
	handler.ServeTrack(trackWriter)
	handler.ServeTrack(trackWriter)
	handler.ServeTrack(trackWriter)

	assert.Equal(t, 3, callCount, "expected handler to be called 3 times")
}

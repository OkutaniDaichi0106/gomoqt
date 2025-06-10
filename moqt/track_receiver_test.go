package moqt

import (
	"context"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewTrackReceiver(t *testing.T) {
	// Create mock send subscribe stream
	ctx := context.Background()
	mockStream := &MockQUICStream{}
	substr := newSendSubscribeStream(ctx, SubscribeID(1), mockStream, &SubscribeConfig{})

	receiver := newTrackReceiver(substr)

	assert.NotNil(t, receiver, "newTrackReceiver should not return nil")
	assert.Equal(t, substr, receiver.substr, "substr should be set correctly")
	assert.NotNil(t, receiver.queue, "queue should be initialized")
	assert.NotNil(t, receiver.queuedCh, "queuedCh should be initialized")
	assert.NotNil(t, receiver.dequeued, "dequeued should be initialized")
}

func TestTrackReceiver_AcceptGroup(t *testing.T) {
	// Create mock send subscribe stream
	ctx := context.Background()
	mockStream := &MockQUICStream{}
	substr := newSendSubscribeStream(ctx, SubscribeID(1), mockStream, &SubscribeConfig{})

	receiver := newTrackReceiver(substr)

	// Test with a timeout to ensure we don't block forever when no groups are available
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected timeout error when no groups are available")
	assert.Equal(t, context.DeadlineExceeded, err, "expected deadline exceeded error")
}

func TestTrackReceiver_Close(t *testing.T) {
	// Create mock send subscribe stream
	ctx := context.Background()
	mockStream := &MockQUICStream{}
	substr := newSendSubscribeStream(ctx, SubscribeID(1), mockStream, &SubscribeConfig{})

	// Mock the Close method
	mockStream.On("Close").Return(nil)

	receiver := newTrackReceiver(substr)

	err := receiver.Close()
	assert.NoError(t, err, "Close() should not return error")
	// Verify Close was called on the underlying stream
	mockStream.AssertCalled(t, "Close")
}

func TestTrackReceiver_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		reason SubscribeErrorCode
	}{
		"close with custom error": {
			reason: SubscribeErrorCode(1),
		},
		"close with zero error code": {
			reason: SubscribeErrorCode(0),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock send subscribe stream
			ctx := context.Background()
			mockStream := &MockQUICStream{}
			substr := newSendSubscribeStream(ctx, SubscribeID(1), mockStream, &SubscribeConfig{})

			// Mock the necessary methods
			mockStream.On("StreamID").Return(quic.StreamID(1))
			mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()

			receiver := newTrackReceiver(substr)

			err := receiver.CloseWithError(tt.reason)
			assert.NoError(t, err, "CloseWithError() should not return error")

			// Verify methods were called on the underlying stream
			mockStream.AssertCalled(t, "CancelWrite", quic.StreamErrorCode(tt.reason))
			mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(tt.reason))
		})
	}
}

func TestTrackReceiver_Interface(t *testing.T) {
	// Verify that trackReceiver implements TrackReader interface
	var _ TrackReader = (*trackReceiver)(nil)
}

func TestTrackReceiver_AcceptGroup_RealImplementation(t *testing.T) {
	// Create mock send subscribe stream
	ctx := context.Background()
	mockStream := &MockQUICStream{}
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(128),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	substr := newSendSubscribeStream(ctx, SubscribeID(1), mockStream, config)

	receiver := newTrackReceiver(substr)

	// Test with a timeout to ensure we don't block forever
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected timeout error when no groups are available")
	assert.Equal(t, context.DeadlineExceeded, err, "expected deadline exceeded error")
}

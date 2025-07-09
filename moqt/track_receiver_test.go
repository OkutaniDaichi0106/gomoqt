package moqt

import (
	"context"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackReceiver(t *testing.T) {
	ctx := context.Background()
	receiver := newTrackReceiver(ctx)

	assert.NotNil(t, receiver, "newTrackReceiver should not return nil")
	assert.NotNil(t, receiver.queue, "queue should be initialized")
	assert.NotNil(t, receiver.queuedCh, "queuedCh should be initialized")
	assert.NotNil(t, receiver.dequeued, "dequeued should be initialized")
}

func TestTrackReceiver_AcceptGroup(t *testing.T) {
	ctx := context.Background()
	receiver := newTrackReceiver(ctx)

	// Test with a timeout to ensure we don't block forever when no groups are available
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected timeout error when no groups are available")
	assert.Equal(t, context.DeadlineExceeded, err, "expected deadline exceeded error")
}

func TestTrackReceiver_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	receiver := newTrackReceiver(ctx)

	// Cancel the context
	cancel()

	// Test that AcceptGroup returns context error when context is cancelled
	testCtx, testCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer testCancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected error when context is cancelled")
	// Should return context.Canceled or DeadlineExceeded
	assert.True(t, err == context.Canceled || err == context.DeadlineExceeded, "expected context error")
}

func TestTrackReceiver_EnqueueGroup(t *testing.T) {
	ctx := context.Background()
	receiver := newTrackReceiver(ctx)

	// Mock receive stream
	mockReceiveStream := &MockQUICReceiveStream{}
	mockReceiveStream.On("StreamID").Return(quic.StreamID(1))

	// Enqueue a group
	receiver.enqueueGroup(GroupSequence(1), mockReceiveStream)

	// Test that we can accept the enqueued group
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	group, err := receiver.AcceptGroup(testCtx)
	assert.NoError(t, err, "should be able to accept enqueued group")
	assert.NotNil(t, group, "accepted group should not be nil")

	mockReceiveStream.AssertExpectations(t)
}

func TestTrackReceiver_Interface(t *testing.T) {
	// Verify that trackReceiver implements TrackReader interface
	var _ TrackReader = (*trackReceiver)(nil)
}

func TestTrackReceiver_AcceptGroup_RealImplementation(t *testing.T) {
	ctx := context.Background()
	receiver := newTrackReceiver(ctx)

	// Test with a timeout to ensure we don't block forever
	testCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(testCtx)
	assert.Error(t, err, "expected timeout error when no groups are available")
	assert.Equal(t, context.DeadlineExceeded, err, "expected deadline exceeded error")
}

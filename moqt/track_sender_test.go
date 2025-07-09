package moqt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTrackSender(t *testing.T) {
	ctx := context.Background()
	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(ctx, mockSendStream, seq), nil
	}

	sender := newTrackSender(ctx, openGroupFunc)

	require.NotNil(t, sender, "newTrackSender should not return nil")
	assert.NotNil(t, sender.queue, "queue should be initialized")
	assert.NotNil(t, sender.openGroupFunc, "openGroupFunc should be set")
}

func TestTrackSender_OpenGroup(t *testing.T) {
	ctx := context.Background()

	var acceptCalled bool
	var acceptedInfo Info
	acceptFunc := func(info Info) {
		acceptCalled = true
		acceptedInfo = info
	}

	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(ctx, mockSendStream, seq), nil
	}

	sender := newTrackSender(ctx, openGroupFunc)
	sender.acceptFunc = acceptFunc

	// Test opening a group
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err, "OpenGroup should not return error")
	assert.NotNil(t, group, "group should not be nil")
	assert.True(t, acceptCalled, "accept function should be called")
	assert.Equal(t, Info{}, acceptedInfo, "accept function should be called with empty info")
}

func TestTrackSender_OpenGroup_ZeroSequence(t *testing.T) {
	ctx := context.Background()
	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		return nil, nil
	}

	sender := newTrackSender(ctx, openGroupFunc)

	// Test opening a group with zero sequence
	group, err := sender.OpenGroup(GroupSequence(0))
	assert.Error(t, err, "OpenGroup should return error for zero sequence")
	assert.Nil(t, group, "group should be nil for zero sequence")
	assert.Contains(t, err.Error(), "group sequence must not be zero")
}

func TestTrackSender_OpenGroup_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context

	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		return nil, nil
	}

	sender := newTrackSender(ctx, openGroupFunc)

	// Test opening a group with canceled context
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error with canceled context")
	assert.Nil(t, group, "group should be nil with canceled context")
	assert.Equal(t, context.Canceled, err, "error should be context.Canceled")
}

func TestTrackSender_OpenGroup_OpenGroupError(t *testing.T) {
	ctx := context.Background()
	expectedError := errors.New("failed to open group")

	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		return nil, expectedError
	}

	sender := newTrackSender(ctx, openGroupFunc)
	sender.acceptFunc = func(Info) {} // Add accept function

	// Test opening a group when openGroupFunc returns error
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error when openGroupFunc fails")
	assert.Nil(t, group, "group should be nil when openGroupFunc fails")
	assert.Equal(t, expectedError, err, "error should match the error from openGroupFunc")
}

func TestTrackSender_OpenGroup_Success(t *testing.T) {
	ctx := context.Background()

	var acceptCalled bool
	acceptFunc := func(info Info) {
		acceptCalled = true
	}

	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(ctx, mockSendStream, seq), nil
	}

	sender := newTrackSender(ctx, openGroupFunc)
	sender.acceptFunc = acceptFunc

	// Test successful group opening
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err, "OpenGroup should not return error")
	assert.NotNil(t, group, "group should not be nil")
	assert.True(t, acceptCalled, "accept function should be called")
	assert.Equal(t, GroupSequence(1), group.GroupSequence(), "group sequence should match")
}

func TestTrackSender_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	openGroupFunc := func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", quic.StreamErrorCode(SubscribeCanceledErrorCode)).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(ctx, mockSendStream, seq), nil
	}

	sender := newTrackSender(ctx, openGroupFunc)
	sender.acceptFunc = func(Info) {}

	// Open a group
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err)
	assert.NotNil(t, group)

	// Cancel the context to simulate cleanup
	cancel()

	// Give some time for cleanup goroutine to process
	time.Sleep(50 * time.Millisecond)

	// Verify that the queue is cleaned up
	sender.mu.Lock()
	queueIsNil := sender.queue == nil
	sender.mu.Unlock()
	assert.True(t, queueIsNil, "queue should be nil after context cancellation")
}

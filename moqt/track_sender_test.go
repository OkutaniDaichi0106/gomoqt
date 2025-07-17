package moqt

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTrackSender(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	acceptFunc := func(info Info) error {
		// Mock accept function, can be empty for this test
		return nil
	}
	onCloseTrack := func() {
		// Mock onCloseTrack function
	}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	require.NotNil(t, sender, "newTrackSender should not return nil")
	assert.NotNil(t, sender.groupsMap, "groupsMap should be initialized")
	assert.NotNil(t, sender.openUniStreamFunc, "openUniStreamFunc should be set")
	assert.NotNil(t, sender.acceptFunc, "acceptFunc should be set")
	assert.NotNil(t, sender.onCloseTrack, "onCloseTrack should be set")
}

func TestTrackSender_OpenGroup(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	var acceptCalled bool
	var acceptedInfo Info
	acceptFunc := func(info Info) error {
		acceptCalled = true
		acceptedInfo = info
		return nil
	}

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}

	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Test opening a group
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err, "OpenGroup should not return error")
	assert.NotNil(t, group, "group should not be nil")
	assert.True(t, acceptCalled, "accept function should be called")
	assert.Equal(t, Info{}, acceptedInfo, "accept function should be called with empty info")
}

func TestTrackSender_OpenGroup_ZeroSequence(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, nil
	}
	acceptFunc := func(info Info) error {
		// Mock accept function, can be empty for this test
		return nil
	}
	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Test opening a group with zero sequence
	group, err := sender.OpenGroup(GroupSequence(0))
	assert.Error(t, err, "OpenGroup should return error for zero sequence")
	assert.Nil(t, group, "group should be nil for zero sequence")
	assert.Contains(t, err.Error(), "group sequence must not be zero")
}

func TestTrackSender_OpenGroup_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context
	logger := slog.Default()

	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, nil
	}
	acceptFunc := func(info Info) error {
		// Mock accept function, can be empty for this test
		return nil
	}
	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Test opening a group with canceled context
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error with canceled context")
	assert.Nil(t, group, "group should be nil with canceled context")
	assert.Equal(t, context.Canceled, err, "error should be context.Canceled")
}

func TestTrackSender_OpenGroup_OpenGroupError(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	expectedError := errors.New("failed to open group")

	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, expectedError
	}
	acceptFunc := func(info Info) error {
		// Mock accept function, can be empty for this test
		return nil
	}
	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Test opening a group when openUniStreamFunc returns error
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error when openUniStreamFunc fails")
	assert.Nil(t, group, "group should be nil when openUniStreamFunc fails")
	assert.Contains(t, err.Error(), expectedError.Error(), "error should contain the error from openUniStreamFunc")
}

func TestTrackSender_OpenGroup_Success(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	var acceptCalled bool
	acceptFunc := func(info Info) error {
		acceptCalled = true
		return nil
	}

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}

	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Test successful group opening
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err, "OpenGroup should not return error")
	assert.NotNil(t, group, "group should not be nil")
	assert.True(t, acceptCalled, "accept function should be called")
	assert.Equal(t, GroupSequence(1), group.GroupSequence(), "group sequence should match")
}

func TestTrackSender_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := slog.Default()

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", quic.StreamErrorCode(SubscribeCanceledErrorCode)).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	acceptFunc := func(info Info) error {
		// Mock accept function, can be empty for this test
		return nil
	}
	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Open a group
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err)
	assert.NotNil(t, group)

	// Cancel the context to simulate cleanup
	cancel()

	// Give some time for cleanup goroutine to process
	time.Sleep(50 * time.Millisecond)

	// Verify that the groupsMap is cleaned up when calling Close
	sender.Close()
	sender.mu.Lock()
	groupsMapIsNil := sender.groupsMap == nil
	sender.mu.Unlock()
	assert.True(t, groupsMapIsNil, "groupsMap should be nil after Close()")
}

func TestTrackSender_Close(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	var onCloseTrackCalled bool
	onCloseTrack := func() {
		onCloseTrackCalled = true
	}

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}

	acceptFunc := func(info Info) error {
		return nil
	}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Open a group to have something in the groupsMap
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err)
	assert.NotNil(t, group)

	// Verify that groupsMap has an entry
	sender.mu.Lock()
	hasEntry := len(sender.groupsMap) > 0
	sender.mu.Unlock()
	assert.True(t, hasEntry, "groupsMap should have an entry")

	// Close the sender
	sender.Close()

	// Verify that onCloseTrack was called
	assert.True(t, onCloseTrackCalled, "onCloseTrack should be called")

	// Verify that groupsMap is nil
	sender.mu.Lock()
	groupsMapIsNil := sender.groupsMap == nil
	sender.mu.Unlock()
	assert.True(t, groupsMapIsNil, "groupsMap should be nil after Close()")
}

func TestTrackSender_OpenGroup_AcceptError(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	expectedError := errors.New("accept failed")

	acceptFunc := func(info Info) error {
		return expectedError
	}

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		return mockSendStream, nil
	}

	onCloseTrack := func() {}

	sender := newTrackSender(ctx, logger, openUniStreamFunc, acceptFunc, onCloseTrack)

	// Test opening a group when acceptFunc returns error
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error when acceptFunc fails")
	assert.Nil(t, group, "group should be nil when acceptFunc fails")
	assert.Equal(t, expectedError, err, "error should match the error from acceptFunc")
}

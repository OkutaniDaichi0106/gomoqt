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
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	substr := newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{})
	onCloseTrack := func() {
		// Mock onCloseTrack function
	}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	require.NotNil(t, sender, "newTrackSender should not return nil")
	assert.NotNil(t, sender.activeGroups, "activeGroups should be initialized")
	assert.NotNil(t, sender.openUniStreamFunc, "openUniStreamFunc should be set")
	assert.NotNil(t, sender.receiveSubscribeStream, "subscribeStream should be set")
	assert.NotNil(t, sender.onCloseTrackFunc, "onCloseTrack should be set")
}

func TestTrackSender_OpenGroup(t *testing.T) {
	var acceptCalled bool
	var acceptedInfo Info

	substr := newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{})

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}

	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test opening a group
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err, "OpenGroup should not return error")
	assert.NotNil(t, group, "group should not be nil")
	assert.True(t, acceptCalled, "accept function should be called")
	assert.Equal(t, Info{}, acceptedInfo, "accept function should be called with empty info")
}

func TestTrackSender_OpenGroup_ZeroSequence(t *testing.T) {
	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, nil
	}
	substr := newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{})
	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test opening a group with zero sequence
	group, err := sender.OpenGroup(GroupSequence(0))
	assert.Error(t, err, "OpenGroup should return error for zero sequence")
	assert.Nil(t, group, "group should be nil for zero sequence")
	assert.Contains(t, err.Error(), "group sequence must not be zero")
}

func TestTrackSender_OpenGroup_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)

	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, nil
	}
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test opening a group with canceled context
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error with canceled context")
	assert.Nil(t, group, "group should be nil with canceled context")
	assert.Equal(t, context.Canceled, err, "error should be context.Canceled")
}

func TestTrackSender_OpenGroup_OpenGroupError(t *testing.T) {
	expectedError := errors.New("failed to open group")

	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, expectedError
	}

	substr := newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{})

	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test opening a group when openUniStreamFunc returns error
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error when openUniStreamFunc fails")
	assert.Nil(t, group, "group should be nil when openUniStreamFunc fails")
	assert.Contains(t, err.Error(), expectedError.Error(), "error should contain the error from openUniStreamFunc")
}

func TestTrackSender_OpenGroup_Success(t *testing.T) {
	var acceptCalled bool
	mockStream := &MockQUICStream{
		WriteFunc: func(b []byte) (int, error) {
			acceptCalled = true
			return len(b), nil
		},
	}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Write").Return(quic.StreamID(1))
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}

	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test successful group opening
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err, "OpenGroup should not return error")
	assert.NotNil(t, group, "group should not be nil")
	assert.True(t, acceptCalled, "accept function should be called")
	assert.Equal(t, GroupSequence(1), group.GroupSequence(), "group sequence should match")
}

func TestTrackSender_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("CancelWrite", quic.StreamErrorCode(SubscribeCanceledErrorCode)).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		mockSendStream.On("Context").Return(ctx)
		return mockSendStream, nil
	}
	substr := newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{})
	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

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
	groupsMapIsNil := sender.activeGroups == nil
	sender.mu.Unlock()
	assert.True(t, groupsMapIsNil, "groupsMap should be nil after Close()")
}

func TestTrackSender_Close(t *testing.T) {

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

	substr := newReceiveSubscribeStream(SubscribeID(1), &MockQUICStream{}, &TrackConfig{})

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Open a group to have something in the groupsMap
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err)
	assert.NotNil(t, group)

	// Verify that activeGroups has an entry
	sender.mu.Lock()
	hasEntry := len(sender.activeGroups) > 0
	sender.mu.Unlock()
	assert.True(t, hasEntry, "activeGroups should have an entry")

	// Close the sender
	sender.Close()

	// Verify that onCloseTrack was called
	assert.True(t, onCloseTrackCalled, "onCloseTrack should be called")

	// Verify that activeGroups is nil
	sender.mu.Lock()
	activeGroupsIsNil := sender.activeGroups == nil
	sender.mu.Unlock()
	assert.True(t, activeGroupsIsNil, "activeGroups should be nil after Close()")
}

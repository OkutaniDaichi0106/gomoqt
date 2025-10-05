package moqt

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTrackWriter(t *testing.T) {
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Context").Return(context.Background())
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	onCloseTrack := func() {
		// Mock onCloseTrack function
	}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	require.NotNil(t, sender, "newTrackWriter should not return nil")
	assert.NotNil(t, sender.activeGroups, "activeGroups should be initialized")
	assert.NotNil(t, sender.openUniStreamFunc, "openUniStreamFunc should be set")
	assert.NotNil(t, sender.receiveSubscribeStream, "subscribeStream should be set")
	assert.NotNil(t, sender.onCloseTrackFunc, "onCloseTrack should be set")
}

func TestTrackWriter_OpenGroup(t *testing.T) {
	var acceptCalled bool

	mockStream := &MockQUICStream{
		WriteFunc: func(b []byte) (int, error) {
			acceptCalled = true
			return len(b), nil
		},
	}
	mockStream.On("Context").Return(context.Background())
	// Mock the Read method to return EOF to stop the background goroutine
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	// Mock the Write method for sending messages
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Context").Return(context.Background())
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
}

func TestTrackWriter_OpenGroup_ZeroSequence(t *testing.T) {
	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, nil
	}
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test opening a group with zero sequence
	group, err := sender.OpenGroup(GroupSequence(0))
	assert.Error(t, err, "OpenGroup should return error for zero sequence")
	assert.Nil(t, group, "group should be nil for zero sequence")
	assert.Contains(t, err.Error(), "group sequence must not be zero")
}

func TestTrackWriter_OpenGroup_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)

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

func TestTrackWriter_OpenGroup_OpenGroupError(t *testing.T) {
	expectedError := errors.New("failed to open group")

	openUniStreamFunc := func() (quic.SendStream, error) {
		return nil, expectedError
	}

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Test opening a group when openUniStreamFunc returns error
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.Error(t, err, "OpenGroup should return error when openUniStreamFunc fails")
	assert.Nil(t, group, "group should be nil when openUniStreamFunc fails")
	assert.Contains(t, err.Error(), expectedError.Error(), "error should contain the error from openUniStreamFunc")
}

func TestTrackWriter_OpenGroup_Success(t *testing.T) {
	var acceptCalled bool
	mockStream := &MockQUICStream{
		WriteFunc: func(b []byte) (int, error) {
			acceptCalled = true
			return len(b), nil
		},
	}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Context").Return(context.Background())
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

func TestTrackWriter_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Context").Return(ctx)
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(ctx)
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	// Open a group first
	group, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err)
	assert.NotNil(t, group)

	// Cancel the context to simulate cancellation
	cancel()

	// Try to open another group - this should fail due to cancelled context
	group2, err := sender.OpenGroup(GroupSequence(2))
	assert.Error(t, err, "OpenGroup should return error with cancelled context")
	assert.Nil(t, group2, "group should be nil with cancelled context")
	assert.Equal(t, context.Canceled, err, "error should be context.Canceled")
}

func TestTrackWriter_Close(t *testing.T) {
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Context").Return(context.Background())
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}

	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.Anything).Return()
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	var onCloseTrackCalled bool
	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, func() {
		onCloseTrackCalled = true
	})

	// Verify that activeGroups is initialized
	sender.groupMapMu.Lock()
	isInitialized := sender.activeGroups != nil
	sender.groupMapMu.Unlock()
	assert.True(t, isInitialized, "activeGroups should be initialized")

	// Close the sender (without opening any groups to avoid deadlock)
	err := sender.Close()
	assert.NoError(t, err, "Close should not return an error")

	// Verify that onCloseTrack was called
	assert.True(t, onCloseTrackCalled, "onCloseTrack should be called")

	// Verify that activeGroups is nil
	sender.groupMapMu.Lock()
	activeGroupsIsNil := sender.activeGroups == nil
	sender.groupMapMu.Unlock()
	assert.True(t, activeGroupsIsNil, "activeGroups should be nil after Close()")
}

func TestTrackWriter_Context(t *testing.T) {
	openUniStreamFunc := func() (quic.SendStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Context").Return(context.Background())
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("Write", mock.Anything).Return(0, nil)
		return mockSendStream, nil
	}
	mockStream := &MockQUICStream{}
	mockStream.On("Context").Return(context.Background())
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.Anything).Return(0, nil)
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &TrackConfig{})
	onCloseTrack := func() {}

	sender := newTrackWriter("/broadcast/path", "track_name", substr, openUniStreamFunc, onCloseTrack)

	ctx := sender.Context()
	assert.NotNil(t, ctx)
}

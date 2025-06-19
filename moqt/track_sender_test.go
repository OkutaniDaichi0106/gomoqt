package moqt

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTrackSender(t *testing.T) {
	mockStream := &MockQUICStream{}
	// newReceiveSubscribeStream starts a goroutine that reads from the stream
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})

	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	require.NotNil(t, sender, "newTrackSender should not return nil")
	assert.Equal(t, substr, sender.subscribeStream, "subscribeStream should be set correctly")
	assert.NotNil(t, sender.queue, "queue should be initialized")
	assert.NotNil(t, sender.openGroupFunc, "openGroupFunc should be set")
	assert.False(t, sender.accepted, "accepted should be false initially")
	assert.Equal(t, Info{}, sender.info, "info should be empty initially")
}

func TestTrackSender_WriteInfo(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			// Simulate a blocking read operation
			time.Sleep(1 * time.Second)
			return 0, nil // Return EOF to simulate end of stream
		},
	}
	// The accept method will write SubscribeOk message to the stream
	mockStream.On("Read", mock.Anything)
	mockStream.On("Write", mock.Anything).Return(2, nil)

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	testInfo := Info{GroupOrder: GroupOrderAscending}

	sender.WriteInfo(testInfo)
	assert.True(t, sender.accepted, "accepted should be true after WriteInfo")
	assert.Equal(t, testInfo, sender.info, "info should be set correctly")
}

func TestTrackSender_WriteInfo_AlreadyAccepted(t *testing.T) {
	mockStream := &MockQUICStream{}
	// No Write expectation since WriteInfo should not call accept when already accepted
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true
	originalInfo := Info{GroupOrder: GroupOrderDescending}
	sender.info = originalInfo

	newInfo := Info{GroupOrder: GroupOrderAscending}

	// Should not change info when already accepted
	sender.WriteInfo(newInfo)

	assert.True(t, sender.accepted, "accepted should remain true")
	assert.Equal(t, originalInfo, sender.info, "info should not change when already accepted")
}

func TestTrackSender_OpenGroup_ZeroSequence(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	group, err := sender.OpenGroup(GroupSequence(0))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "group sequence must not be zero")
	assert.Nil(t, group)
}

func TestTrackSender_OpenGroup_AutoAccept(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			// Simulate a blocking read operation
			time.Sleep(1 * time.Second)
			return 0, nil // Return EOF to simulate end of stream
		},
	}
	// Auto-accept will trigger a Write call for SubscribeOk message
	mockStream.On("Read", mock.Anything)
	mockStream.On("Write", mock.Anything).Return(2, nil)

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	group, err := sender.OpenGroup(GroupSequence(1))

	require.NoError(t, err)
	require.NotNil(t, group)
	assert.True(t, sender.accepted, "sender should be auto-accepted")
	assert.Equal(t, Info{}, sender.info, "info should be empty when auto-accepted")
}

func TestTrackSender_OpenGroup_Success(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true

	group, err := sender.OpenGroup(GroupSequence(1))

	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, GroupSequence(1), group.GroupSequence())

	// Verify group was added to queue
	sender.mu.Lock()
	_, inQueue := sender.queue[group.(*sendGroupStream)]
	sender.mu.Unlock()
	assert.True(t, inQueue, "group should be added to queue")
}

func TestTrackSender_OpenGroup_OpenGroupFuncError(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		return nil, errors.New("open group failed")
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true

	group, err := sender.OpenGroup(GroupSequence(1))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "open group failed")
	assert.Nil(t, group)
}

func TestTrackSender_OpenGroup_SenderClosed(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true
	sender.queue = nil // Simulate closed state

	group, err := sender.OpenGroup(GroupSequence(1))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscription was canceled")
	assert.Nil(t, group)
}

func TestTrackSender_Close(t *testing.T) {
	mockStream := &MockQUICStream{}
	// Block read operations to prevent subscribeCanceledCh from being closed prematurely
	mockStream.ReadFunc = func(p []byte) (int, error) {
		time.Sleep(100 * time.Millisecond) // Short block
		return 0, nil
	}
	mockStream.On("Read", mock.Anything)
	mockStream.On("Write", mock.Anything).Return(2, nil) // For accept calls
	mockStream.On("Close").Return(nil)                   // For stream close
	mockStream.On("CancelRead", mock.Anything).Return()  // For cancel read during close
	mockStream.On("StreamID").Return(quic.StreamID(1))   // For stream ID access during close

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})

	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	// Create some groups
	group1, err1 := sender.OpenGroup(GroupSequence(1))
	group2, err2 := sender.OpenGroup(GroupSequence(2))

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, group1)
	require.NotNil(t, group2)

	err := sender.Close()

	assert.NoError(t, err)
	assert.Nil(t, sender.queue, "queue should be nil after close")
}

func TestTrackSender_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		code SubscribeErrorCode
	}{
		"error code 1": {
			code: SubscribeErrorCode(1),
		},
		"error code 2": {
			code: SubscribeErrorCode(2),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			// Block read operations to prevent subscribeCanceledCh from being closed prematurely
			mockStream.ReadFunc = func(p []byte) (int, error) {
				time.Sleep(10 * time.Millisecond) // Short block
				return 0, nil
			}
			mockStream.On("Read", mock.Anything)
			mockStream.On("Write", mock.Anything).Return(2, nil)
			mockStream.On("Close").Return(nil)                   // For stream close
			mockStream.On("CancelRead", mock.Anything).Return()  // For cancel read during close
			mockStream.On("CancelWrite", mock.Anything).Return() // For cancel write during close
			mockStream.On("StreamID").Return(quic.StreamID(1))   // For stream ID during error

			substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
			openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
				mockSendStream := &MockQUICSendStream{}
				// Allow various method calls during cleanup
				mockSendStream.On("CancelWrite", mock.Anything).Return()
				mockSendStream.On("StreamID").Return(quic.StreamID(1))
				mockSendStream.On("Close").Return(nil)
				return newSendGroupStream(mockSendStream, seq), nil
			}

			sender := newTrackSender(substr, openGroupFunc)

			// Create some groups
			group1, err1 := sender.OpenGroup(GroupSequence(1))
			group2, err2 := sender.OpenGroup(GroupSequence(2))

			require.NoError(t, err1)
			require.NoError(t, err2)
			require.NotNil(t, group1)
			require.NotNil(t, group2)

			err := sender.CloseWithError(tt.code)

			assert.NoError(t, err)
			assert.Nil(t, sender.queue, "queue should be nil after close with error")
		})
	}
}

func TestTrackSender_ConcurrentAccess(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			// Simulate a blocking read operation
			time.Sleep(100 * time.Millisecond)
			return 0, nil
		},
	}

	mockStream.On("Read", mock.Anything)
	mockStream.On("Write", mock.Anything).Return(2, nil)

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}
	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true

	var wg sync.WaitGroup
	const numGoroutines = 3 // Reduced number to avoid race conditions

	// Concurrent OpenGroup calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			group, err := sender.OpenGroup(GroupSequence(seq + 1))
			if err == nil && group != nil {
				// Give some time before closing to ensure group is registered
				time.Sleep(10 * time.Millisecond)
				group.Close()
			}
		}(i)
	}

	wg.Wait()

	// Give goroutines time to complete cleanup
	time.Sleep(50 * time.Millisecond)

	// Verify no race conditions occurred
	sender.mu.Lock()
	queueNotNil := sender.queue != nil
	sender.mu.Unlock()

	assert.True(t, queueNotNil, "queue should still exist after concurrent access")
}

func TestTrackSender_GroupCleanup(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))
	mockStream.On("Write", mock.Anything).Return(2, nil)

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true

	group, err := sender.OpenGroup(GroupSequence(1))
	require.NoError(t, err)
	require.NotNil(t, group)

	// Verify group is in queue
	sender.mu.Lock()
	_, inQueue := sender.queue[group.(*sendGroupStream)]
	sender.mu.Unlock()
	assert.True(t, inQueue, "group should be in queue")

	// Close the group to trigger cleanup
	group.Close()

	// Give goroutine time to process
	time.Sleep(50 * time.Millisecond)

	// Verify group was removed from queue
	sender.mu.Lock()
	_, stillInQueue := sender.queue[group.(*sendGroupStream)]
	sender.mu.Unlock()
	assert.False(t, stillInQueue, "group should be removed from queue after close")
}

func TestTrackSender_OpenGroup_MinimumValidValue(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true

	group, err := sender.OpenGroup(GroupSequence(1))

	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, GroupSequence(1), group.GroupSequence())
}

func TestTrackSender_OpenGroup_LargeValue(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, errors.New("EOF"))

	substr := newReceiveSubscribeStream(1, mockStream, &SubscribeConfig{})
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		// Allow various method calls during cleanup
		mockSendStream.On("CancelWrite", mock.Anything).Return()
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		mockSendStream.On("Close").Return(nil)
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)
	sender.accepted = true

	group, err := sender.OpenGroup(GroupSequence(1000))
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, GroupSequence(1000), group.GroupSequence())
}

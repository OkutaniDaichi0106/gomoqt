package moqt

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewTrackSender(t *testing.T) {
	// Create mock receive subscribe stream
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.AnythingOfType("[]uint8")).Return(0, io.EOF)
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("StreamID").Return(uint64(1))
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

	// Create mock open group function
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		return &sendGroupStream{}, nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	assert.NotNil(t, sender, "newTrackSender should not return nil")
	assert.NotNil(t, sender.openGroupFunc, "openGroupFunc should be set")
	assert.Equal(t, substr, sender.subscribeStream, "subscribeStream should be set correctly")
}

func TestTrackSender_OpenGroup(t *testing.T) {
	tests := map[string]struct {
		streamClosed   bool
		openGroupError error
		seq            GroupSequence
		expectError    bool
	}{
		"successful open": {
			streamClosed: false,
			seq:          GroupSequence(1),
			expectError:  false,
		},
		"stream closed": {
			streamClosed: true,
			seq:          GroupSequence(2),
			expectError:  true,
		},
		"open group error": {
			streamClosed:   false,
			openGroupError: errors.New("mock error"), seq: GroupSequence(3),
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) { // Create mock receive subscribe stream
			mockStream := &MockQUICStream{}
			mockStream.On("Read", mock.AnythingOfType("[]uint8")).Return(0, io.EOF)
			mockStream.On("Close").Return(nil)
			mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("StreamID").Return(quic.StreamID(1))

			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

			// Close the stream if needed
			if tt.streamClosed {
				substr.close()
			}

			// Create test stream with proper closedCh
			testStream := &sendGroupStream{
				closedCh: make(chan struct{}),
			}

			// Create mock open group function
			openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
				if tt.openGroupError != nil {
					return nil, tt.openGroupError
				}
				return testStream, nil
			}

			sender := newTrackSender(substr, openGroupFunc)

			// Initial queue should be empty
			sender.mu.Lock()
			initialQueueSize := len(sender.queue)
			sender.mu.Unlock()

			assert.Equal(t, 0, initialQueueSize, "Initial queue size should be 0")

			groupWriter, err := sender.OpenGroup(tt.seq)

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				assert.Nil(t, groupWriter, "expected nil GroupWriter on error")

				// Queue should still be empty on error
				sender.mu.Lock()
				queueSize := len(sender.queue)
				sender.mu.Unlock()
				assert.Equal(t, 0, queueSize, "Queue size should be 0 after error")
			} else {
				assert.NoError(t, err, "unexpected error")
				assert.NotNil(t, groupWriter, "expected non-nil GroupWriter")

				// Verify that stream was added to the queue
				sender.mu.Lock()
				queueSize := len(sender.queue)
				_, hasTestStream := sender.queue[testStream]
				sender.mu.Unlock()

				assert.Equal(t, 1, queueSize, "Queue size should be 1 after success")
				assert.True(t, hasTestStream, "Test stream not found in queue")

				// Test that the group is removed when stream is closed
				close(testStream.closedCh)

				// Small delay to allow goroutine to run
				time.Sleep(10 * time.Millisecond)

				// Verify that stream was removed from queue
				sender.mu.Lock()
				finalQueueSize := len(sender.queue)
				sender.mu.Unlock()

				assert.Equal(t, 0, finalQueueSize, "Queue size should be 0 after stream closed")
			}
		})
	}
}

func TestTrackSender_Close(t *testing.T) {
	tests := map[string]struct {
		streamClosed bool
		expectError  bool
	}{
		"successful close": {
			streamClosed: false,
			expectError:  false,
		},
		"already closed": {
			streamClosed: true,
			expectError:  false, // Should handle multiple close calls gracefully
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock receive subscribe stream
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					// Block indefinitely to prevent listenUpdates from closing the stream
					select {}
				},
			}
			mockStream.On("Close").Return(nil)
			mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("StreamID").Return(quic.StreamID(1))
			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

			// Close the stream if needed
			if tt.streamClosed {
				substr.close()
			}

			// Create mock send stream for the group stream
			mockSendStream := &MockQUICSendStream{}
			mockSendStream.On("Close").Return(nil)
			mockSendStream.On("StreamID").Return(quic.StreamID(1))

			// Create test stream using constructor
			testStream := newSendGroupStream(mockSendStream, GroupSequence(1))

			// Create mock open group function
			openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
				return testStream, nil
			}

			sender := newTrackSender(substr, openGroupFunc)

			// Add a test stream to the queue if stream is not closed
			if !tt.streamClosed {
				_, err := sender.OpenGroup(GroupSequence(1))
				assert.NoError(t, err, "Failed to open group for test setup")

				// Verify the stream was added
				sender.mu.Lock()
				initialQueueSize := len(sender.queue)
				sender.mu.Unlock()

				assert.Equal(t, 1, initialQueueSize, "Initial queue size should be 1")
			}

			err := sender.Close()

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
			} else {
				assert.NoError(t, err, "unexpected error")

				// Verify queue is cleared
				sender.mu.Lock()
				finalQueueSize := len(sender.queue)
				sender.mu.Unlock()

				assert.Equal(t, 0, finalQueueSize, "Queue size should be 0 after close")
			}
		})
	}
}

func TestTrackSender_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		streamClosed bool
		reason       SubscribeErrorCode
		expectError  bool
	}{
		"close with custom error": {
			streamClosed: false,
			reason:       SubscribeErrorCode(1),
			expectError:  false,
		},
		"close with zero error code": {
			streamClosed: false,
			reason:       SubscribeErrorCode(0),
			expectError:  false,
		}, "already closed": {
			streamClosed: true,
			reason:       SubscribeErrorCode(2),
			expectError:  false, // Should handle multiple close calls gracefully
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock receive subscribe stream
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					// Block indefinitely to prevent listenUpdates from closing the stream
					select {}
				},
			}
			mockStream.On("Close").Return(nil)
			mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("StreamID").Return(quic.StreamID(1))
			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

			// Close the stream if needed
			if tt.streamClosed {
				substr.close()
			}

			// Create mock send stream for test
			mockSendStream := &MockQUICSendStream{}
			mockSendStream.On("StreamID").Return(quic.StreamID(1))
			mockSendStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()

			// Create mock open group function
			openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
				return newSendGroupStream(mockSendStream, seq), nil
			}

			sender := newTrackSender(substr, openGroupFunc)

			// Add a test stream to the queue if stream is not closed
			if !tt.streamClosed {
				_, err := sender.OpenGroup(GroupSequence(1))
				assert.NoError(t, err, "Failed to open group for test setup")

				// Verify the stream was added
				sender.mu.Lock()
				initialQueueSize := len(sender.queue)
				sender.mu.Unlock()

				assert.Equal(t, 1, initialQueueSize, "Initial queue size should be 1")
			}

			err := sender.CloseWithError(tt.reason)

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
			} else {
				assert.NoError(t, err, "unexpected error")

				// Verify queue is cleared
				sender.mu.Lock()
				finalQueueSize := len(sender.queue)
				sender.mu.Unlock()

				assert.Equal(t, 0, finalQueueSize, "Queue size should be 0 after close with error")
			}
		})
	}
}

func TestTrackSender_ConcurrentOperations(t *testing.T) {
	// Create mock receive subscribe stream
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			// Block indefinitely to prevent listenUpdates from closing the stream
			select {}
		},
	}
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("StreamID").Return(quic.StreamID(1))
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

	// Keep track of streams for verification
	streams := make([]*sendGroupStream, 0)

	// Create open group function that remembers streams
	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		// Create mock send stream for each group
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("StreamID").Return(quic.StreamID(int(seq)))

		stream := newSendGroupStream(mockSendStream, seq)
		streams = append(streams, stream)
		return stream, nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	// Open multiple groups concurrently
	const numGroups = 10
	errChan := make(chan error, numGroups)
	for i := 0; i < numGroups; i++ {
		go func(seq GroupSequence) {
			_, err := sender.OpenGroup(seq)
			errChan <- err
		}(GroupSequence(i))
	}

	// Wait for all operations to complete
	for i := 0; i < numGroups; i++ {
		err := <-errChan
		assert.NoError(t, err, "Unexpected error opening group")
	}

	// Verify queue has correct number of streams
	sender.mu.Lock()
	queueSize := len(sender.queue)
	sender.mu.Unlock()

	assert.Equal(t, numGroups, queueSize, "Expected %d streams in queue, got %d", numGroups, queueSize)

	// Close the sender and verify all streams are closed
	err := sender.Close()
	assert.NoError(t, err, "Unexpected error closing track sender")

	// Verify queue is empty after close
	sender.mu.Lock()
	finalQueueSize := len(sender.queue)
	sender.mu.Unlock()

	assert.Equal(t, 0, finalQueueSize, "Queue size should be 0 after close")
}

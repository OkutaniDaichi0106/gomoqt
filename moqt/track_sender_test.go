package moqt

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

func TestNewTrackSender(t *testing.T) {
	// Create test contexts
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)
	trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

	// Create real queue
	queue := newOutgoingGroupStreamQueue()

	// Create mock open group function
	openGroupFunc := func(groupCtx *groupContext) (*sendGroupStream, error) {
		return &sendGroupStream{}, nil
	}

	sender := newTrackSender(trackCtx, queue, openGroupFunc)

	if sender == nil {
		t.Fatal("newTrackSender returned nil")
	}

	if sender.trackCtx != trackCtx {
		t.Error("trackCtx not set correctly")
	}

	if sender.groupQueue != queue {
		t.Error("groupQueue not set correctly")
	}

	if sender.openGroupFunc == nil {
		t.Error("openGroupFunc not set")
	}
}

func TestTrackSender_OpenGroup(t *testing.T) {
	tests := []struct {
		name           string
		contextClosed  bool
		openGroupError error
		seq            GroupSequence
		expectError    bool
	}{
		{
			name:          "successful open",
			contextClosed: false,
			seq:           GroupSequence(1),
			expectError:   false,
		},
		{
			name:          "context closed",
			contextClosed: true,
			seq:           GroupSequence(2),
			expectError:   true,
		},
		{
			name:           "open group error",
			contextClosed:  false,
			openGroupError: errors.New("mock error"),
			seq:            GroupSequence(3),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test contexts
			sessCtx := newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				NewParameters(),
				slog.Default(),
				nil,
			)
			trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

			if tt.contextClosed {
				trackCtx.cancel(ErrClosedTrack)
			}

			// Create real queue
			queue := newOutgoingGroupStreamQueue()

			// Create test stream
			testStream := &sendGroupStream{}

			// Create mock open group function
			openGroupFunc := func(groupCtx *groupContext) (*sendGroupStream, error) {
				if tt.openGroupError != nil {
					return nil, tt.openGroupError
				}
				// Return the test stream so we can verify it was added to the queue
				return testStream, nil
			}

			sender := newTrackSender(trackCtx, queue, openGroupFunc)

			// Initial queue should be empty
			queue.mu.Lock()
			initialQueueSize := len(queue.queue)
			queue.mu.Unlock()

			if initialQueueSize != 0 {
				t.Errorf("Initial queue size should be 0, got %d", initialQueueSize)
			}

			groupWriter, err := sender.OpenGroup(tt.seq)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if groupWriter != nil {
					t.Error("expected nil GroupWriter on error")
				}

				// Queue should still be empty on error
				queue.mu.Lock()
				queueSize := len(queue.queue)
				queue.mu.Unlock()
				if queueSize != 0 {
					t.Errorf("Queue size should be 0 after error, got %d", queueSize)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if groupWriter == nil {
					t.Error("expected non-nil GroupWriter")
				}

				// Verify that stream was added to the queue
				queue.mu.Lock()
				queueSize := len(queue.queue)
				hasTestStream := false
				for stream := range queue.queue {
					if stream == testStream {
						hasTestStream = true
						break
					}
				}
				queue.mu.Unlock()

				if queueSize != 1 {
					t.Errorf("Queue size should be 1 after success, got %d", queueSize)
				}
				if !hasTestStream {
					t.Error("Test stream not found in queue")
				}

				// Test that the group is removed when context is done
				grpCtx := groupWriter.(*sendGroupStream).groupCtx
				grpCtx.cancel(errors.New("test done"))

				// Small delay to allow goroutine to run
				time.Sleep(10 * time.Millisecond)

				// Verify that stream was removed from queue
				queue.mu.Lock()
				finalQueueSize := len(queue.queue)
				queue.mu.Unlock()

				if finalQueueSize != 0 {
					t.Errorf("Queue size should be 0 after context done, got %d", finalQueueSize)
				}
			}
		})
	}
}

func TestTrackSenderClose(t *testing.T) {
	tests := []struct {
		name          string
		contextClosed bool
		expectError   bool
	}{
		{
			name:          "successful close",
			contextClosed: false,
			expectError:   false,
		},
		{
			name:          "already closed",
			contextClosed: true,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test contexts
			sessCtx := newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				NewParameters(),
				slog.Default(),
				nil,
			)
			trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

			if tt.contextClosed {
				trackCtx.cancel(ErrClosedTrack)
			}

			// Create real queue
			queue := newOutgoingGroupStreamQueue()

			// Create mock open group function
			openGroupFunc := func(groupCtx *groupContext) (*sendGroupStream, error) {
				return &sendGroupStream{groupCtx: groupCtx}, nil
			}

			sender := newTrackSender(trackCtx, queue, openGroupFunc)

			// Add a test stream to the queue if context is not closed
			if !tt.contextClosed {
				testStream, _ := openGroupFunc(newGroupContext(trackCtx, GroupSequence(1)))
				queue.add(testStream)

				// Verify the stream was added
				queue.mu.Lock()
				initialQueueSize := len(queue.queue)
				queue.mu.Unlock()

				if initialQueueSize != 1 {
					t.Errorf("Initial queue size should be 1, got %d", initialQueueSize)
				}
			}

			err := sender.Close()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check that context is cancelled after close
				select {
				case <-trackCtx.Done():
					// Expected
					if cause := context.Cause(trackCtx); cause != ErrClosedGroup {
						t.Errorf("expected context cause to be ErrClosedGroup, got: %v", cause)
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("context should be cancelled after Close()")
				}

				// Verify queue is cleared
				queue.mu.Lock()
				finalQueueSize := len(queue.queue)
				queue.mu.Unlock()

				if finalQueueSize != 0 {
					t.Errorf("Queue size should be 0 after close, got %d", finalQueueSize)
				}
			}
		})
	}
}

func TestTrackSenderCloseWithError(t *testing.T) {
	tests := []struct {
		name          string
		contextClosed bool
		reason        error
		expectError   bool
	}{
		{
			name:          "close with custom error",
			contextClosed: false,
			reason:        errors.New("custom error"),
			expectError:   false,
		},
		{
			name:          "close with nil error",
			contextClosed: false,
			reason:        nil,
			expectError:   false,
		},
		{
			name:          "already closed",
			contextClosed: true,
			reason:        errors.New("test error"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test contexts
			sessCtx := newSessionContext(
				context.Background(),
				protocol.Version(0x1),
				"/test",
				NewParameters(),
				NewParameters(),
				slog.Default(),
				nil,
			)
			trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

			if tt.contextClosed {
				trackCtx.cancel(ErrClosedTrack)
			}

			// Create real queue
			queue := newOutgoingGroupStreamQueue()

			// Create mock open group function
			openGroupFunc := func(groupCtx *groupContext) (*sendGroupStream, error) {
				return &sendGroupStream{groupCtx: groupCtx}, nil
			}

			sender := newTrackSender(trackCtx, queue, openGroupFunc)

			// Add a test stream to the queue if context is not closed
			if !tt.contextClosed {
				testStream, _ := openGroupFunc(newGroupContext(trackCtx, GroupSequence(1)))
				queue.add(testStream)

				// Verify the stream was added
				queue.mu.Lock()
				initialQueueSize := len(queue.queue)
				queue.mu.Unlock()

				if initialQueueSize != 1 {
					t.Errorf("Initial queue size should be 1, got %d", initialQueueSize)
				}
			}

			err := sender.CloseWithError(tt.reason)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check that context is cancelled after close
				select {
				case <-trackCtx.Done():
					// Expected - check the cause
					expectedCause := tt.reason
					if expectedCause == nil {
						expectedCause = ErrClosedGroup
					}

					if cause := context.Cause(trackCtx); cause != expectedCause {
						t.Errorf("context cause = %v, want %v", cause, expectedCause)
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("context should be cancelled after CloseWithError()")
				}

				// Verify queue is cleared with correct error
				queue.mu.Lock()
				finalQueueSize := len(queue.queue)
				queue.mu.Unlock()

				if finalQueueSize != 0 {
					t.Errorf("Queue size should be 0 after close with error, got %d", finalQueueSize)
				}
			}
		})
	}
}

func TestTrackSenderConcurrentOperations(t *testing.T) {
	// Create test contexts
	sessCtx := newSessionContext(
		context.Background(),
		protocol.Version(0x1),
		"/test",
		NewParameters(),
		NewParameters(),
		slog.Default(),
		nil,
	)
	trackCtx := newTrackContext(sessCtx, SubscribeID(1), BroadcastPath("/test"), TrackName("test"))

	// Create real queue
	queue := newOutgoingGroupStreamQueue()

	// Keep track of streams for verification
	streams := make([]*sendGroupStream, 0)

	// Create open group function that remembers streams
	openGroupFunc := func(groupCtx *groupContext) (*sendGroupStream, error) {
		stream := &sendGroupStream{groupCtx: groupCtx}
		streams = append(streams, stream)
		return stream, nil
	}

	sender := newTrackSender(trackCtx, queue, openGroupFunc)

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
		if err := <-errChan; err != nil {
			t.Errorf("Unexpected error opening group: %v", err)
		}
	}

	// Verify queue has correct number of streams
	queue.mu.Lock()
	queueSize := len(queue.queue)
	queue.mu.Unlock()

	if queueSize != numGroups {
		t.Errorf("Expected %d streams in queue, got %d", numGroups, queueSize)
	}

	// Close the sender and verify all streams are closed
	err := sender.Close()
	if err != nil {
		t.Errorf("Unexpected error closing track sender: %v", err)
	}

	// Verify queue is empty after close
	queue.mu.Lock()
	finalQueueSize := len(queue.queue)
	queue.mu.Unlock()

	if finalQueueSize != 0 {
		t.Errorf("Queue size should be 0 after close, got %d", finalQueueSize)
	}
}

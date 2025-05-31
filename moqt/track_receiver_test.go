package moqt

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

func TestNewTrackReceiver(t *testing.T) {
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

	// Create mock queue
	queue := &incomingGroupStreamQueue{}

	receiver := newTrackReceiver(trackCtx, queue)

	if receiver == nil {
		t.Fatal("newTrackReceiver returned nil")
	}

	if receiver.trackCtx != trackCtx {
		t.Error("trackCtx not set correctly")
	}

	if receiver.groupQueue != queue {
		t.Error("groupQueue not set correctly")
	}
}

func TestTrackReceiver_AcceptGroup(t *testing.T) {
	tests := []struct {
		name            string
		setupQueue      func() *incomingGroupStreamQueue
		contextTimeout  time.Duration
		expectError     bool
		expectedErrType error
	}{
		{
			name: "successful accept",
			setupQueue: func() *incomingGroupStreamQueue {
				config := func() *SubscribeConfig {
					return &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100}
				}
				queue := newIncomingGroupStreamQueue(config)

				// Add a stream to the queue
				mockStream := &MockQUICReceiveStream{}
				stream := newReceiveGroupStream(SubscribeID(1), GroupSequence(42), mockStream)
				queue.enqueue(stream)

				return queue
			},
			contextTimeout: time.Second,
			expectError:    false,
		},
		{
			name: "timeout when no groups available",
			setupQueue: func() *incomingGroupStreamQueue {
				config := func() *SubscribeConfig {
					return &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100}
				}
				return newIncomingGroupStreamQueue(config)
			},
			contextTimeout:  50 * time.Millisecond,
			expectError:     true,
			expectedErrType: context.DeadlineExceeded,
		},
		{
			name: "context cancellation",
			setupQueue: func() *incomingGroupStreamQueue {
				config := func() *SubscribeConfig {
					return &SubscribeConfig{MinGroupSequence: 0, MaxGroupSequence: 100}
				}
				return newIncomingGroupStreamQueue(config)
			},
			contextTimeout:  time.Second,
			expectError:     true,
			expectedErrType: context.Canceled,
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

			// Setup queue
			queue := tt.setupQueue()
			receiver := newTrackReceiver(trackCtx, queue)

			// Create context with timeout or cancellation
			var ctx context.Context
			var cancel context.CancelFunc

			if tt.expectedErrType == context.Canceled {
				ctx, cancel = context.WithCancel(context.Background())
				// Cancel immediately for cancellation test
				cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), tt.contextTimeout)
				defer cancel()
			}

			groupReader, err := receiver.AcceptGroup(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.expectedErrType != nil && err != tt.expectedErrType {
					t.Errorf("expected error type %v, got %v", tt.expectedErrType, err)
				}
				if groupReader != nil {
					t.Error("expected nil GroupReader on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if groupReader == nil {
					t.Error("expected non-nil GroupReader")
				}

				// Verify the returned GroupReader
				expectedSeq := GroupSequence(42)
				if groupReader.GroupSequence() != expectedSeq {
					t.Errorf("GroupSequence() = %v, want %v", groupReader.GroupSequence(), expectedSeq)
				}
			}
		})
	}
}

func TestTrackReceiverClose(t *testing.T) {
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

	// Create mock queue
	queue := &incomingGroupStreamQueue{}

	receiver := newTrackReceiver(trackCtx, queue)

	// Context should not be done initially
	select {
	case <-trackCtx.Done():
		t.Error("context should not be done initially")
	default:
	}

	err := receiver.Close()
	if err != nil {
		t.Errorf("unexpected error from Close(): %v", err)
	}

	// Context should be done after close
	select {
	case <-trackCtx.Done():
		// Expected - check the cause
		if cause := context.Cause(trackCtx); cause != ErrClosedTrack {
			t.Errorf("context cause = %v, want %v", cause, ErrClosedTrack)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("context should be done after Close()")
	}
}

func TestTrackReceiverCloseWithError(t *testing.T) {
	tests := []struct {
		name   string
		reason error
	}{
		{
			name:   "close with custom error",
			reason: ErrInternalError,
		},
		{
			name:   "close with nil error",
			reason: nil,
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

			// Create mock queue
			queue := &incomingGroupStreamQueue{}

			receiver := newTrackReceiver(trackCtx, queue)

			// Context should not be done initially
			select {
			case <-trackCtx.Done():
				t.Error("context should not be done initially")
			default:
			}

			err := receiver.CloseWithError(tt.reason)
			if err != nil {
				t.Errorf("unexpected error from CloseWithError(): %v", err)
			}

			// Context should be done after close
			select {
			case <-trackCtx.Done():
				// Expected - check the cause
				if cause := context.Cause(trackCtx); cause != ErrClosedTrack {
					t.Errorf("context cause = %v, want %v", cause, ErrClosedTrack)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("context should be done after CloseWithError()")
			}
		})
	}
}

func TestTrackReceiverInterface(t *testing.T) {
	// Verify that trackReceiver implements TrackReader interface
	var _ TrackReader = (*trackReceiver)(nil)
}

func TestTrackReceiverAcceptGroupRealImplementation(t *testing.T) {
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
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(128),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	queue := newIncomingGroupStreamQueue(func() *SubscribeConfig { return config })

	receiver := newTrackReceiver(trackCtx, queue)

	// Test with a timeout to ensure we don't block forever
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := receiver.AcceptGroup(ctx)
	if err == nil {
		t.Error("expected timeout error when no groups are available")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded error, got: %v", err)
	}
}

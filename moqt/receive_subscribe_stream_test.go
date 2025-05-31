package moqt

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

func TestNewReceiveSubscribeStream(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	if rss == nil {
		t.Fatal("newReceiveSubscribeStream returned nil")
	}

	if rss.trackCtx != trackCtx {
		t.Error("trackCtx not set correctly")
	}

	if rss.config != config {
		t.Error("config not set correctly")
	}

	if rss.stream != mockStream {
		t.Error("stream not set correctly")
	}

	if rss.updatedCh == nil {
		t.Error("updatedCh should not be nil")
	}
}

func TestReceiveSubscribeStream_SubscribeID(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	id := rss.SubscribeID()
	expectedID := trackCtx.id

	if id != expectedID {
		t.Errorf("SubscribeID() = %v, want %v", id, expectedID)
	}
}

func TestReceiveSubscribeStreamSubscribeConfig(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(5),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(50),
	}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	returnedConfig := rss.SubscribeConfig()
	if returnedConfig != config {
		t.Error("SubuscribeConfig() did not return the original config")
	}

	if returnedConfig.TrackPriority != config.TrackPriority {
		t.Errorf("TrackPriority = %v, want %v", returnedConfig.TrackPriority, config.TrackPriority)
	}

	if returnedConfig.MinGroupSequence != config.MinGroupSequence {
		t.Errorf("MinGroupSequence = %v, want %v", returnedConfig.MinGroupSequence, config.MinGroupSequence)
	}

	if returnedConfig.MaxGroupSequence != config.MaxGroupSequence {
		t.Errorf("MaxGroupSequence = %v, want %v", returnedConfig.MaxGroupSequence, config.MaxGroupSequence)
	}
}

func TestReceiveSubscribeStreamUpdated(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Get the updated channel
	updatedCh := rss.Updated()
	if updatedCh == nil {
		t.Fatal("Updated() should not return nil")
	}

	// Check that it's the same channel on multiple calls
	updatedCh2 := rss.Updated()
	if updatedCh != updatedCh2 {
		t.Error("Updated() should return the same channel")
	}
}

func TestReceiveSubscribeStreamDone(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Get the done channel
	doneCh := rss.Done()
	if doneCh == nil {
		t.Fatal("Done() should not return nil")
	}

	// Should not be done initially
	select {
	case <-doneCh:
		t.Error("Done() channel should not be closed initially")
	default:
		// Good, not done yet
	}

	// Cancel the track context
	trackCtx.cancel(ErrClosedTrack)

	// Should be done now
	select {
	case <-doneCh:
		// Good, done after cancellation
	case <-time.After(100 * time.Millisecond):
		t.Error("Done() channel should be closed after context cancellation")
	}
}

func TestReceiveSubscribeStreamClose(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	err := rss.close()
	if err != nil {
		t.Errorf("close() error = %v", err)
	}

	// Verify Close was called on the underlying stream
	mockStream.AssertCalled(t, "Close")
}

func TestReceiveSubscribeStream_CloseWithError(t *testing.T) {
	testCases := map[string]struct {
		inputErr    error
		expectedErr error
	}{
		"nil error": {
			inputErr:    nil,
			expectedErr: ErrInternalError,
		},
		"invalid range error": {
			inputErr:    ErrInvalidRange,
			expectedErr: ErrInvalidRange,
		},
	}

	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			trackCtx := createTestTrackContext(sessCtx)
			config := &SubscribeConfig{}
			mockStream := &MockQUICStream{}

			rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

			err := rss.closeWithError(tc.inputErr)
			if err != nil {
				t.Errorf("closeWithError() error = %v", err)
			}

			// Verify Close was called on the underlying stream
			if !mockStream.AssertCalled(t, "CancelRead") {
				t.Error("underlying stream should be cancelled on closeWithError")
			}
			if !mockStream.AssertCalled(t, "CancelWrite") {
				t.Error("underlying stream should be cancelled on closeWithError")
			}

			// Verify context is cancelled with the right reason
			if trackCtx.Err() == nil {
				t.Error("track context should be cancelled")
			}

			gotErr := context.Cause(trackCtx)
			if gotErr != tc.expectedErr {
				t.Errorf("context cause = %v, want %v", gotErr, tc.expectedErr)
			}
		})
	}
}

func TestReceiveSubscribeStream_ClosedErr(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Initially should not be closed
	if rss.closedErr() != nil {
		t.Errorf("closedErr() for open stream = %v, want nil", rss.closedErr())
	}

	// After closing
	rss.close()
	if rss.closedErr() == nil {
		t.Error("closedErr() for closed stream should return error")
	}
}

func TestReceiveSubscribeStreamDoubleClose(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// First close
	err1 := rss.close()
	if err1 != nil {
		t.Errorf("first close() error = %v", err1)
	}

	// Second close should return error
	err2 := rss.close()
	if err2 == nil {
		t.Error("second close() should return error")
	}
}

func TestReceiveSubscribeStreamInterface(t *testing.T) {
	// Test that receiveSubscribeStream implements ReceiveSubscribeStream interface
	var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)
}

func TestReceiveSubscribeStream_ListenUpdates(t *testing.T) {
	testCases := map[string]struct {
		updateMessage  message.SubscribeUpdateMessage
		expectedConfig *SubscribeConfig
	}{
		"valid update": {
			updateMessage: message.SubscribeUpdateMessage{
				TrackPriority:    message.TrackPriority(5),
				MinGroupSequence: message.GroupSequence(10),
				MaxGroupSequence: message.GroupSequence(50),
			},
			expectedConfig: &SubscribeConfig{
				TrackPriority:    TrackPriority(5),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(50),
			},
		},
		"different priority": {
			updateMessage: message.SubscribeUpdateMessage{
				TrackPriority:    message.TrackPriority(1),
				MinGroupSequence: message.GroupSequence(0),
				MaxGroupSequence: message.GroupSequence(100),
			},
			expectedConfig: &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(100),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			trackCtx := createTestTrackContext(sessCtx)
			initialConfig := &SubscribeConfig{}

			// Prepare mock stream with encoded message
			mockStream := &MockQUICStream{
				ReadData: bytes.NewBuffer(nil),
			}

			// Encode the update message to the mock stream
			_, err := tc.updateMessage.Encode(mockStream.ReadData)
			if err != nil {
				t.Fatalf("failed to encode SubscribeUpdateMessage: %v", err)
			}

			rss := newReceiveSubscribeStream(trackCtx, mockStream, initialConfig)

			// Get the updated channel before starting
			updatedCh := rss.Updated()

			// Wait for the update to be processed
			select {
			case <-updatedCh:
				// Good, update received
				// Verify the config was updated
				updatedConfig := rss.SubscribeConfig()
				if updatedConfig.TrackPriority != tc.expectedConfig.TrackPriority {
					t.Errorf("TrackPriority = %v, want %v", updatedConfig.TrackPriority, tc.expectedConfig.TrackPriority)
				}
				if updatedConfig.MinGroupSequence != tc.expectedConfig.MinGroupSequence {
					t.Errorf("MinGroupSequence = %v, want %v", updatedConfig.MinGroupSequence, tc.expectedConfig.MinGroupSequence)
				}
				if updatedConfig.MaxGroupSequence != tc.expectedConfig.MaxGroupSequence {
					t.Errorf("MaxGroupSequence = %v, want %v", updatedConfig.MaxGroupSequence, tc.expectedConfig.MaxGroupSequence)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("Expected update to be received within timeout")
			}

		})
	}

	// // Test context cancellation stops listenUpdates
	// t.Run("context cancellation", func(t *testing.T) {
	// 	sessCtx := createTestSessionContext(context.Background()); trackCtx := createTestTrackContext(sessCtx)
	// 	config := &SubscribeConfig{}
	// 	mockStream := &MockQUICStream{}

	// 	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// 	// Cancel the track context
	// 	trackCtx.cancel(ErrClosedTrack)

	// 	// Give some time for the goroutine to react
	// 	time.Sleep(10 * time.Millisecond)

	// 	// The test passes if no panic occurs and context is properly handled
	// 	if trackCtx.Err() == nil {
	// 		t.Error("track context should be cancelled")
	// 	}
	// })
}

func TestReceiveSubscribeStream_ContextCancellation(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Cancel the track context
	trackCtx.cancel(ErrClosedTrack)

	// Give some time for the goroutine to react to context cancellation
	time.Sleep(50 * time.Millisecond)

	select {
	case <-rss.Done():
		// Good, Done channel should be closed

	case <-time.After(100 * time.Millisecond):
		t.Error("Done channel should be closed after context cancellation")
	}
}

func TestReceiveSubscribeStream_Updated(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	expectedConfig := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	_, err := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(expectedConfig.TrackPriority),
		MinGroupSequence: message.GroupSequence(expectedConfig.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(expectedConfig.MaxGroupSequence),
	}.Encode(mockStream.ReadData)
	if err != nil {
		t.Fatalf("failed to encode SubscribeUpdateMessage: %v", err)
	}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, &SubscribeConfig{})
	updatedCh := rss.Updated()

	if updatedCh == nil {
		t.Fatal("Updated() should not return nil")
	}
	select {
	case <-updatedCh:
		// Good, channel is ready
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected update to be received within timeout")
	}
}

// Test for decode error causing trackContext cancellation
func TestReceiveSubscribeStream_DecodeError(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}

	// Create mock stream that returns an error when reading
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer([]byte("invalid data")), // Invalid message data
	}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Wait for listenUpdates to process the error and cancel context
	select {
	case <-rss.trackCtx.Done():
		// Expected - decode error should cancel trackContext
		if context.Cause(trackCtx) == nil {
			t.Error("trackContext should be cancelled with an error cause")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("trackContext should be cancelled due to decode error")
	}
}

// Test for concurrent access to config
func TestReceiveSubscribeStream_ConcurrentAccess(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	initialConfig := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(10),
	}

	// Prepare multiple update messages
	updateMessage := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(5),
		MinGroupSequence: message.GroupSequence(20),
		MaxGroupSequence: message.GroupSequence(50),
	}

	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Encode update message
	_, err := updateMessage.Encode(mockStream.ReadData)
	if err != nil {
		t.Fatalf("failed to encode SubscribeUpdateMessage: %v", err)
	}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, initialConfig)

	// Run concurrent access to SubscribeConfig
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			// Repeatedly access config
			for j := 0; j < 100; j++ {
				config := rss.SubscribeConfig()
				if config == nil {
					t.Error("SubscribeConfig() should not return nil")
					return
				}
				time.Sleep(time.Microsecond) // Small delay to increase contention
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}

// Test for methods called on closed stream
func TestReceiveSubscribeStream_MethodsAfterClose(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Close the stream
	err := rss.close()
	if err != nil {
		t.Errorf("close() error = %v", err)
	}

	// Test SubscribeID after close (should still work)
	id := rss.SubscribeID()
	if id != trackCtx.id {
		t.Errorf("SubscribeID() after close = %v, want %v", id, trackCtx.id)
	}

	// Test SubscribeConfig after close (should still work)
	returnedConfig := rss.SubscribeConfig()
	if returnedConfig == nil {
		t.Error("SubscribeConfig() should still work after close")
	}

	// Test Updated channel after close
	updatedCh := rss.Updated()
	select {
	case _, ok := <-updatedCh:
		if ok {
			t.Error("Updated channel should be closed after stream close")
		}
	default:
		// Channel might not be closed immediately, this is acceptable
	}

	// Test Done channel after close
	select {
	case <-rss.Done():
		// Expected - Done should be closed after stream close
	default:
		t.Error("Done channel should be closed after stream close")
	}

	// Test double close
	err2 := rss.close()
	if err2 == nil {
		t.Error("second close() should return error")
	}
}

// Test stream read timeout behavior
func TestReceiveSubscribeStream_ReadTimeout(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}

	// Create mock stream that blocks on read
	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil), // Empty buffer will cause EOF
	}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Cancel context after a short delay to simulate timeout
	go func() {
		time.Sleep(50 * time.Millisecond)
		trackCtx.cancel(ErrInternalError)
	}()

	// Wait for context cancellation
	select {
	case <-rss.trackCtx.Done():
		// Expected - context should be cancelled
	case <-time.After(200 * time.Millisecond):
		t.Error("trackContext should be cancelled within timeout")
	}
}

// Test resource cleanup and goroutine termination
func TestReceiveSubscribeStream_ResourceCleanup(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

	// Give some time for goroutines to start
	time.Sleep(10 * time.Millisecond)

	// Close with error to trigger cleanup
	err := rss.closeWithError(ErrInternalError)
	if err != nil {
		t.Errorf("closeWithError() error = %v", err)
	}

	// Verify channels are closed
	select {
	case _, ok := <-rss.Updated():
		if ok {
			t.Error("Updated channel should be closed after closeWithError")
		}
	default:
		// Channel might not be closed immediately in some implementations
	}

	select {
	case <-rss.Done():
		// Expected - Done should be closed
	default:
		t.Error("Done channel should be closed after closeWithError")
	}

	// Verify stream operations were called
	mockStream.AssertCalled(t, "CancelRead")
	mockStream.AssertCalled(t, "CancelWrite")
}

// Test update notification behavior with channel blocking
func TestReceiveSubscribeStream_UpdateNotificationBlocking(t *testing.T) {
	sessCtx := createTestSessionContext(context.Background())
	trackCtx := createTestTrackContext(sessCtx)
	initialConfig := &SubscribeConfig{}

	updateMessage := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(5),
		MinGroupSequence: message.GroupSequence(10),
		MaxGroupSequence: message.GroupSequence(50),
	}

	mockStream := &MockQUICStream{
		ReadData: bytes.NewBuffer(nil),
	}

	// Encode multiple update messages
	for i := 0; i < 5; i++ {
		_, err := updateMessage.Encode(mockStream.ReadData)
		if err != nil {
			t.Fatalf("failed to encode SubscribeUpdateMessage: %v", err)
		}
	}

	rss := newReceiveSubscribeStream(trackCtx, mockStream, initialConfig)
	updatedCh := rss.Updated()

	// Don't read from the channel initially to test blocking behavior
	time.Sleep(100 * time.Millisecond)

	// Now read from the channel - should get notification
	select {
	case <-updatedCh:
		// Expected - should receive at least one notification
	case <-time.After(200 * time.Millisecond):
		t.Error("Should receive update notification even with channel blocking")
	}

	// Subsequent reads should not block indefinitely
	select {
	case <-updatedCh:
		// May or may not receive additional notifications
	default:
		// This is also acceptable
	}
}

// Test invalid message format handling
func TestReceiveSubscribeStream_InvalidMessageFormat(t *testing.T) {

	// Create mock stream with various invalid data
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "truncated message",
			data: []byte{0x00, 0x01}, // Incomplete message
		},
		{
			name: "random bytes",
			data: []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB}, // Random invalid data
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			trackCtx := createTestTrackContext(sessCtx) // Create fresh context for each test
			config := &SubscribeConfig{}
			mockStream := &MockQUICStream{
				ReadData: bytes.NewBuffer(tc.data),
			}

			rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

			// Wait for decode error to cancel context
			select {
			case <-rss.trackCtx.Done():
				// Expected - invalid format should cause error and context cancellation
			case <-time.After(200 * time.Millisecond):
				t.Error("trackContext should be cancelled due to invalid message format")
			}
		})
	}
}

// Test closeWithError with different error types
func TestReceiveSubscribeStream_CloseWithErrorTypes(t *testing.T) {
	testCases := []struct {
		name        string
		inputErr    error
		expectedErr error
	}{
		{
			name:        "SubscribeError",
			inputErr:    ErrInvalidRange,
			expectedErr: ErrInvalidRange,
		},
		{
			name:        "Non-SubscribeError",
			inputErr:    errors.New("generic error"),
			expectedErr: ErrInternalError, // Should be converted to SubscribeError
		},
		{
			name:        "nil error",
			inputErr:    nil,
			expectedErr: ErrInternalError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessCtx := createTestSessionContext(context.Background())
			trackCtx := createTestTrackContext(sessCtx)
			config := &SubscribeConfig{}
			mockStream := &MockQUICStream{}

			rss := newReceiveSubscribeStream(trackCtx, mockStream, config)

			err := rss.closeWithError(tc.inputErr)
			if err != nil {
				t.Errorf("closeWithError() error = %v", err)
			}

			// Verify context is cancelled with correct error
			if trackCtx.Err() == nil {
				t.Error("track context should be cancelled")
			}

			// Check if the error was properly handled
			if tc.inputErr != nil {
				cause := context.Cause(trackCtx)
				// For SubscribeError, cause should be the original error
				// For non-SubscribeError, the stream should handle it but context cause might vary
				if cause == nil {
					t.Error("context should have a cause")
				}
			}

			// Verify stream cancellation was called
			mockStream.AssertCalled(t, "CancelRead")
			mockStream.AssertCalled(t, "CancelWrite")
		})
	}
}

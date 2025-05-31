package moqt

import (
	"context"
	"sync"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
)

// func createTestTrackContext() *trackContext {
// 	sessCtx := createTestSessionContext(context.Background())
// 	return createTestTrackContext(sessCtx)
// }

func TestNewSendSubscribeStream(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	if sss == nil {
		t.Fatal("newSendSubscribeStream returned nil")
	}

	if sss.trackCtx != trackCtx {
		t.Error("trackCtx not set correctly")
	}

	if sss.config != config {
		t.Error("config not set correctly")
	}

	if sss.stream != mockStream {
		t.Error("stream not set correctly")
	}
}

func TestSendSubscribeStreamSubscribeID(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	id := sss.SubscribeID()
	expectedID := trackCtx.id

	if id != expectedID {
		t.Errorf("SubscribeID() = %v, want %v", id, expectedID)
	}
}

func TestSendSubscribeStreamSubscribeConfig(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(5),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(50),
	}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	returnedConfig := sss.SubscribeConfig()
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

func TestSendSubscribeStreamUpdateSubscribe(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	// Test valid update
	newConfig := &SubscribeConfig{
		TrackPriority:    TrackPriority(2),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(90),
	}

	err := sss.UpdateSubscribe(newConfig)
	if err != nil {
		t.Errorf("UpdateSubscribe() error = %v", err)
	}

	// Verify config was updated
	updatedConfig := sss.SubscribeConfig()
	if updatedConfig.TrackPriority != newConfig.TrackPriority {
		t.Errorf("TrackPriority after update = %v, want %v", updatedConfig.TrackPriority, newConfig.TrackPriority)
	}
}

func TestSendSubscribeStreamUpdateSubscribeInvalidRange(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	tests := []struct {
		name      string
		newConfig *SubscribeConfig
		wantError bool
	}{
		{
			name: "min > max",
			newConfig: &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(50),
				MaxGroupSequence: GroupSequence(30),
			},
			wantError: true,
		},
		{
			name: "decrease min when old min != 0",
			newConfig: &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(5), // less than original 10
				MaxGroupSequence: GroupSequence(100),
			},
			wantError: true,
		},
		{
			name: "increase max when old max != 0",
			newConfig: &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(200), // more than original 100
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sss.UpdateSubscribe(tt.newConfig)
			if tt.wantError && err == nil {
				t.Error("UpdateSubscribe() should return error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("UpdateSubscribe() error = %v", err)
			}
		})
	}
}

func TestSendSubscribeStreamClose(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	err := sss.close()
	if err != nil {
		t.Errorf("close() error = %v", err)
	}

	// Verify Close was called on the underlying stream
	mockStream.AssertCalled(t, "Close")
}

func TestSendSubscribeStreamCloseWithError(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	testErr := ErrInvalidRange
	err := sss.closeWithError(testErr)
	if err != nil {
		t.Errorf("closeWithError() error = %v", err)
	}

	// Verify Close was called on the underlying stream
	mockStream.AssertCalled(t, "Close")
}

func TestSendSubscribeStreamCloseWithNilError(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	err := sss.closeWithError(nil)
	if err != nil {
		t.Errorf("closeWithError(nil) error = %v", err)
	}

	// Should still close the stream
	mockStream.AssertCalled(t, "Close")
}

func TestSendSubscribeStream_ConcurrentUpdate(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	// Test concurrent updates
	var wg sync.WaitGroup

	go func() {
		newConfig := &SubscribeConfig{
			TrackPriority:    TrackPriority(2),
			MinGroupSequence: GroupSequence(5),
			MaxGroupSequence: GroupSequence(95),
		}
		_ = sss.UpdateSubscribe(newConfig)
		wg.Done()
	}()

	go func() {
		newConfig := &SubscribeConfig{
			TrackPriority:    TrackPriority(3),
			MinGroupSequence: GroupSequence(10),
			MaxGroupSequence: GroupSequence(90),
		}
		_ = sss.UpdateSubscribe(newConfig)
		wg.Done()
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Both updates should have completed without crashing
	// The final config should be one of the two updates
	finalConfig := sss.SubscribeConfig()
	if finalConfig.TrackPriority != TrackPriority(2) && finalConfig.TrackPriority != TrackPriority(3) {
		t.Errorf("unexpected final TrackPriority = %v", finalConfig.TrackPriority)
	}
}

func TestSendSubscribeStream_ContextCancellation(t *testing.T) {
	trackCtx := createTestTrackContext(createTestSessionContext(context.Background()))
	config := &SubscribeConfig{}
	mockStream := &MockQUICStream{}

	sss := newSendSubscribeStream(trackCtx, mockStream, config, moqtrace.DefaultQUICStreamAccepted(0))

	expectedReason := ErrClosedTrack
	// Cancel the track context
	trackCtx.cancel(expectedReason)

	select {
	case <-sss.trackCtx.Done():
		// Context should be cancelled
		if err := sss.trackCtx.Err(); err != context.Canceled {
			t.Errorf("track context should be cancelled, got %v", err)
		}
		if reason := context.Cause(sss.trackCtx); reason != expectedReason {
			t.Errorf("track context cause = %v, want %v", reason, expectedReason)
		}
	default:
	}
}

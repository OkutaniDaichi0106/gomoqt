package moqt

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSendSubscribeStream(t *testing.T) {
	id := SubscribeID(123)
	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sss := newSendSubscribeStream(id, mockStream, config)

	assert.NotNil(t, sss, "newSendSubscribeStream should not return nil")
	assert.Equal(t, id, sss.id, "id should be set correctly")
	assert.Equal(t, config, sss.config, "config should be set correctly")
	assert.Equal(t, mockStream, sss.stream, "stream should be set correctly")
	assert.False(t, sss.ctx.Err() != nil, "stream should not be closed initially")
}

func TestSendSubscribeStream_SubscribeID(t *testing.T) {
	id := SubscribeID(456)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sss := newSendSubscribeStream(id, mockStream, config)

	returnedID := sss.SubscribeID()

	assert.Equal(t, id, returnedID, "SubscribeID() should return the correct ID")
}

func TestSendSubscribeStream_SubscribeConfig(t *testing.T) {
	id := SubscribeID(789)
	config := &TrackConfig{
		TrackPriority:    TrackPriority(5),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(50),
	}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sss := newSendSubscribeStream(id, mockStream, config)

	returnedConfig := sss.TrackConfig()
	assert.Equal(t, config, returnedConfig, "TrackConfig() should return the original config")
	assert.Equal(t, config.TrackPriority, returnedConfig.TrackPriority, "TrackPriority should match")
	assert.Equal(t, config.MinGroupSequence, returnedConfig.MinGroupSequence, "MinGroupSequence should match")
	assert.Equal(t, config.MaxGroupSequence, returnedConfig.MaxGroupSequence, "MaxGroupSequence should match")
}

func TestSendSubscribeStream_UpdateSubscribe(t *testing.T) {
	id := SubscribeID(101)
	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	// Mock the Write method for UpdateSubscribe
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sss := newSendSubscribeStream(id, mockStream, config)

	// Test valid update
	newConfig := &TrackConfig{
		TrackPriority:    TrackPriority(2),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(90),
	}

	err := sss.UpdateSubscribe(newConfig)
	assert.NoError(t, err, "UpdateSubscribe() should not return error for valid config")

	// Verify config was updated
	updatedConfig := sss.TrackConfig()
	assert.Equal(t, newConfig.TrackPriority, updatedConfig.TrackPriority, "TrackPriority should be updated")
	assert.Equal(t, newConfig.MinGroupSequence, updatedConfig.MinGroupSequence, "MinGroupSequence should be updated")
	assert.Equal(t, newConfig.MaxGroupSequence, updatedConfig.MaxGroupSequence, "MaxGroupSequence should be updated")

	mockStream.AssertExpectations(t)
}

func TestSendSubscribeStream_UpdateSubscribe_InvalidRange(t *testing.T) {
	id := SubscribeID(102)
	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	sss := newSendSubscribeStream(id, mockStream, config)

	tests := map[string]struct {
		newConfig *TrackConfig
		wantError bool
	}{
		"min > max": {
			newConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(50),
				MaxGroupSequence: GroupSequence(30),
			},
			wantError: true,
		},
		"decrease min when old min != 0": {
			newConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(5), // less than original 10
				MaxGroupSequence: GroupSequence(100),
			},
			wantError: true,
		},
		"increase max when old max != 0": {
			newConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(200), // more than original 100
			},
			wantError: true,
		},
		"nil config": {
			newConfig: nil,
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := sss.UpdateSubscribe(tt.newConfig)
			if tt.wantError {
				assert.Error(t, err, "UpdateSubscribe() should return error for %s", name)
			} else {
				assert.NoError(t, err, "UpdateSubscribe() should not return error for %s", name)
			}
		})
	}
}

func TestSendSubscribeStream_Close(t *testing.T) {
	id := SubscribeID(103)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	// Mock the Close method
	mockStream.On("Close").Return(nil)

	sss := newSendSubscribeStream(id, mockStream, config)

	err := sss.close()
	assert.NoError(t, err, "Close() should not return error")
	assert.True(t, sss.ctx.Err() != nil, "stream should be marked as closed")

	// Verify Close was called on the underlying stream
	mockStream.AssertCalled(t, "Close")
}

func TestSendSubscribeStream_CloseWithError(t *testing.T) {
	id := SubscribeID(104)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	// Mock StreamID method
	mockStream.On("StreamID").Return(quic.StreamID(1))
	// Mock CancelWrite and CancelRead methods
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()

	sss := newSendSubscribeStream(id, mockStream, config)

	testErrCode := InternalSubscribeErrorCode
	err := sss.closeWithError(testErrCode)
	assert.NoError(t, err, "CloseWithError() should not return error")
	assert.True(t, sss.ctx.Err() != nil, "stream should be marked as closed")
	assert.ErrorAs(t, Cause(sss.ctx), &SubscribeError{}, "closeErr should be a SubscribeError")

	// Verify CancelWrite and CancelRead were called on the underlying stream
	mockStream.AssertCalled(t, "CancelWrite", quic.StreamErrorCode(testErrCode))
	mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(testErrCode))
}

func TestSendSubscribeStream_CloseWithError_NilError(t *testing.T) {
	id := SubscribeID(105)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	// Mock StreamID method
	mockStream.On("StreamID").Return(quic.StreamID(1))
	// Mock CancelWrite and CancelRead methods
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()

	sss := newSendSubscribeStream(id, mockStream, config)

	testErrCode := SubscribeErrorCode(0) // Using zero error code
	err := sss.closeWithError(testErrCode)
	assert.NoError(t, err, "CloseWithError() should not return error")
	assert.True(t, sss.ctx.Err() != nil, "stream should be marked as closed")

	// Should still cancel the stream operations
	mockStream.AssertCalled(t, "CancelWrite", quic.StreamErrorCode(testErrCode))
	mockStream.AssertCalled(t, "CancelRead", quic.StreamErrorCode(testErrCode))
}

func TestSendSubscribeStream_ConcurrentUpdate(t *testing.T) {
	id := SubscribeID(106)
	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	// Mock Write method for UpdateSubscribe calls
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sss := newSendSubscribeStream(id, mockStream, config)

	// Test concurrent updates
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		newConfig := &TrackConfig{
			TrackPriority:    TrackPriority(2),
			MinGroupSequence: GroupSequence(5),
			MaxGroupSequence: GroupSequence(95),
		}
		err := sss.UpdateSubscribe(newConfig)
		if err != nil {
			t.Logf("First concurrent update failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		newConfig := &TrackConfig{
			TrackPriority:    TrackPriority(3),
			MinGroupSequence: GroupSequence(5), // Use same min to avoid conflict
			MaxGroupSequence: GroupSequence(90),
		}
		err := sss.UpdateSubscribe(newConfig)
		if err != nil {
			t.Logf("Second concurrent update failed: %v", err)
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Both updates should have completed without crashing
	// The final config should be one of the two updates
	finalConfig := sss.TrackConfig()
	assert.Contains(t, []TrackPriority{TrackPriority(2), TrackPriority(3)},
		finalConfig.TrackPriority, "Final config should be from one of the updates")
}

func TestSendSubscribeStream_ContextCancellation(t *testing.T) {
	id := SubscribeID(107)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	mockStream.On("Context").Return(ctx)

	sss := newSendSubscribeStream(id, mockStream, config)

	// Cancel the context
	cancel()

	// Check that the stream's context is properly cancelled
	select {
	case <-sss.ctx.Done():
		// Context should be cancelled
		assert.Error(t, sss.ctx.Err(), "context should have an error when cancelled")
	default:
		t.Error("context should be cancelled")
	}
}

func TestSendSubscribeStream_UpdateSubscribeWriteError(t *testing.T) {
	id := SubscribeID(108)
	config := &TrackConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	// Mock Write to return an error
	writeErr := assert.AnError
	mockStream.On("Write", mock.Anything).Return(0, writeErr)
	mockStream.On("StreamID").Return(quic.StreamID(1))
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()

	sss := newSendSubscribeStream(id, mockStream, config)

	newConfig := &TrackConfig{
		TrackPriority:    TrackPriority(2),
		MinGroupSequence: GroupSequence(10),
		MaxGroupSequence: GroupSequence(90),
	}

	err := sss.UpdateSubscribe(newConfig)
	assert.Error(t, err, "UpdateSubscribe() should return error when Write fails")
	assert.True(t, sss.ctx.Err() != nil, "stream should be marked as closed after write error")
	assert.Error(t, context.Cause(sss.ctx), "closeErr should be set")
	assert.ErrorAs(t, Cause(sss.ctx), &SubscribeError{}, "closeErr should be a SubscribeError")

	mockStream.AssertExpectations(t)
}

func TestSendSubscribeStream_UpdateSubscribeClosedStream(t *testing.T) {
	id := SubscribeID(109)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	mockStream.On("Close").Return(nil)

	sss := newSendSubscribeStream(id, mockStream, config)

	// Close the stream first
	err := sss.close()
	assert.NoError(t, err, "Close() should succeed")

	// Try to update after closing
	newConfig := &TrackConfig{
		TrackPriority: TrackPriority(1),
	}

	err = sss.UpdateSubscribe(newConfig)
	assert.Error(t, err, "UpdateSubscribe() should return error on closed stream")
	assert.Contains(t, err.Error(), "closed", "error should mention stream is closed")

	mockStream.AssertExpectations(t)
}

func TestSendSubscribeStream_CloseAlreadyClosed(t *testing.T) {
	id := SubscribeID(110)
	config := &TrackConfig{}
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	mockStream.On("Close").Return(nil).Once()

	sss := newSendSubscribeStream(id, mockStream, config)

	// Close once
	err1 := sss.close()
	assert.NoError(t, err1, "first Close() should succeed")
	assert.True(t, sss.ctx.Err() != nil, "stream should be marked as closed")

	// Close again
	err2 := sss.close()
	assert.NoError(t, err2, "second Close() should not return error")

	mockStream.AssertExpectations(t)
}

func TestSendSubscribeStream_CloseWithError_MultipleClose(t *testing.T) {

	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}

	mockStream.On("StreamID").Return(quic.StreamID(1))
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return().Once()
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return().Once()

	sss := newSendSubscribeStream(SubscribeID(111), mockStream, &TrackConfig{})
	// Close with error once
	testErrCode := InternalSubscribeErrorCode
	err1 := sss.closeWithError(testErrCode)
	assert.NoError(t, err1, "first CloseWithError() should succeed")
	assert.True(t, sss.ctx.Err() != nil, "stream should be marked as closed") // Close with error again
	err2 := sss.closeWithError(testErrCode)

	assert.Error(t, err2, "second CloseWithError() should return the existing error")
	assert.ErrorAs(t, Cause(sss.ctx), &SubscribeError{}, "closeErr should be a SubscribeError")

	mockStream.AssertExpectations(t)
}

func TestSendSubscribeStream_UpdateSubscribeValidRangeTransitions(t *testing.T) {
	id := SubscribeID(112)

	tests := map[string]struct {
		initialConfig *TrackConfig
		newConfig     *TrackConfig
		expectError   bool
		description   string
	}{
		"increase min when old min is 0": {
			initialConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(100),
			},
			newConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			expectError: false,
			description: "should allow increasing min when old min is 0",
		},
		"decrease max when old max is 0": {
			initialConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(0),
			},
			newConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(50),
			},
			expectError: false,
			description: "should allow setting max when old max is 0",
		},
		"valid range within bounds": {
			initialConfig: &TrackConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(10),
				MaxGroupSequence: GroupSequence(100),
			},
			newConfig: &TrackConfig{
				TrackPriority:    TrackPriority(2),
				MinGroupSequence: GroupSequence(20),
				MaxGroupSequence: GroupSequence(80),
			},
			expectError: false,
			description: "should allow valid range within bounds",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}
			if !tt.expectError {
				mockStream.On("Write", mock.Anything).Return(0, nil)
			}

			sss := newSendSubscribeStream(id, mockStream, tt.initialConfig)

			err := sss.UpdateSubscribe(tt.newConfig)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				// Verify config was updated
				updatedConfig := sss.TrackConfig()
				assert.Equal(t, tt.newConfig, updatedConfig, "config should be updated")
			}

			mockStream.AssertExpectations(t)
		})
	}
}

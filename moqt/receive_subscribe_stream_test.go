package moqt

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReceiveSubscribeStream(t *testing.T) {
	tests := map[string]struct {
		subscribeID SubscribeID
		config      *SubscribeConfig
	}{
		"valid creation": {
			subscribeID: SubscribeID(123),
			config: &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(100),
			},
		},
		"zero subscribe ID": {
			subscribeID: SubscribeID(0),
			config: &SubscribeConfig{
				TrackPriority:    TrackPriority(0),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(10),
			},
		},
		"large subscribe ID": {
			subscribeID: SubscribeID(4294967295),
			config: &SubscribeConfig{
				TrackPriority:    TrackPriority(255),
				MinGroupSequence: GroupSequence(1000),
				MaxGroupSequence: GroupSequence(2000),
			},
		},
		"nil config": {
			subscribeID: SubscribeID(1),
			config:      nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					// Return EOF to close the goroutine cleanly
					return 0, io.EOF
				},
			}

			rss := newReceiveSubscribeStream(tt.subscribeID, mockStream, tt.config)

			assert.NotNil(t, rss, "newReceiveSubscribeStream should not return nil")
			assert.Equal(t, tt.subscribeID, rss.SubscribeID(), "SubscribeID should match")
			assert.NotNil(t, rss.Updated(), "Updated channel should not be nil")

			// Wait for goroutine to process EOF and close
			select {
			case <-rss.Updated():
				// Channel closed due to EOF
			case <-time.After(100 * time.Millisecond):
				t.Log("Timeout waiting for update channel to close")
			}
		})
	}
}

func TestReceiveSubscribeStream_SubscribeID(t *testing.T) {
	tests := map[string]struct {
		subscribeID SubscribeID
	}{
		"minimum value": {
			subscribeID: SubscribeID(0),
		},
		"small value": {
			subscribeID: SubscribeID(1),
		},
		"medium value": {
			subscribeID: SubscribeID(1000),
		},
		"large value": {
			subscribeID: SubscribeID(1000000),
		},
		"maximum uint64": {
			subscribeID: SubscribeID(^uint64(0)),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}
			config := &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(100),
			}

			rss := newReceiveSubscribeStream(tt.subscribeID, mockStream, config)

			result := rss.SubscribeID()
			assert.Equal(t, tt.subscribeID, result, "SubscribeID should match expected value")

			// Wait for goroutine to close cleanly
			select {
			case <-rss.Updated():
			case <-time.After(100 * time.Millisecond):
			}
		})
	}
}

func TestReceiveSubscribeStream_SubscribeConfig(t *testing.T) {
	tests := map[string]struct {
		name   string
		config *SubscribeConfig
	}{
		"valid config": {
			config: &SubscribeConfig{
				TrackPriority:    TrackPriority(10),
				MinGroupSequence: GroupSequence(5),
				MaxGroupSequence: GroupSequence(100),
			},
		},
		"zero values": {
			config: &SubscribeConfig{
				TrackPriority:    TrackPriority(0),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(0),
			},
		},
		"nil config": {
			config: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subscribeID := SubscribeID(123)
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					return 0, io.EOF
				},
			}

			rss := newReceiveSubscribeStream(subscribeID, mockStream, tt.config)

			resultConfig, err := rss.SubscribeConfig()
			assert.NoError(t, err, "SubscribeConfig should not return error")

			if tt.config != nil {
				assert.NotNil(t, resultConfig, "SubscribeConfig should not be nil")
				assert.Equal(t, tt.config.TrackPriority, resultConfig.TrackPriority, "TrackPriority should match")
				assert.Equal(t, tt.config.MinGroupSequence, resultConfig.MinGroupSequence, "MinGroupSequence should match")
				assert.Equal(t, tt.config.MaxGroupSequence, resultConfig.MaxGroupSequence, "MaxGroupSequence should match")
			}

			// Wait for goroutine to close cleanly
			select {
			case <-rss.Updated():
			case <-time.After(100 * time.Millisecond):
			}
		})
	}
}

func TestReceiveSubscribeStream_SubscribeConfig_UnreadableState(t *testing.T) {
	tests := map[string]struct {
		name        string
		setupStream func() *receiveSubscribeStream
		expectError bool
		expectedErr error
	}{
		"unreadable with read error": {
			setupStream: func() *receiveSubscribeStream {
				mockStream := &MockQUICStream{}
				rss := &receiveSubscribeStream{
					id:         SubscribeID(123),
					stream:     mockStream,
					config:     &SubscribeConfig{TrackPriority: TrackPriority(1)},
					updatedCh:  make(chan struct{}, 1),
					unreadable: true,
					readErr: &SubscribeError{
						StreamError: &quic.StreamError{
							StreamID:  quic.StreamID(123),
							ErrorCode: quic.StreamErrorCode(InternalSubscribeErrorCode),
						},
					},
				}
				return rss
			},
			expectError: true,
		},
		"unreadable with EOF": {
			setupStream: func() *receiveSubscribeStream {
				mockStream := &MockQUICStream{}
				rss := &receiveSubscribeStream{
					id:         SubscribeID(123),
					stream:     mockStream,
					config:     &SubscribeConfig{TrackPriority: TrackPriority(1)},
					updatedCh:  make(chan struct{}, 1),
					unreadable: true,
					readErr:    nil,
				}
				return rss
			},
			expectError: true,
			expectedErr: io.EOF,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rss := tt.setupStream()

			config, err := rss.SubscribeConfig()

			if tt.expectError {
				assert.Error(t, err, "SubscribeConfig should return error when unreadable")
				assert.Nil(t, config, "Config should be nil when error occurs")
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err, "Should return expected error")
				}
			} else {
				assert.NoError(t, err, "SubscribeConfig should not return error")
				assert.NotNil(t, config, "Config should not be nil")
			}
		})
	}
}

func TestReceiveSubscribeStream_Updated(t *testing.T) {
	subscribeID := SubscribeID(123)
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	rss := newReceiveSubscribeStream(subscribeID, mockStream, config)

	updatedCh := rss.Updated()
	assert.NotNil(t, updatedCh, "Updated channel should not be nil")

	// Check that we can receive from the channel (should close due to EOF)
	select {
	case <-updatedCh:
		// Channel should be closed due to EOF
		t.Log("Update channel closed as expected")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected update channel to be closed due to EOF")
	}
}

func TestReceiveSubscribeStream_ListenUpdates_WithSubscribeUpdateMessage(t *testing.T) {
	subscribeID := SubscribeID(123)

	// Create a valid SubscribeUpdateMessage
	updateMsg := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(5),
		MinGroupSequence: message.GroupSequence(10),
		MaxGroupSequence: message.GroupSequence(50),
	}

	// Encode the message
	buf := &bytes.Buffer{}
	_, err := updateMsg.Encode(buf)
	require.NoError(t, err)

	data := buf.Bytes()
	readPos := 0

	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			if readPos < len(data) {
				n := copy(p, data[readPos:])
				readPos += n
				return n, nil
			}
			// After reading all data, return EOF
			return 0, io.EOF
		},
	}

	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	rss := newReceiveSubscribeStream(subscribeID, mockStream, config)

	// Wait for the update to be processed
	select {
	case <-rss.Updated():
		// Should receive update notification
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive update notification")
	}

	// Check that config was updated
	updatedConfig, err := rss.SubscribeConfig()
	if err == nil {
		assert.Equal(t, TrackPriority(5), updatedConfig.TrackPriority, "TrackPriority should be updated")
		assert.Equal(t, GroupSequence(10), updatedConfig.MinGroupSequence, "MinGroupSequence should be updated")
		assert.Equal(t, GroupSequence(50), updatedConfig.MaxGroupSequence, "MaxGroupSequence should be updated")
	}
}

func TestReceiveSubscribeStream_ListenUpdates_StreamError(t *testing.T) {
	subscribeID := SubscribeID(123)
	streamError := &quic.StreamError{
		StreamID:  quic.StreamID(123),
		ErrorCode: quic.StreamErrorCode(InternalSubscribeErrorCode),
	}

	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, streamError
		},
	}

	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	rss := newReceiveSubscribeStream(subscribeID, mockStream, config)

	// Wait for error to be processed
	select {
	case <-rss.Updated():
		// Channel should close due to error
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected update channel to close due to stream error")
	}

	// Verify error state
	_, err := rss.SubscribeConfig()
	assert.Error(t, err, "SubscribeConfig should return error after stream error")

	var subscribeErr *SubscribeError
	assert.True(t, errors.As(err, &subscribeErr), "Error should be SubscribeError")
	assert.Equal(t, streamError.ErrorCode, subscribeErr.StreamError.ErrorCode, "Error codes should match")
}

func TestReceiveSubscribeStream_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		errorCode SubscribeErrorCode
		expectErr bool
	}{
		"internal error": {
			errorCode: InternalSubscribeErrorCode,
			expectErr: false,
		},
		"invalid range error": {
			errorCode: InvalidRangeErrorCode,
			expectErr: false,
		},
		"track not found error": {
			errorCode: TrackNotFoundErrorCode,
			expectErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subscribeID := SubscribeID(123)
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					// Block to prevent automatic closure
					select {}
				},
			}
			mockStream.On("StreamID").Return(quic.StreamID(123))
			mockStream.On("CancelWrite", quic.StreamErrorCode(tt.errorCode)).Return()
			mockStream.On("CancelRead", quic.StreamErrorCode(tt.errorCode)).Return()

			config := &SubscribeConfig{
				TrackPriority:    TrackPriority(1),
				MinGroupSequence: GroupSequence(0),
				MaxGroupSequence: GroupSequence(100),
			}

			// Create stream manually to avoid goroutine interference
			rss := &receiveSubscribeStream{
				id:                  subscribeID,
				config:              config,
				stream:              mockStream,
				updatedCh:           make(chan struct{}, 1),
				subscribeCanceledCh: make(chan *SubscribeError, 1),
			}

			// Mark listenOnce as done to prevent goroutine from starting
			rss.listenOnce.Do(func() {})

			err := rss.closeWithError(tt.errorCode)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, rss.unreadable, "Stream should be marked as unreadable")
			assert.True(t, rss.unwritable, "Stream should be marked as unwritable")

			mockStream.AssertExpectations(t)
		})
	}
}

func TestReceiveSubscribeStream_CloseWithError_AlreadyClosed(t *testing.T) {
	subscribeID := SubscribeID(123)
	mockStream := &MockQUICStream{}

	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	// Create stream manually
	rss := &receiveSubscribeStream{
		id:         subscribeID,
		config:     config,
		stream:     mockStream,
		updatedCh:  make(chan struct{}, 1),
		unreadable: true, // Already closed
		unwritable: true, // Already closed
	}

	// Mark listenOnce as done
	rss.listenOnce.Do(func() {})

	err := rss.closeWithError(InternalSubscribeErrorCode)
	assert.NoError(t, err, "closeWithError should not return error when already closed")
}

func TestReceiveSubscribeStream_ConcurrentAccess(t *testing.T) {
	subscribeID := SubscribeID(123)
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			return 0, io.EOF
		},
	}
	config := &SubscribeConfig{
		TrackPriority:    TrackPriority(1),
		MinGroupSequence: GroupSequence(0),
		MaxGroupSequence: GroupSequence(100),
	}

	rss := newReceiveSubscribeStream(subscribeID, mockStream, config)

	// Test concurrent access to SubscribeID (should be safe as it's read-only)
	var wg sync.WaitGroup
	numGoroutines := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			id := rss.SubscribeID()
			assert.Equal(t, subscribeID, id)
		}()
	}

	// Test concurrent access to SubscribeConfig
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			config, err := rss.SubscribeConfig()
			// Either should succeed or fail consistently
			if err == nil {
				assert.NotNil(t, config)
			}
		}()
	}

	// Wait for all goroutines to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All concurrent accesses completed successfully
	case <-time.After(1 * time.Second):
		t.Error("Concurrent access test timed out")
	}

	// Clean up - wait for update channel to close
	select {
	case <-rss.Updated():
	case <-time.After(100 * time.Millisecond):
	}
}

func TestReceiveSubscribeStream_UpdateChannelBehavior(t *testing.T) {
	t.Run("channel closes on EOF", func(t *testing.T) {
		subscribeID := SubscribeID(123)
		mockStream := &MockQUICStream{
			ReadFunc: func(p []byte) (int, error) {
				return 0, io.EOF
			},
		}
		config := &SubscribeConfig{TrackPriority: TrackPriority(1)}

		rss := newReceiveSubscribeStream(subscribeID, mockStream, config)

		select {
		case <-rss.Updated():
			// Channel should close
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected update channel to close on EOF")
		}

		// Verify channel is closed by trying to receive again
		select {
		case _, ok := <-rss.Updated():
			assert.False(t, ok, "Channel should be closed")
		default:
			t.Error("Channel should be closed and ready to receive")
		}
	})

	t.Run("multiple updates sent to channel", func(t *testing.T) {
		subscribeID := SubscribeID(123)

		// Create multiple update messages
		updates := []message.SubscribeUpdateMessage{
			{
				TrackPriority:    message.TrackPriority(1),
				MinGroupSequence: message.GroupSequence(0),
				MaxGroupSequence: message.GroupSequence(10),
			},
			{
				TrackPriority:    message.TrackPriority(2),
				MinGroupSequence: message.GroupSequence(5),
				MaxGroupSequence: message.GroupSequence(15),
			},
		}

		buf := &bytes.Buffer{}
		for _, update := range updates {
			_, err := update.Encode(buf)
			require.NoError(t, err)
		}

		data := buf.Bytes()
		readPos := 0

		mockStream := &MockQUICStream{
			ReadFunc: func(p []byte) (int, error) {
				if readPos < len(data) {
					n := copy(p, data[readPos:])
					readPos += n
					return n, nil
				}
				return 0, io.EOF
			},
		}

		config := &SubscribeConfig{TrackPriority: TrackPriority(0)}
		rss := newReceiveSubscribeStream(subscribeID, mockStream, config)

		// Should receive multiple update notifications
		updateCount := 0
		for {
			select {
			case _, ok := <-rss.Updated():
				if !ok {
					// Channel closed
					assert.GreaterOrEqual(t, updateCount, 1, "Should receive at least one update")
					return
				}
				updateCount++
			case <-time.After(200 * time.Millisecond):
				t.Error("Timeout waiting for updates")
				return
			}
		}
	})
}

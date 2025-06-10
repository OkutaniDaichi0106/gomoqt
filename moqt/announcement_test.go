package moqt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAnnouncement(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		expected string
	}{
		"valid path": {
			path:     BroadcastPath("test/path"),
			expected: "test/path",
		},
		"empty path": {
			path:     BroadcastPath(""),
			expected: "",
		},
		"path with special characters": {
			path:     BroadcastPath("test/path with spaces/and-dashes_and.dots"),
			expected: "test/path with spaces/and-dashes_and.dots",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			assert.NotNil(t, announcement)
			assert.Equal(t, tt.path, announcement.path)
			assert.NotNil(t, announcement.ctx)
			assert.NotNil(t, announcement.cancel)
		})
	}
}

func TestAnnouncement_BroadcastPath(t *testing.T) {
	tests := map[string]struct {
		path     BroadcastPath
		expected BroadcastPath
	}{
		"standard path": {
			path:     BroadcastPath("test/path"),
			expected: BroadcastPath("test/path"),
		},
		"empty path": {
			path:     BroadcastPath(""),
			expected: BroadcastPath(""),
		},
		"complex path": {
			path:     BroadcastPath("complex/path/with/multiple/segments"),
			expected: BroadcastPath("complex/path/with/multiple/segments"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			result := announcement.BroadcastPath()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnnouncement_String(t *testing.T) {
	tests := map[string]struct {
		path           BroadcastPath
		shouldEnd      bool
		expectedActive string
		expectedEnded  string
	}{
		"standard path": {
			path:           BroadcastPath("test/path"),
			shouldEnd:      true,
			expectedActive: "{ announce_status: active, broadcast_path: test/path }",
			expectedEnded:  "{ announce_status: ended, broadcast_path: test/path }",
		},
		"empty path": {
			path:           BroadcastPath(""),
			shouldEnd:      true,
			expectedActive: "{ announce_status: active, broadcast_path:  }",
			expectedEnded:  "{ announce_status: ended, broadcast_path:  }",
		},
		"special characters": {
			path:           BroadcastPath("test/path with spaces/and-dashes_and.dots"),
			shouldEnd:      false,
			expectedActive: "{ announce_status: active, broadcast_path: test/path with spaces/and-dashes_and.dots }",
			expectedEnded:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			result := announcement.String()
			assert.Equal(t, tt.expectedActive, result)

			if tt.shouldEnd {
				announcement.End()
				result = announcement.String()
				assert.Equal(t, tt.expectedEnded, result)
			}
		})
	}
}

func TestAnnouncement_IsActive(t *testing.T) {
	tests := map[string]struct {
		path              BroadcastPath
		shouldEnd         bool
		expectedBeforeEnd bool
		expectedAfterEnd  bool
	}{
		"active announcement": {
			path:              BroadcastPath("test/path"),
			shouldEnd:         false,
			expectedBeforeEnd: true,
			expectedAfterEnd:  true,
		},
		"ended announcement": {
			path:              BroadcastPath("test/path"),
			shouldEnd:         true,
			expectedBeforeEnd: true,
			expectedAfterEnd:  false,
		},
		"empty path active": {
			path:              BroadcastPath(""),
			shouldEnd:         false,
			expectedBeforeEnd: true,
			expectedAfterEnd:  true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			assert.Equal(t, tt.expectedBeforeEnd, announcement.IsActive())

			if tt.shouldEnd {
				announcement.End()
				assert.Equal(t, tt.expectedAfterEnd, announcement.IsActive())
			}
		})
	}
}

func TestAnnouncement_End(t *testing.T) {
	tests := map[string]struct {
		path             BroadcastPath
		multipleEndCalls bool
		expectedAfterEnd bool
	}{
		"single end call": {
			path:             BroadcastPath("test/path"),
			multipleEndCalls: false,
			expectedAfterEnd: false,
		},
		"multiple end calls": {
			path:             BroadcastPath("test/path"),
			multipleEndCalls: true,
			expectedAfterEnd: false,
		},
		"empty path end": {
			path:             BroadcastPath(""),
			multipleEndCalls: false,
			expectedAfterEnd: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			assert.True(t, announcement.IsActive())

			announcement.End()
			if tt.multipleEndCalls {
				announcement.End()
				announcement.End()
			}

			assert.Equal(t, tt.expectedAfterEnd, announcement.IsActive())
		})
	}
}

func TestAnnouncement_AwaitEnd(t *testing.T) {
	tests := map[string]struct {
		path        BroadcastPath
		endDelay    time.Duration
		timeout     time.Duration
		expectClose bool
	}{
		"end after delay": {
			path:        BroadcastPath("test/path"),
			endDelay:    100 * time.Millisecond,
			timeout:     200 * time.Millisecond,
			expectClose: true,
		},
		"no end call": {
			path:        BroadcastPath("test/path"),
			endDelay:    0,
			timeout:     50 * time.Millisecond,
			expectClose: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			// Test that AwaitEnd returns a channel that is not closed initially
			select {
			case <-announcement.AwaitEnd():
				if !tt.expectClose {
					t.Error("Expected AwaitEnd() channel to not be closed initially")
				}
			default:
				// This is the expected behavior for non-closed channels
			}

			if tt.endDelay > 0 {
				go func() {
					time.Sleep(tt.endDelay)
					announcement.End()
				}()

				select {
				case <-announcement.AwaitEnd():
					assert.True(t, tt.expectClose, "Channel closed when not expected")
				case <-time.After(tt.timeout):
					assert.False(t, tt.expectClose, "Expected channel to be closed but timeout occurred")
				}
			}
		})
	}
}

func TestAnnouncement_WithCancelledContext(t *testing.T) {
	tests := map[string]struct {
		path              BroadcastPath
		cancelImmediately bool
		sleepDuration     time.Duration
		expectedActive    bool
	}{
		"cancel after creation": {
			path:              BroadcastPath("test/path"),
			cancelImmediately: false,
			sleepDuration:     10 * time.Millisecond,
			expectedActive:    false,
		},
		"already cancelled": {
			path:              BroadcastPath("test/path"),
			cancelImmediately: true,
			sleepDuration:     0,
			expectedActive:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tt.cancelImmediately {
				cancel()
			}

			announcement := NewAnnouncement(ctx, tt.path)

			if !tt.cancelImmediately {
				assert.True(t, announcement.IsActive())
				cancel()
				if tt.sleepDuration > 0 {
					time.Sleep(tt.sleepDuration)
				}
			}

			assert.Equal(t, tt.expectedActive, announcement.IsActive())
		})
	}
}

func TestAnnouncement_Fork(t *testing.T) {
	tests := map[string]struct {
		path             BroadcastPath
		endOriginal      bool
		expectedOriginal bool
		expectedForked   bool
	}{
		"fork active announcement": {
			path:             BroadcastPath("test/path"),
			endOriginal:      false,
			expectedOriginal: true,
			expectedForked:   true,
		},
		"fork then end original": {
			path:             BroadcastPath("test/path"),
			endOriginal:      true,
			expectedOriginal: false,
			expectedForked:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			original := NewAnnouncement(ctx, tt.path)

			forked := original.Fork()
			assert.Equal(t, original.path, forked.path)
			assert.True(t, original.IsActive())
			assert.True(t, forked.IsActive())

			if tt.endOriginal {
				original.End()
			}

			assert.Equal(t, tt.expectedOriginal, original.IsActive())
			assert.Equal(t, tt.expectedForked, forked.IsActive())
		})
	}
}

func TestAnnouncement_ForkWithCancelledParent(t *testing.T) {
	tests := map[string]struct {
		path                   BroadcastPath
		sleepDuration          time.Duration
		expectedOriginalActive bool
		expectedForkedActive   bool
	}{
		"fork from cancelled parent": {
			path:                   BroadcastPath("test/path"),
			sleepDuration:          10 * time.Millisecond,
			expectedOriginalActive: false,
			expectedForkedActive:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			original := NewAnnouncement(ctx, tt.path)

			cancel()
			time.Sleep(tt.sleepDuration)

			assert.Equal(t, tt.expectedOriginalActive, original.IsActive())

			forked := original.Fork()
			assert.Equal(t, tt.expectedForkedActive, forked.IsActive())
		})
	}
}

func TestAnnouncement_ConcurrentAwaitEnd(t *testing.T) {
	tests := map[string]struct {
		path          BroadcastPath
		numGoroutines int
		endDelay      time.Duration
		timeout       time.Duration
	}{
		"concurrent await end": {
			path:          BroadcastPath("test/path"),
			numGoroutines: 10,
			endDelay:      50 * time.Millisecond,
			timeout:       500 * time.Millisecond,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			announcement := NewAnnouncement(ctx, tt.path)

			results := make(chan bool, tt.numGoroutines)

			for i := 0; i < tt.numGoroutines; i++ {
				go func() {
					select {
					case <-announcement.AwaitEnd():
						results <- true
					case <-time.After(tt.timeout):
						results <- false
					}
				}()
			}

			time.Sleep(tt.endDelay)
			announcement.End()

			for i := 0; i < tt.numGoroutines; i++ {
				select {
				case result := <-results:
					assert.True(t, result, "Expected all goroutines to receive end signal")
				case <-time.After(200 * time.Millisecond):
					t.Error("Timeout waiting for goroutine to complete")
				}
			}
		})
	}
}

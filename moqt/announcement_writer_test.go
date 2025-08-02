package moqt

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAnnouncementWriter(t *testing.T) {
	mockStream := &MockQUICStream{}
	prefix := "/test/"
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)

	sas := newAnnouncementWriter(mockStream, prefix)

	require.NotNil(t, sas)
	assert.Equal(t, prefix, sas.prefix)
	assert.Equal(t, mockStream, sas.stream)
	assert.NotNil(t, sas.actives)
	assert.NotNil(t, sas.ctx)
	assert.NotNil(t, sas.initCh)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_Init(t *testing.T) {
	tests := map[string]struct {
		announcements    []*Announcement
		expectError      bool
		expectedActives  int
		expectedSuffixes []string
		setupMocks       func(*MockQUICStream)
	}{
		"empty initialization": {
			announcements:    []*Announcement{},
			expectError:      false,
			expectedActives:  0,
			expectedSuffixes: []string{},
			setupMocks: func(mockStream *MockQUICStream) {
				ctx := context.Background()
				mockStream.On("Context").Return(ctx)
				mockStream.On("Write", mock.Anything).Return(0, nil).Once() // For AnnounceInitMessage
			},
		},
		"single active announcement": {
			announcements:    nil, // Will be set in test
			expectError:      false,
			expectedActives:  1,
			expectedSuffixes: []string{"stream1"},
			setupMocks: func(mockStream *MockQUICStream) {
				ctx := context.Background()
				mockStream.On("Context").Return(ctx)
				mockStream.On("Write", mock.Anything).Return(0, nil).Once() // For AnnounceInitMessage
			},
		},
		"multiple active announcements": {
			announcements:    nil, // Will be set in test
			expectError:      false,
			expectedActives:  2,
			expectedSuffixes: []string{"stream1", "stream2"},
			setupMocks: func(mockStream *MockQUICStream) {
				ctx := context.Background()
				mockStream.On("Context").Return(ctx)
				mockStream.On("Write", mock.Anything).Return(0, nil).Once() // For AnnounceInitMessage
			},
		},
		"inactive announcement": {
			announcements:    nil, // Will be set in test
			expectError:      false,
			expectedActives:  0,
			expectedSuffixes: []string{},
			setupMocks: func(mockStream *MockQUICStream) {
				ctx := context.Background()
				mockStream.On("Context").Return(ctx)
				mockStream.On("Write", mock.Anything).Return(0, nil).Once() // For AnnounceInitMessage
			},
		},
		"write error": {
			announcements:   []*Announcement{},
			expectError:     true,
			expectedActives: 0,
			setupMocks: func(mockStream *MockQUICStream) {
				ctx := context.Background()
				mockStream.On("Context").Return(ctx)
				mockStream.On("Write", mock.Anything).Return(0, errors.New("write error")).Once()
			},
		},
		"invalid path announcement": {
			announcements:    nil, // Will be set in test
			expectError:      false,
			expectedActives:  0,
			expectedSuffixes: []string{},
			setupMocks: func(mockStream *MockQUICStream) {
				ctx := context.Background()
				mockStream.On("Context").Return(ctx)
				mockStream.On("Write", mock.Anything).Return(0, nil).Once() // For AnnounceInitMessage
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			ctx := context.Background()

			tt.setupMocks(mockStream)

			sas := newAnnouncementWriter(mockStream, "/test/")

			// Prepare announcements based on test case
			switch name {
			case "single active announcement":
				ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
				tt.announcements = []*Announcement{ann}
			case "multiple active announcements":
				ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
				ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))
				tt.announcements = []*Announcement{ann1, ann2}
			case "inactive announcement":
				ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
				ann.End() // Make it inactive
				tt.announcements = []*Announcement{ann}
			case "invalid path announcement":
				ann := NewAnnouncement(ctx, BroadcastPath("/other/stream1")) // Different prefix
				tt.announcements = []*Announcement{ann}
			}

			err := sas.init(tt.announcements)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sas.actives, tt.expectedActives)

				for _, suffix := range tt.expectedSuffixes {
					assert.Contains(t, sas.actives, suffix)
				}

				// Verify initCh is closed
				assert.Nil(t, sas.initCh)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestAnnouncementWriter_Init_OnlyOnce(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil).Once() // Should only be called once

	sas := newAnnouncementWriter(mockStream, "/test/")

	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	// Call init multiple times
	err1 := sas.init([]*Announcement{ann})
	err2 := sas.init([]*Announcement{}) // Second call should be ignored

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 1)
	assert.Contains(t, sas.actives, "stream1")

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_Init_StreamError(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	streamError := &quic.StreamError{
		StreamID:  quic.StreamID(123),
		ErrorCode: quic.StreamErrorCode(42),
	}

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, streamError).Once()

	sas := newAnnouncementWriter(mockStream, "/test/")

	err := sas.init([]*Announcement{})

	require.Error(t, err)
	var announceErr *AnnounceError
	assert.ErrorAs(t, err, &announceErr)
	assert.Equal(t, streamError, announceErr.StreamError)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_Init_DuplicateAnnouncements(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil) // Allow any number of Write calls

	sas := newAnnouncementWriter(mockStream, "/test/")

	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream1")) // Same suffix - should replace first

	err := sas.init([]*Announcement{ann1, ann2})

	assert.NoError(t, err)
	assert.Len(t, sas.actives, 1)
	assert.Contains(t, sas.actives, "stream1")
	assert.Equal(t, ann2, sas.actives["stream1"])

	// First announcement should be ended
	assert.False(t, ann1.IsActive())
	assert.True(t, ann2.IsActive())

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_Init_MultipleDifferentAnnouncements(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil) // Allow any number of Write calls

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Create two announcements with different paths
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

	err := sas.init([]*Announcement{ann1, ann2})

	assert.NoError(t, err)
	assert.Len(t, sas.actives, 2)
	assert.Contains(t, sas.actives, "stream1")
	assert.Contains(t, sas.actives, "stream2")
	assert.Equal(t, ann1, sas.actives["stream1"])
	assert.Equal(t, ann2, sas.actives["stream2"])

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_Init_DeadlockIssue(t *testing.T) {
	// This test verifies that init() with duplicate announcements doesn't cause deadlock
	// after the implementation was fixed to use goroutines in OnEnd callbacks.

	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream1")) // Same suffix - should replace first

	// This should not deadlock anymore
	err := sas.init([]*Announcement{ann1, ann2})
	assert.NoError(t, err)

	// Allow time for background processing of OnEnd callbacks
	time.Sleep(50 * time.Millisecond)

	assert.Len(t, sas.actives, 1)
	assert.Contains(t, sas.actives, "stream1")
	assert.Equal(t, ann2, sas.actives["stream1"])

	// First announcement should be ended
	assert.False(t, ann1.IsActive())
	assert.True(t, ann2.IsActive())

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_SendAnnouncement(t *testing.T) {
	tests := map[string]struct {
		prefix         string
		broadcastPath  string
		expectError    bool
		shouldBeActive bool
	}{
		"valid path": {
			prefix:         "/test/",
			broadcastPath:  "/test/stream1",
			expectError:    false,
			shouldBeActive: true,
		},
		"invalid path": {
			prefix:         "/test/",
			broadcastPath:  "/other/stream1",
			expectError:    true,
			shouldBeActive: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			ctx := context.Background()

			mockStream.On("Context").Return(ctx)
			if !tt.expectError {
				mockStream.On("Write", mock.Anything).Return(0, nil).Times(2) // One for init, one for SendAnnouncement
			} else {
				mockStream.On("Write", mock.Anything).Return(0, nil).Times(1) // Only for init
			}

			sas := newAnnouncementWriter(mockStream, tt.prefix)
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			// Initialize the AnnouncementWriter first
			err := sas.init([]*Announcement{})
			require.NoError(t, err)

			err = sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err)
				assert.Len(t, sas.actives, 0)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sas.actives, 1)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestAnnouncementWriter_SendAnnouncement_WriteError(t *testing.T) {
	tests := map[string]struct {
		writeError   error
		expectAnnErr bool
	}{
		"stream error": {
			writeError: &quic.StreamError{
				StreamID:  quic.StreamID(123),
				ErrorCode: quic.StreamErrorCode(42),
			},
			expectAnnErr: true,
		},
		"generic error": {
			writeError:   errors.New("generic write error"),
			expectAnnErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			ctx := context.Background()

			mockStream.On("Context").Return(ctx)
			mockStream.On("Write", mock.Anything).Return(0, nil).Once()           // For init
			mockStream.On("Write", mock.Anything).Return(0, tt.writeError).Once() // For SendAnnouncement

			sas := newAnnouncementWriter(mockStream, "/test/")
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

			// Initialize the AnnouncementWriter first
			err := sas.init([]*Announcement{})
			require.NoError(t, err)

			err = sas.SendAnnouncement(ann)

			assert.Error(t, err)

			if tt.expectAnnErr {
				var announceErr *AnnounceError
				assert.ErrorAs(t, err, &announceErr)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestAnnouncementWriter_Close(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.Anything).Return()

	sas := newAnnouncementWriter(mockStream, "/test/")

	err := sas.Close()

	assert.NoError(t, err)
	assert.Nil(t, sas.actives)
	assert.Nil(t, sas.initCh)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		errorCode AnnounceErrorCode
	}{
		"internal error": {
			errorCode: InternalAnnounceErrorCode,
		},
		"duplicated announce error": {
			errorCode: DuplicatedAnnounceErrorCode,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			ctx := context.Background()

			mockStream.On("Context").Return(ctx)
			mockStream.On("CancelWrite", quic.StreamErrorCode(tt.errorCode)).Return()
			mockStream.On("CancelRead", quic.StreamErrorCode(tt.errorCode)).Return()

			sas := newAnnouncementWriter(mockStream, "/test/")

			err := sas.CloseWithError(tt.errorCode)

			assert.NoError(t, err)
			assert.Nil(t, sas.actives)
			assert.Nil(t, sas.initCh)

			mockStream.AssertExpectations(t)
		})
	}
}

func TestAnnouncementWriter_SendAnnouncement_MultipleAnnouncements(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(3) // One for init, two for SendAnnouncement

	sas := newAnnouncementWriter(mockStream, "/test/")

	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	err1 := sas.SendAnnouncement(ann1)
	err2 := sas.SendAnnouncement(ann2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 2)
	assert.Contains(t, sas.actives, "stream1")
	assert.Contains(t, sas.actives, "stream2")
	assert.Equal(t, ann1, sas.actives["stream1"])
	assert.Equal(t, ann2, sas.actives["stream2"])

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_SendAnnouncement_ReplaceExisting(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream1")) // Same suffix

	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	err1 := sas.SendAnnouncement(ann1)
	assert.NoError(t, err1)

	// Send second announcement with same path - should replace the first
	err2 := sas.SendAnnouncement(ann2)
	assert.NoError(t, err2)

	// Should have only one active announcement (the newer one)
	assert.Len(t, sas.actives, 1)
	assert.Contains(t, sas.actives, "stream1")
	assert.Equal(t, ann2, sas.actives["stream1"])

	// First announcement should be ended
	assert.False(t, ann1.IsActive())
	assert.True(t, ann2.IsActive())

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_SendAnnouncement_SameInstance(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(2) // One for init, one for SendAnnouncement

	sas := newAnnouncementWriter(mockStream, "/test/")
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	err1 := sas.SendAnnouncement(ann)
	err2 := sas.SendAnnouncement(ann)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 1)
	assert.True(t, ann.IsActive())

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_AnnouncementEnd_BackgroundProcessing(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(3) // init, ACTIVE, and ENDED messages

	sas := newAnnouncementWriter(mockStream, "/test/")
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	err = sas.SendAnnouncement(ann)
	assert.NoError(t, err)
	assert.Len(t, sas.actives, 1)

	ann.End()

	// Allow time for background goroutine to process
	time.Sleep(100 * time.Millisecond)

	assert.Len(t, sas.actives, 0)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_BoundaryValues(t *testing.T) {
	tests := map[string]struct {
		prefix        string
		broadcastPath string
		expectError   bool
	}{
		"root prefix": {
			prefix:        "/",
			broadcastPath: "/stream1",
			expectError:   false,
		},
		"matching prefix path": {
			prefix:        "/test/",
			broadcastPath: "/test/",
			expectError:   false,
		},
		"different root": {
			prefix:        "/test/",
			broadcastPath: "/other/stream1",
			expectError:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{}
			ctx := context.Background()

			mockStream.On("Context").Return(ctx)
			if !tt.expectError {
				mockStream.On("Write", mock.Anything).Return(0, nil).Times(2) // One for init, one for SendAnnouncement
			} else {
				mockStream.On("Write", mock.Anything).Return(0, nil).Times(1) // Only for init
			}

			sas := newAnnouncementWriter(mockStream, tt.prefix)
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			// Initialize the AnnouncementWriter first
			err := sas.init([]*Announcement{})
			require.NoError(t, err)

			err = sas.SendAnnouncement(ann)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStream.AssertExpectations(t)
		})
	}
}

func TestAnnouncementWriter_Performance_LargeNumberOfAnnouncements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil) // Will handle init + multiple announcements

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	const numAnnouncements = 100 // Reduced for test efficiency

	start := time.Now()
	for i := 0; i < numAnnouncements; i++ {
		ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream%d", i)))
		err := sas.SendAnnouncement(ann)
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	t.Logf("Time to send %d announcements: %v", numAnnouncements, duration)
	assert.Len(t, sas.actives, numAnnouncements)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_CleanupResourceLeaks(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	// Create and end many announcements to test cleanup
	for i := 0; i < 10; i++ {
		ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream%d", i)))
		err := sas.SendAnnouncement(ann)
		assert.NoError(t, err)
		ann.End()
	}

	// Allow time for cleanup
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, sas.actives, 0)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_PartialCleanup(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	// Create multiple announcements
	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))
	ann3 := NewAnnouncement(ctx, BroadcastPath("/test/stream3"))

	assert.NoError(t, sas.SendAnnouncement(ann1))
	assert.NoError(t, sas.SendAnnouncement(ann2))
	assert.NoError(t, sas.SendAnnouncement(ann3))

	// End only some announcements
	ann2.End()

	// Allow time for background processing
	time.Sleep(100 * time.Millisecond)

	assert.Len(t, sas.actives, 2)
	assert.Contains(t, sas.actives, "stream1")
	assert.NotContains(t, sas.actives, "stream2")
	assert.Contains(t, sas.actives, "stream3")

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_ConcurrentAccess(t *testing.T) {
	// NOTE: This test may occasionally cause deadlocks in the current implementation
	// when multiple goroutines compete for the same suffix, as the implementation
	// can deadlock between mutex acquisition and OnEnd callback processing.
	// The implementation should be fixed to avoid holding the mutex while calling End().

	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	// Test concurrent access to DIFFERENT suffixes to avoid deadlock
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream_a_%d", i)))
			if err := sas.SendAnnouncement(ann); err != nil {
				errors <- err
				return
			}
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			ann := NewAnnouncement(ctx, BroadcastPath(fmt.Sprintf("/test/stream_b_%d", i)))
			if err := sas.SendAnnouncement(ann); err != nil {
				errors <- err
				return
			}
			time.Sleep(time.Microsecond)
		}
	}()

	<-done
	<-done

	close(errors)
	for err := range errors {
		t.Errorf("Unexpected error: %v", err)
	}

	assert.Len(t, sas.actives, 10) // Should have 10 different streams

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_ConcurrentAccess_SameSuffix_DeadlockRisk(t *testing.T) {
	// This test verifies that concurrent access to the same suffix doesn't cause deadlock
	// after the implementation was fixed to use goroutines in OnEnd callbacks.

	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	// Test concurrent access to the SAME suffix - this should no longer cause deadlock
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
			if err := sas.SendAnnouncement(ann); err != nil {
				errors <- err
				return
			}
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
			if err := sas.SendAnnouncement(ann); err != nil {
				errors <- err
				return
			}
			time.Sleep(time.Microsecond)
		}
	}()

	<-done
	<-done

	// Allow time for background processing
	time.Sleep(50 * time.Millisecond)

	close(errors)
	for err := range errors {
		t.Errorf("Unexpected error: %v", err)
	}

	assert.Len(t, sas.actives, 1)
	assert.Contains(t, sas.actives, "stream1")

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_MultipleClose(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Close").Return(nil).Times(2)                  // Allow multiple close calls
	mockStream.On("CancelRead", mock.Anything).Return().Times(2) // Allow multiple CancelRead calls

	sas := newAnnouncementWriter(mockStream, "/test/")

	err1 := sas.Close()
	assert.NoError(t, err1)
	assert.Nil(t, sas.actives)

	err2 := sas.Close()
	assert.NoError(t, err2)

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_Context(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)

	sas := newAnnouncementWriter(mockStream, "/test/")

	assert.NotNil(t, sas.Context())
	assert.NoError(t, sas.Context().Err())

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_StressTest_HeavyConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Initialize the AnnouncementWriter first
	err := sas.init([]*Announcement{})
	require.NoError(t, err)

	const numGoroutines = 50
	const numOperationsPerGoroutine = 20
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

	// Launch multiple goroutines that compete for the same suffix aggressively
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			for i := 0; i < numOperationsPerGoroutine; i++ {
				// Half of them use the same suffix, half use different suffixes
				var suffixPath string
				if i%2 == 0 {
					suffixPath = "/test/contested_stream" // Same suffix - high contention
				} else {
					suffixPath = fmt.Sprintf("/test/stream_%d_%d", goroutineID, i)
				}

				ann := NewAnnouncement(ctx, BroadcastPath(suffixPath))
				if err := sas.SendAnnouncement(ann); err != nil {
					errors <- err
					return
				}
				// Randomly end some announcements to trigger OnEnd callbacks
				if i%3 == 0 {
					ann.End()
				}
			}
		}(g)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Allow time for background processing
	time.Sleep(100 * time.Millisecond)

	close(errors)
	for err := range errors {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that we still have some active announcements
	assert.True(t, len(sas.actives) > 0, "Should have some active announcements remaining")

	mockStream.AssertExpectations(t)
}

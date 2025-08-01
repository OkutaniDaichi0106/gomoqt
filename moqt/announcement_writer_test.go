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
				mockStream.On("Write", mock.Anything).Return(0, nil)
			}

			sas := newAnnouncementWriter(mockStream, tt.prefix)
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			err := sas.SendAnnouncement(ann)

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
			mockStream.On("Write", mock.Anything).Return(0, tt.writeError)

			sas := newAnnouncementWriter(mockStream, "/test/")
			ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

			err := sas.SendAnnouncement(ann)

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
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	ann1 := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))
	ann2 := NewAnnouncement(ctx, BroadcastPath("/test/stream2"))

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

	err1 := sas.SendAnnouncement(ann1)
	err2 := sas.SendAnnouncement(ann2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, sas.actives, 1)
	assert.Equal(t, ann2, sas.actives["stream1"])
	assert.False(t, ann1.IsActive()) // First announcement should be ended

	mockStream.AssertExpectations(t)
}

func TestAnnouncementWriter_SendAnnouncement_SameInstance(t *testing.T) {
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

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
	mockStream.On("Write", mock.Anything).Return(0, nil).Times(2) // ACTIVE and ENDED messages

	sas := newAnnouncementWriter(mockStream, "/test/")
	ann := NewAnnouncement(ctx, BroadcastPath("/test/stream1"))

	err := sas.SendAnnouncement(ann)
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
				mockStream.On("Write", mock.Anything).Return(0, nil)
			}

			sas := newAnnouncementWriter(mockStream, tt.prefix)
			ann := NewAnnouncement(ctx, BroadcastPath(tt.broadcastPath))

			err := sas.SendAnnouncement(ann)

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
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

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
	mockStream := &MockQUICStream{}
	ctx := context.Background()

	mockStream.On("Context").Return(ctx)
	mockStream.On("Write", mock.Anything).Return(0, nil)

	sas := newAnnouncementWriter(mockStream, "/test/")

	// Test concurrent access to the same suffix
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
	mockStream.On("Close").Return(nil).Once()
	mockStream.On("CancelRead", mock.Anything).Return().Once()

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

// Package moqt provides announcement buffer testing functionality for the MOQT system.
package moqt

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAnnouncementBuffer tests the creation of a new announcement buffer.
func TestNewAnnouncementBuffer(t *testing.T) {
	t.Parallel()
	buffer := newAnnouncementsBuffer()

	assert.NotNil(t, buffer.announcements)
	assert.NotNil(t, buffer.cond)
	assert.False(t, buffer.closed)
	assert.Nil(t, buffer.closedErr)
}

// TestAnnouncementsBuffer_SendAnnouncements tests the SendAnnouncements method of the announcementsBuffer.
func TestAnnouncementsBuffer_SendAnnouncements(t *testing.T) {
	tc := map[string]func(t *testing.T){
		// Tests that announcements are successfully added to the buffer and verifies the content
		"send announcements successfully": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			announcement := NewAnnouncement(TrackPath("/test/path"))

			err := buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			assert.Len(t, buffer.announcements, 1)
			assert.Equal(t, announcement, buffer.announcements[0])
		},
		// Tests that sending an empty announcements slice works correctly and results in no changes to the buffer
		"send empty announcements": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			err := buffer.SendAnnouncements([]*Announcement{})
			require.NoError(t, err)

			assert.Len(t, buffer.announcements, 0)
		},
		// Tests that sending nil announcement is handled gracefully and doesn't add anything to the buffer
		"send nil announcement": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			err := buffer.SendAnnouncements([]*Announcement{nil})
			require.NoError(t, err)

			assert.Len(t, buffer.announcements, 0)
		},
		// Tests that attempting to send announcements to a closed buffer returns an error
		"send to closed buffer": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			buffer.Close()

			err := buffer.SendAnnouncements([]*Announcement{NewAnnouncement("/test")})
			require.Error(t, err)
			assert.Len(t, buffer.announcements, 0)
		},
		// Tests that sending to a buffer closed with error returns the specific error the buffer was closed with
		"send to closed buffer with error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			testErr := errors.New("test error")

			buffer.CloseWithError(testErr)

			err := buffer.SendAnnouncements([]*Announcement{NewAnnouncement("/test")})
			require.ErrorIs(t, err, testErr)
			assert.Len(t, buffer.announcements, 0)
		},
		// Tests that multiple announcements can be added to the buffer sequentially
		"send additional announcements": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			announcement1 := NewAnnouncement(TrackPath("/test/path1"))
			announcement2 := NewAnnouncement(TrackPath("/test/path2"))

			err := buffer.SendAnnouncements([]*Announcement{announcement1})
			require.NoError(t, err)

			err = buffer.SendAnnouncements([]*Announcement{announcement2})
			require.NoError(t, err)

			assert.Len(t, buffer.announcements, 2)
			assert.Equal(t, announcement1, buffer.announcements[0])
			assert.Equal(t, announcement2, buffer.announcements[1])
		},
		// Tests duplicate announcement handling - verifying that duplicates are not added multiple times
		"send duplicate announcements": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			announcement := NewAnnouncement(TrackPath("/test/path"))

			err := buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			err = buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			assert.Len(t, buffer.announcements, 1)
			assert.Equal(t, announcement, buffer.announcements[0])
			// TODO: How to test duplicate announcements?
		},
		// Tests that sending announcements after closing the buffer returns an error and leaves original content unchanged
		"send additional announcements after closing": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			announcement1 := NewAnnouncement(TrackPath("/test/path1"))
			announcement2 := NewAnnouncement(TrackPath("/test/path2"))

			err := buffer.SendAnnouncements([]*Announcement{announcement1})
			require.NoError(t, err)

			buffer.Close()

			err = buffer.SendAnnouncements([]*Announcement{announcement2})
			require.Error(t, err)

			assert.Len(t, buffer.announcements, 1)
			assert.Equal(t, announcement1, buffer.announcements[0])
		},
		// Tests that inactive (ended) announcements are not added to the buffer
		"send inactive announcement": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			announcement := NewAnnouncement(TrackPath("/test/path"))
			announcement.End()

			err := buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			assert.Len(t, buffer.announcements, 0)
		},
	}

	for name, fn := range tc {
		name, fn := name, fn
		t.Run(name, fn)
	}
}

// TestAnnouncementsBuffer_ServeAnnouncements tests the ServeAnnouncements method of the announcementsBuffer.
func TestAnnouncementsBuffer_ServeAnnouncements(t *testing.T) {
	tc := map[string]func(t *testing.T){
		// Tests that announcements are successfully delivered to the writer and verifies the announcement content
		"serve announcements successfully": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			var announced []*Announcement
			wg := sync.WaitGroup{}

			// Mock writer to capture announced announcements
			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					announced = append(announced, announcements...)
					wg.Done()
					return nil
				},
			}

			// Add initial announcements
			announcement := NewAnnouncement(TrackPath("/test/path"))
			err := buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			// Run ServeAnnouncements in a goroutine

			wg.Add(1)
			go func() {
				buffer.deliverAnnouncements(writer)
			}()

			wg.Wait()

			// Verify the results
			assert.Len(t, announced, 1)
			assert.Equal(t, announcement, announced[0])
		},
		// Tests error handling when the writer returns an error during announcement delivery
		// and verifies that the writer is properly closed with the error
		"serve with writer returning error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			testErr := errors.New("write error")
			var closed bool
			var closedWithErr error
			wg := sync.WaitGroup{}

			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					return testErr
				},
				CloseFunc: func() error {
					closed = true
					return nil
				},
				CloseWithErrorFunc: func(err error) error {
					closed = true
					closedWithErr = err
					return nil
				},
			}

			// Add initial announcements
			announcement := NewAnnouncement(TrackPath("/test/path"))
			err := buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			wg.Add(1)
			go func() {
				defer wg.Done()
				buffer.deliverAnnouncements(writer)

				// When the writer returns an error, it should exit the serve loop
			}()

			wg.Wait()

			// Verify the results
			assert.True(t, closed)
			assert.Equal(t, testErr, closedWithErr)
		},
	}

	for name, fn := range tc {
		name, fn := name, fn
		t.Run(name, fn)
	}
}

// TestAnnouncementsBuffer_ExitServeAnnouncements tests different scenarios where ServeAnnouncements
// should exit and verifies the proper exit behavior.
func TestAnnouncementsBuffer_ExitServeAnnouncements(t *testing.T) {
	tc := map[string]func(t *testing.T){
		// Tests the behavior when the buffer is closed while serving announcements
		// and verifies that the writer is properly closed
		"exit when buffer is closed": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			var closed bool
			var announced bool
			wg := sync.WaitGroup{}

			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					announced = true
					return nil
				},
				CloseFunc: func() error {
					closed = true
					return nil
				},
				CloseWithErrorFunc: func(err error) error {
					closed = true
					return nil
				},
			}

			// Run ServeAnnouncements in a goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				buffer.deliverAnnouncements(writer)
			}()

			// Close the buffer during serving
			buffer.Close()

			wg.Wait()

			// Verify the results
			assert.False(t, closed)
			assert.False(t, announced)
		},
		// Tests the behavior when the buffer is closed with an error while serving announcements
		// and verifies that the writer exits the serve loop
		// The writer is not neither closed nor closed with error
		"exit when buffer is closed with error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			var announced bool
			var closed bool
			wg := sync.WaitGroup{}

			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					announced = true
					return nil
				},
				CloseFunc: func() error {
					closed = true
					return nil
				},
				CloseWithErrorFunc: func(err error) error {
					closed = true
					return nil
				},
			}

			// Run ServeAnnouncements in a goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				buffer.deliverAnnouncements(writer)
			}()

			// Close the buffer with an error
			buffer.CloseWithError(errors.New("test error"))

			wg.Wait()

			// Verify the results
			assert.False(t, closed)
			assert.False(t, announced)
		},
		// Tests that ServeAnnouncements exits when the writer returns an error
		"exit when writer returns error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			testErr := errors.New("write error")
			var closed bool
			var announced bool
			wg := sync.WaitGroup{}

			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					announced = true
					return testErr
				},
				CloseFunc: func() error {
					closed = true
					return nil
				},
				CloseWithErrorFunc: func(err error) error {
					closed = true
					return nil
				},
			}

			// Add initial announcements
			announcement := NewAnnouncement(TrackPath("/test/path"))
			err := buffer.SendAnnouncements([]*Announcement{announcement})
			require.NoError(t, err)

			wg.Add(1)
			go func() {
				defer wg.Done()
				buffer.deliverAnnouncements(writer)
			}()

			wg.Wait()

			// Verify the results
			assert.True(t, announced)
			assert.False(t, closed)
		},
	}

	for name, fn := range tc {
		name, fn := name, fn
		t.Run(name, fn)
	}
}

// TestAnnouncementsBuffer_Close tests the Close method of the announcementsBuffer.
func TestAnnouncementsBuffer_Close(t *testing.T) {
	tc := map[string]func(t *testing.T){
		// Tests that closing an open buffer succeeds and marks the buffer as closed
		"close successfully": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			err := buffer.Close()
			require.NoError(t, err)
			assert.True(t, buffer.closed)
		},
		// Tests that attempting to close an already closed buffer returns an appropriate error
		"close already closed buffer": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			buffer.closed = true

			err := buffer.Close()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already closed")
		},
		// Tests that closing a buffer with an error works correctly and stores the error
		"close with error successfully": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			testErr := errors.New("test error")

			err := buffer.CloseWithError(testErr)
			require.NoError(t, err)

			assert.True(t, buffer.closed)
			assert.Equal(t, testErr, buffer.closedErr)
		},
		// Tests that attempting to close an already closed buffer with error returns an appropriate error
		"close already closed buffer with error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			buffer.Close()

			err := buffer.CloseWithError(errors.New("test error"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already closed")
		},
		"close and end all announcements": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			announcement1 := NewAnnouncement(TrackPath("/test/path1"))
			announcement2 := NewAnnouncement(TrackPath("/test/path2"))
			err := buffer.SendAnnouncements([]*Announcement{announcement1, announcement2})
			require.NoError(t, err)

			err = buffer.Close()
			require.NoError(t, err)

			assert.True(t, buffer.closed)
			assert.False(t, announcement1.IsActive())
			assert.False(t, announcement2.IsActive())
		},
	}

	for name, fn := range tc {
		name, fn := name, fn
		t.Run(name, fn)
	}
}

// TestAnnouncementsBuffer_CloseWithError tests the CloseWithError method of the announcementsBuffer.
func TestAnnouncementsBuffer_CloseWithError(t *testing.T) {
	tc := map[string]func(t *testing.T){
		// Tests that closing with a specific error works correctly and stores the error
		"close with error successfully": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			testErr := errors.New("test error")

			err := buffer.CloseWithError(testErr)
			require.NoError(t, err)

			assert.True(t, buffer.closed)
			assert.Equal(t, testErr, buffer.closedErr)
		},
		// Tests that closing with nil error substitutes a default internal error
		"close with nil error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			err := buffer.CloseWithError(nil)
			require.NoError(t, err)

			assert.True(t, buffer.closed)
			assert.Equal(t, ErrInternalError, buffer.closedErr)
		},
		// Tests that attempting to close an already closed buffer with error returns an appropriate error
		"close already closed buffer": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			buffer.closed = true

			err := buffer.CloseWithError(errors.New("new error"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already closed")
		},
		// Tests that closing a buffer that was already closed with an error returns an appropriate error message
		"close already closed buffer with error": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()
			originalErr := errors.New("original error")
			buffer.closed = true
			buffer.closedErr = originalErr

			err := buffer.CloseWithError(errors.New("new error"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already closed with error")
		},
	}

	for name, fn := range tc {
		name, fn := name, fn
		t.Run(name, fn)
	}
}

func TestAnnouncementsBuffer_DeliverAnnouncements(t *testing.T) {
	tc := map[string]func(t *testing.T){
		"deliver all unique announcements": func(t *testing.T) {
			t.Parallel()
			// Initialize the buffer
			buffer := newAnnouncementsBuffer()

			var delivered []*Announcement
			wg := sync.WaitGroup{}

			// Mock writer to capture delivered announcements
			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					defer wg.Done()
					// Append the delivered announcements
					delivered = append(delivered, announcements...)

					return nil
				},
			}

			// Start the announcement delivery process in a separate goroutine
			// This simulates the normal operation where deliverAnnouncements runs continuously
			go func() {
				buffer.deliverAnnouncements(writer)
			}()

			// Test scenario: Send multiple announcements in different batches to verify proper handling

			// Step 1: Add the first batch of unique announcements to the buffer
			// The wg.Add(1) ensures we wait for these announcements to be delivered before proceeding
			wg.Add(1)
			announcement1 := NewAnnouncement(TrackPath("/test/path1"))
			announcement2 := NewAnnouncement(TrackPath("/test/path2"))
			err := buffer.SendAnnouncements([]*Announcement{announcement1, announcement2})
			require.NoError(t, err)

			// Step 2: Send a duplicate announcement (identical reference to announcement1)
			// This tests that sending the same announcement reference multiple times doesn't cause issues
			// This doesn't trigger the SendAnnouncements method to be called again
			err = buffer.SendAnnouncements([]*Announcement{announcement1})
			require.NoError(t, err)

			// Step 3: Send another announcement with the same path but different instance
			// This tests the handling of functionally duplicate announcements (different objects, same path)
			// The wg.Add(1) ensures we wait for this announcement to be delivered before proceeding
			wg.Add(1)
			announcement2_dup := NewAnnouncement(TrackPath("/test/path2"))
			err = buffer.SendAnnouncements([]*Announcement{announcement2_dup})
			require.NoError(t, err)

			// Wait for all announcements to be processed by the mock writer before proceeding with verification
			wg.Wait()

			// Verification section:
			// The tests below validate the complete announcement delivery pipeline from buffer to client

			// Verify the correct number of announcements were delivered to the writer
			// We expect exactly 3 announcements because:
			// 1. The first batch of two unique announcements (announcement1, announcement2)
			// 2. The duplicate reference to announcement1 (which shouldn't trigger additional delivery)
			// 3. The new instance with the same path as announcement2 (announcement2_dup)
			assert.Len(t, delivered, 3, "Expected delivery of all unique announcement instances")

			// Verify the delivery order matches the send order, which is critical for clients
			// that rely on receiving announcements in the correct sequence
			assert.Equal(t, announcement1, delivered[0], "First announcement should be delivered first")
			assert.Equal(t, announcement2, delivered[1], "Second announcement should be delivered second")
			assert.Equal(t, announcement2_dup, delivered[2], "Duplicate announcement should be delivered last")

			// Verify the active state of announcements after delivery, which confirms
			// the buffer's duplicate handling mechanism is working correctly:
			// - announcement1: Remains active because it was referenced but not duplicated with a new instance
			// - announcement2: Marked inactive because a new instance with the same path was processed
			// - announcement2_dup: Remains active as it's the most recent instance for its path
			assert.True(t, delivered[0].IsActive(), "Original announcement1 should remain active")
			assert.False(t, delivered[1].IsActive(), "Original announcement2 should be inactive after duplicate was processed")
			assert.True(t, delivered[2].IsActive(), "Duplicate announcement2 should remain active")
		},
		"deliver announcements with pattern": func(t *testing.T) {
			t.Parallel()
			// Initialize buffer with a specific track pattern that will filter announcements
			// The "/test/**" pattern will only match paths that start with "/test/"

			buffer := newAnnouncementsBuffer()

			// Array to store delivered announcements for verification
			var delivered []*Announcement
			wg := sync.WaitGroup{}

			// Create a mock writer that captures all delivered announcements
			// This allows verification of which announcements pass the pattern filter
			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					defer wg.Done()
					delivered = append(delivered, announcements...)
					return nil
				},
			}

			go func() {
				buffer.deliverAnnouncements(writer)
			}()

			// Add announcements to the buffer
			wg.Add(1)
			announcement1 := NewAnnouncement(TrackPath("/test/path1"))
			announcement2 := NewAnnouncement(TrackPath("/test/path2"))
			announcement3 := NewAnnouncement(TrackPath("/other/path3"))
			err := buffer.SendAnnouncements([]*Announcement{announcement1, announcement2, announcement3})
			require.NoError(t, err)

			wg.Wait()

			// Verify the delivered announcements
			assert.Len(t, delivered, 2) // Only announcements matching the pattern should be delivered.
			assert.Equal(t, announcement1, delivered[0])
			assert.Equal(t, announcement2, delivered[1])
		},
		"deliver some duplicate announcements": func(t *testing.T) {
			t.Parallel()
			buffer := newAnnouncementsBuffer()

			var delivered []*Announcement
			wg := sync.WaitGroup{}

			// Mock writer to capture delivered announcements
			writer := &MockAnnouncementWriter{
				SendAnnouncementsFunc: func(announcements []*Announcement) error {
					defer wg.Done()

					// Append the delivered announcements
					delivered = append(delivered, announcements...)
					return nil
				},
			}

			go func() {
				buffer.deliverAnnouncements(writer)
			}()

			// Add announcements to the buffer.
			// First add announcements to the buffer
			// The first announcement should be delivered via the SendAnnouncements method
			wg.Add(1)
			announcement1 := NewAnnouncement(TrackPath("/test/path1"))
			announcement2 := NewAnnouncement(TrackPath("/test/path2"))
			err := buffer.SendAnnouncements([]*Announcement{announcement1, announcement2})
			require.NoError(t, err)

			// Send duplicate announcements already in the buffer
			// This should not trigger the SendAnnouncements method to be called again
			err = buffer.SendAnnouncements([]*Announcement{announcement1})
			require.NoError(t, err)

			// Send new duplicate announcements
			// The wg.Add(1) ensures we wait for these announcements to be delivered before proceeding
			wg.Add(1)
			announcement2_dup := NewAnnouncement(TrackPath("/test/path2"))
			err = buffer.SendAnnouncements([]*Announcement{announcement2_dup})
			require.NoError(t, err)

			wg.Wait()

			// Verify the delivered announcements
			assert.Len(t, delivered, 3) // All unique announcements should be delivered.
			assert.Equal(t, announcement1, delivered[0])
			assert.Equal(t, announcement2, delivered[1])
			assert.Equal(t, announcement2_dup, delivered[2])

			// Verify that the duplicate announcements are not delivered again
			assert.True(t, announcement1.IsActive())
			assert.False(t, announcement2.IsActive())
			assert.True(t, announcement2_dup.IsActive())
		},
	}

	for name, fn := range tc {
		name, fn := name, fn
		t.Run(name, fn)
	}
}

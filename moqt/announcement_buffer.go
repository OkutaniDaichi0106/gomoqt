package moqt

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// AnnouncementWriter インターフェースの実装を確認する
var _ AnnouncementWriter = (*announcementsBuffer)(nil)

// newAnnouncementsBuffer creates a new announcements buffer.
// The buffer stores announcements and allows efficient delivery to subscribers.
func newAnnouncementsBuffer() *announcementsBuffer {
	return &announcementsBuffer{
		announcements: make([]*Announcement, 0),
		mapping:       make(map[TrackPath]int),
		cond:          sync.NewCond(&sync.Mutex{}),
	}
}

// announcementsBuffer implements the AnnouncementWriter interface and provides a buffer for announcements.
// It manages announcement states and delivery to writers efficiently.
// The buffer must be closed using Close() or CloseWithError() when done to prevent resource leaks.
type announcementsBuffer struct {
	// mapping maps TrackPath to its index in the announcements slice
	mapping map[TrackPath]int

	// cond synchronizes access to announcements and signals when new announcements are added
	cond *sync.Cond

	// announcements stores active announcements in the buffer
	announcements []*Announcement

	// closed indicates if the buffer has been closed
	closed bool

	// closedErr stores the error the buffer was closed with, if any
	closedErr error
}

// SendAnnouncements adds announcements to the buffer.
// If the buffer is closed, this method returns an error.
// Only announcements that match the buffer's track pattern are added.
// If an announcement with the same path already exists in the buffer, the old announcement is ended and the new one is added.
// This method returns nil if all announcements were added successfully, even if some were filtered out due to pattern mismatch.
func (ab *announcementsBuffer) SendAnnouncements(announcements []*Announcement) error {
	if len(announcements) == 0 {
		return nil
	}

	// Check if the buffer is closed
	if ab.closed {
		if ab.closedErr == nil {
			return errors.New("announcement buffer is already closed")
		}
		return ab.closedErr
	}

	ab.cond.L.Lock()
	defer ab.cond.L.Unlock()

	defer ab.cond.Broadcast()

	addedCount := 0
	for _, ann := range announcements {
		if ann == nil {
			slog.Debug("skipping nil announcement")
			continue
		}

		path := ann.path

		// Check if the announcement already exists
		if pos, ok := ab.mapping[path]; ok {
			// Update the existing announcement
			old := ab.announcements[pos]
			if old == ann {
				// Skip if the announcement is the same
				slog.Debug("skipping duplicate announcement reference",
					"track_path", path.String(),
				)
				continue
			} else {
				// End the old announcement
				slog.Debug("ending previous announcement for path",
					"track_path", path.String(),
				)
				old.End()
				delete(ab.mapping, path)
			}
		}

		if ann.IsActive() {
			// Add the mapping
			ab.mapping[path] = len(ab.announcements)
			// Append the new active announcement
			ab.announcements = append(ab.announcements, ann)
			addedCount++
			slog.Debug("added announcement to buffer",
				"track_path", path.String(),
				"active", ann.IsActive())
		} else {
			slog.Debug("skipping inactive announcement",
				"track_path", path.String())
		}
	}

	if addedCount > 0 {
		slog.Debug("added announcements to buffer",
			"count", addedCount,
			"total_active", len(ab.announcements))
	}

	return nil
}

// deliverAnnouncements sends announcements to the provided writer until the buffer is closed.
// The caller must close the buffer using Close() or CloseWithError() when done to ensure this method exits.
// This method will block waiting for new announcements or buffer closure, so it should typically be run in a separate goroutine.
func (buf *announcementsBuffer) deliverAnnouncements(w AnnouncementWriter) {
	if w == nil {
		slog.Error("cannot deliver announcements to nil writer")
		return
	}

	buf.cond.L.Lock()
	defer buf.cond.L.Unlock()

	slog.Debug("started announcement delivery", "buffer_size", len(buf.announcements))
	pos := 0
	for {
		// Wait for new announcements or buffer closure
		for pos >= len(buf.announcements) && !buf.closed {
			buf.cond.Wait()
		}

		// Check if the buffer is closed
		if buf.closed {
			slog.Debug("buffer closed, stopping announcement delivery")
			// Exit the loop and serve no more announcements if the buffer is closed
			return
		}

		// Create a snapshot of announcements to deliver while minimizing lock time
		nextAnnouncements := make([]*Announcement, len(buf.announcements[pos:]))
		copy(nextAnnouncements, buf.announcements[pos:])
		nextPos := len(buf.announcements)

		slog.Debug("delivering announcements",
			"count", len(nextAnnouncements),
			"start_pos", pos,
			"total", nextPos,
		)

		// Unlock before sending to avoid deadlock if writer calls back into buffer
		buf.cond.L.Unlock()
		err := w.SendAnnouncements(nextAnnouncements)
		buf.cond.L.Lock()

		if err != nil {
			slog.Error("error sending announcements", "error", err)
			// Exit the loop and serve no more announcements if the writer returns an error
			return
		}

		// Update the position after successful send
		pos = nextPos
	}
}

// Close closes the buffer and ends all active announcements.
// After the buffer is closed, no more announcements can be added and deliverAnnouncements will exit.
// Returns an error if the buffer is already closed.
func (ab *announcementsBuffer) Close() error {
	// ab.cond.L.Lock()
	// defer ab.cond.L.Unlock()

	// if ab.closed {
	// 	return errors.New("announcement buffer is already closed")
	// }

	// defer ab.cond.Broadcast()

	// // End all active announcements before closing
	// for _, ann := range ab.announcements {
	// 	if ann.IsActive() {
	// 		err := ann.End()
	// 		if err != nil {
	// 			slog.Warn("failed to end announcement during buffer close",
	// 				"track_path", ann.TrackPath().String(),
	// 				"error", err)
	// 		}
	// 	}
	// }

	// ab.closed = true

	// // Clear the internal state to free resources
	// slog.Debug("closing announcement buffer", "announcement_count", len(ab.announcements))
	// ab.announcements = nil
	// ab.mapping = nil

	return nil
}

// CloseWithError closes the buffer with an error.
// After the buffer is closed, no more announcements can be added and deliverAnnouncements will exit.
// SendAnnouncements will return the error that was provided to this method.
// If err is nil, a default internal error will be used.
// Returns an error if the buffer is already closed.
func (ab *announcementsBuffer) CloseWithError(err error) error {
	ab.cond.L.Lock()
	defer ab.cond.L.Unlock()

	if ab.closed {
		if ab.closedErr != nil {
			return fmt.Errorf("announcement buffer is already closed with error: %w", ab.closedErr)
		}
		return errors.New("announcement buffer is already closed")
	}

	defer ab.cond.Broadcast()

	// End all active announcements before closing
	for _, ann := range ab.announcements {
		if ann.IsActive() {
			err := ann.End()
			if err != nil {
				slog.Warn("failed to end announcement during buffer close",
					"track_path", ann.TrackPath().String(),
					"error", err,
				)
			}
		}
	}

	// If no error is provided, use a default error
	if err == nil {
		err = ErrInternalError
	}

	ab.closedErr = err
	ab.closed = true

	return nil
}

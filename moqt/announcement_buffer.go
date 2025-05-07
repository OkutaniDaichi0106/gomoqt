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
	buf := announcementsBuffer{
		mapping:   make(map[TrackPath]int),
		announced: make([]*Announcement, 0),
		dests:     make(map[AnnouncementWriter]struct{}),
	}
	buf.cond = sync.NewCond(&buf.mu)

	return &buf
}

// announcementsBuffer implements the AnnouncementWriter interface and provides a buffer for announcements.
// It manages announcement states and delivery to writers efficiently.
// The buffer must be closed using Close() or CloseWithError() when done to prevent resource leaks.
type announcementsBuffer struct {
	mu sync.Mutex

	// cond synchronizes access to announcements and signals when new announcements are added
	cond *sync.Cond

	// announced stores active announced in the buffer
	// mapping maps TrackPath to its index in the announcements slice
	mapping map[TrackPath]int

	// announced stores active announced in the buffer
	announced []*Announcement

	// diff []*Announcement

	dests map[AnnouncementWriter]struct{}

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

	ab.mu.Lock()
	defer ab.mu.Unlock()

	// Check if the buffer is closed
	if ab.closed {
		if ab.closedErr == nil {
			return errors.New("announcement buffer is already closed")
		}
		return ab.closedErr
	}

	// Map the announcements
	var path TrackPath

	diff := make([]*Announcement, 0, len(announcements))
	// copy(diff, announcements)

	for _, ann := range announcements {
		if ann == nil {
			slog.Debug("skipping nil announcement")
			continue
		}

		path = ann.path

		// Check if the announcement already exists
		// If it does, end the old one and replace it with the new one
		if pos, ok := ab.mapping[path]; ok {
			// Update the existing announcement
			old := ab.announced[pos]
			if old == ann {
				// Skip if the announcement is the same
				slog.Debug("skipping duplicate announcement reference",
					"track_path", path.String(),
				)
				continue
			} else {
				// End the old announcement
				old.End()
			}
		}

		// Map the new announcement
		ab.announced = append(ab.announced, ann)
		ab.mapping[path] = len(ab.announced) - 1
		diff = append(diff, ann)
	}

	for w := range ab.dests {
		w.SendAnnouncements(diff)
	}

	return nil
}

// serveAnnouncements sends announcements to the provided writer until the buffer is closed.
// The caller must close the buffer using Close() or CloseWithError() when done to ensure this method exits.
// This method will block waiting for new announcements or buffer closure, so it should typically be run in a separate goroutine.
func (buf *announcementsBuffer) serveAnnouncements(w AnnouncementWriter) {
	if w == nil {
		slog.Error("cannot deliver announcements to nil writer")
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	if buf.closed {
		slog.Error("cannot deliver announcements to closed buffer")
		return
	}

	// Register the writer
	buf.dests[w] = struct{}{}
	defer delete(buf.dests, w)

	// Send initial announcements
	if len(buf.announced) == 0 {
		slog.Debug("no initial announcements to send")
		return
	}

	err := w.SendAnnouncements(buf.announced)
	if err != nil {
		slog.Error("error sending initial announcements", "error", err)
		return
	}

	for !buf.closed {
		buf.cond.Wait()
	}
}

// Close closes the buffer and ends all active announcements.
// After the buffer is closed, no more announcements can be added and deliverAnnouncements will exit.
// Returns an error if the buffer is already closed.
func (ab *announcementsBuffer) Close() error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed {
		return errors.New("announcement buffer is already closed")
	}

	// End all active announcements before closing
	for _, ann := range ab.announced {
		if ann.IsActive() {
			err := ann.End()
			if err != nil {
				slog.Warn("failed to end announcement during buffer close",
					"track_path", ann.TrackPath().String(),
					"error", err)
			}
		}
	}

	ab.closed = true

	// Clear the internal state to free resources
	slog.Debug("closing announcement buffer", "announcement_count", len(ab.announced))
	ab.announced = nil
	ab.mapping = nil

	ab.cond.Broadcast()

	return nil
}

// CloseWithError closes the buffer with an error.
// After the buffer is closed, no more announcements can be added and deliverAnnouncements will exit.
// SendAnnouncements will return the error that was provided to this method.
// If err is nil, a default internal error will be used.
// Returns an error if the buffer is already closed.
func (ab *announcementsBuffer) CloseWithError(err error) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed {
		if ab.closedErr != nil {
			return fmt.Errorf("announcement buffer is already closed with error: %w", ab.closedErr)
		}
		return errors.New("announcement buffer is already closed")
	}

	// End all active announcements before closing
	for _, ann := range ab.announced {
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

	ab.cond.Broadcast()

	return nil
}

package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

// newAnnouncementWriter creates a new AnnouncementWriter for the given stream and prefix.
func newAnnouncementWriter(stream quic.Stream, prefix prefix) *AnnouncementWriter {
	if !isValidPrefix(prefix) {
		panic("invalid prefix for AnnouncementWriter")
	}

	sas := &AnnouncementWriter{
		prefix:  prefix,
		stream:  stream,
		ctx:     context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeAnnounce),
		actives: make(map[suffix]*activeAnnouncement),
		initCh:  make(chan struct{}),
	}

	return sas
}

// AnnouncementWriter manages the sending of announcements over a QUIC stream.
// It handles initialization, sending active announcements, and cleanup.
type AnnouncementWriter struct {
	prefix prefix
	stream quic.Stream
	ctx    context.Context

	mu      sync.RWMutex
	actives map[suffix]*activeAnnouncement

	initCh   chan struct{}
	initOnce sync.Once
}

// init initializes the AnnouncementWriter with the given announcements.
// It sends an AnnounceInitMessage and sets up end handlers for active announcements.
func (aw *AnnouncementWriter) init(announcements map[*Announcement]struct{}) error {
	var err error
	aw.initOnce.Do(func() {
		if aw.ctx.Err() != nil {
			err = Cause(aw.ctx)
			return
		}

		actives := make(map[suffix]*activeAnnouncement)
		suffixes := make([]suffix, 0, len(announcements))

		for ann := range announcements {
			if !ann.IsActive() {
				continue
			}
			sfx, ok := ann.BroadcastPath().GetSuffix(aw.prefix)
			if !ok {
				continue
			}
			// Always replace with the latest active announcement for the suffix
			actives[sfx] = &activeAnnouncement{announcement: ann}
			suffixes = append(suffixes, sfx)
		}

		err = message.AnnounceInitMessage{
			Suffixes: suffixes,
		}.Encode(aw.stream)
		if err != nil {
			var strErr *quic.StreamError
			if errors.As(err, &strErr) {
				err = &AnnounceError{StreamError: strErr}
			}
			return
		}

		aw.actives = actives

		// Register end functions for each active announcement
		for sfx, active := range actives {
			aw.registerEndHandler(sfx, active.announcement)
		}
		close(aw.initCh)
	})
	return err
}

// registerEndHandler registers handlers for when the announcement ends.
// It sets up AfterFunc to clean up when the announcement becomes inactive.
func (aw *AnnouncementWriter) registerEndHandler(sfx suffix, ann *Announcement) {
	stop := ann.AfterFunc(func() {
		aw.mu.Lock()
		defer aw.mu.Unlock()
		current, exists := aw.actives[sfx]
		if exists && current.announcement == ann {
			delete(aw.actives, sfx)
			message.AnnounceMessage{
				AnnounceStatus: message.ENDED,
				TrackSuffix:    sfx,
			}.Encode(aw.stream)
		}
	})

	aw.actives[sfx].end = func() {
		if !stop() {
			return
		}
		aw.mu.Lock()
		defer aw.mu.Unlock()
		delete(aw.actives, sfx)
		message.AnnounceMessage{
			AnnounceStatus: message.ENDED,
			TrackSuffix:    sfx,
		}.Encode(aw.stream)
	}
}

// SendAnnouncement sends an announcement if it's active and not already sent.
// It replaces any existing announcement for the same suffix.
func (aw *AnnouncementWriter) SendAnnouncement(announcement *Announcement) error {
	// Wait for initialization to complete
	select {
	case <-aw.initCh:
		// Initialization complete
	case <-aw.ctx.Done():
		return Cause(aw.ctx)
	}

	if !announcement.IsActive() {
		return nil // No need to send inactive announcements
	}

	// Get suffix for this announcement
	suffix, ok := announcement.BroadcastPath().GetSuffix(aw.prefix)
	if !ok {
		return errors.New("moq: broadcast path with invalid prefix")
	}

	// Check for previous announcement and get old endFunc
	aw.mu.Lock()
	defer aw.mu.Unlock()

	active, exists := aw.actives[suffix]
	if exists && active.announcement == announcement {
		return nil // Already active, no need to re-announce
	}

	// If there's an existing announcement for this suffix, end it first
	if exists {
		// Call end function without holding the lock to avoid deadlock
		aw.mu.Unlock()
		active.end()
		aw.mu.Lock()
	}

	// Encode and send ACTIVE announcement
	err := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}.Encode(aw.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			return &AnnounceError{
				StreamError: strErr,
			}
		}

		return err
	}

	aw.actives[suffix] = &activeAnnouncement{announcement: announcement}
	aw.registerEndHandler(suffix, announcement)

	return nil
}

// Close gracefully closes the AnnouncementWriter, ending all active announcements
// and closing the underlying stream.
func (aw *AnnouncementWriter) Close() error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	// End all active announcements
	for _, active := range aw.actives {
		active.end()
	}
	aw.actives = nil

	return aw.stream.Close()
}

// CloseWithError closes the AnnouncementWriter with an error, ending all active announcements
// and canceling the stream with the specified error code.
func (aw *AnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	// End all active announcements
	for _, active := range aw.actives {
		active.end()
	}
	aw.actives = nil

	strErrCode := quic.StreamErrorCode(code)
	aw.stream.CancelWrite(strErrCode)
	aw.stream.CancelRead(strErrCode)

	return nil
}

// Context returns the context associated with the AnnouncementWriter.
func (aw *AnnouncementWriter) Context() context.Context {
	return aw.ctx
}

// activeAnnouncement represents an active announcement being managed by AnnouncementWriter.
type activeAnnouncement struct {
	announcement *Announcement
	end          func() // Function to clean up the activeAnnouncement
}

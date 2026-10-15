package moqt

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	"github.com/qumo-dev/gomoqt/moqt/internal/message"
	"github.com/qumo-dev/gomoqt/transport"
)

// newAnnouncementWriter creates a new AnnouncementWriter for the given stream and prefix.
func newAnnouncementWriter(stream transport.Stream, prefix prefix, localHopID uint64, excludeHop uint64, logger *slog.Logger) *AnnouncementWriter {
	if !isValidPrefix(prefix) {
		panic("invalid prefix for AnnouncementWriter")
	}

	sas := &AnnouncementWriter{
		prefix:     prefix,
		stream:     stream,
		ctx:        context.WithValue(stream.Context(), biStreamTypeCtxKey, message.StreamTypeAnnounce),
		actives:    make(map[suffix]*activeAnnouncement),
		initDone:   make(chan struct{}),
		logger:     logger,
		localHopID: localHopID,
		excludeHop: excludeHop,
	}

	return sas
}

// AnnouncementWriter manages the sending of announcements for a specified prefix.
// It handles initialization, sending active announcements, and cleanup.
type AnnouncementWriter struct {
	prefix     prefix
	stream     transport.Stream
	ctx        context.Context
	logger     *slog.Logger
	localHopID uint64 // this node's Hop ID to append when forwarding
	excludeHop uint64 // skip announcements whose HopIDs contain this value

	// allow, when set, reports whether the peer may see a broadcast path. An
	// announcement it refuses is held back in hidden until rescope allows it.
	allow func(BroadcastPath) bool

	mu      sync.RWMutex
	actives map[suffix]*activeAnnouncement
	// hidden maps each held-back announcement to the stop of the watch that
	// forgets it when it ends.
	hidden map[*Announcement]func() bool

	initDone chan struct{}
	initOnce sync.Once
	initErr  error
}

func (aw *AnnouncementWriter) logError(msg string, err error, args ...any) {
	if aw == nil || err == nil {
		return
	}
	if aw.logger != nil {
		aw.logger.Error(msg, append(args, "error", err)...)
	}
}

// shouldExclude returns true if the announcement should be skipped because
// its hop path — including this node's own Hop ID, which is the implicit
// trailing entry carried by ANNOUNCE_OK — contains the excludeHop value.
func (aw *AnnouncementWriter) shouldExclude(ann *Announcement) bool {
	if aw.excludeHop == 0 {
		return false
	}
	if aw.excludeHop == aw.localHopID {
		return true
	}
	return slices.Contains(ann.HopIDs(), aw.excludeHop)
}

// holdBack reports whether ann is outside what the peer may see, and if so
// keeps it in hidden for a later rescope. An allowed announcement is dropped
// from hidden.
func (aw *AnnouncementWriter) holdBack(ann *Announcement) bool {
	if aw.allow == nil {
		return false
	}
	allowed := aw.allow(ann.BroadcastPath())

	aw.mu.Lock()
	defer aw.mu.Unlock()
	stop, held := aw.hidden[ann]
	if allowed {
		if held {
			stop()
			delete(aw.hidden, ann)
		}
		return false
	}
	if !held {
		if aw.hidden == nil {
			aw.hidden = make(map[*Announcement]func() bool)
		}
		aw.hidden[ann] = ann.AfterFunc(func() {
			aw.mu.Lock()
			defer aw.mu.Unlock()
			delete(aw.hidden, ann)
		})
	}
	return true
}

// rescope re-applies allow after it changed: it ends each active announcement
// the peer may no longer see, holding it back, and announces each held-back
// one it now may.
func (aw *AnnouncementWriter) rescope() {
	if aw.allow == nil {
		return
	}
	select {
	case <-aw.initDone:
	case <-aw.ctx.Done():
		return
	}

	aw.mu.Lock()
	var ends []func()
	var withdrawn, shown []*Announcement
	for _, active := range aw.actives {
		if active.end != nil && !aw.allow(active.announcement.BroadcastPath()) {
			ends = append(ends, active.end)
			withdrawn = append(withdrawn, active.announcement)
		}
	}
	for ann := range aw.hidden {
		if aw.allow(ann.BroadcastPath()) {
			shown = append(shown, ann)
		}
	}
	aw.mu.Unlock()

	// The end functions take aw.mu themselves.
	for _, end := range ends {
		end()
	}
	for _, ann := range withdrawn {
		if ann.IsActive() {
			aw.holdBack(ann)
		}
	}
	for _, ann := range shown {
		if err := aw.SendAnnouncement(ann); err != nil {
			aw.logError("failed to send a newly allowed announcement", err)
			return
		}
	}
}

// dropHidden forgets every held-back announcement. Caller MUST hold aw.mu.
func (aw *AnnouncementWriter) dropHidden() {
	for _, stop := range aw.hidden {
		stop()
	}
	aw.hidden = nil
}

// buildHopIDs returns the Hop ID list to encode in the outgoing
// ANNOUNCE_BROADCAST message. This node's own Hop ID is NOT included: it is
// carried once in ANNOUNCE_OK as the implicit trailing entry, and the spec
// forbids repeating it as the last entry of the broadcast's list.
func (aw *AnnouncementWriter) buildHopIDs(ann *Announcement) []uint64 {
	return ann.HopIDs()
}

// init snapshots the currently active announcements, sends an ACTIVE AnnounceBroadcastMessage
// for each active track suffix on the announce stream, and sets up end handlers.
func (aw *AnnouncementWriter) init(announcements map[*Announcement]struct{}) error {
	var err error
	aw.initOnce.Do(func() {
		defer close(aw.initDone)

		if aw.ctx.Err() != nil {
			err = Cause(aw.ctx)
			aw.mu.Lock()
			aw.initErr = err
			aw.mu.Unlock()
			return
		}

		actives := make(map[suffix]*activeAnnouncement)

		for ann := range announcements {
			if !ann.IsActive() {
				continue
			}
			sfx, ok := ann.BroadcastPath().GetSuffix(aw.prefix)
			if !ok {
				continue
			}
			if aw.shouldExclude(ann) {
				continue
			}
			if aw.holdBack(ann) {
				continue
			}
			// Always replace with the latest active announcement for the suffix.
			actives[sfx] = &activeAnnouncement{announcement: ann}
		}

		aw.mu.Lock()
		aw.actives = actives
		aw.mu.Unlock()

		// ANNOUNCE_OK is sent exactly once, before any ANNOUNCE_BROADCAST,
		// carrying this node's Hop ID and the size of the initial set.
		err = message.AnnounceOkMessage{
			HopID:       aw.localHopID,
			ActiveCount: uint64(len(actives)),
		}.Encode(aw.stream)
		if err != nil {
			if strErr, ok := errors.AsType[*transport.StreamError](err); ok {
				err = &AnnounceError{StreamError: strErr}
			}
			aw.mu.Lock()
			aw.initErr = err
			aw.mu.Unlock()
			return
		}

		for sfx, active := range actives {
			err = message.AnnounceBroadcastMessage{
				AnnounceStatus:      message.ACTIVE,
				BroadcastPathSuffix: sfx,
				HopIDs:              aw.buildHopIDs(active.announcement),
			}.Encode(aw.stream)
			if err != nil {
				if strErr, ok := errors.AsType[*transport.StreamError](err); ok {
					err = &AnnounceError{StreamError: strErr}
				}
				aw.mu.Lock()
				aw.initErr = err
				aw.mu.Unlock()
				return
			}

			aw.mu.Lock()
			aw.registerEndHandler(sfx, active.announcement)
			aw.mu.Unlock()
		}
	})

	aw.mu.RLock()
	defer aw.mu.RUnlock()
	if err != nil {
		return err
	}
	return aw.initErr
}

// registerEndHandler registers handlers for when the announcement ends.
// It sets up AfterFunc to clean up when the announcement becomes inactive.
// Caller MUST hold aw.mu.
func (aw *AnnouncementWriter) registerEndHandler(sfx suffix, ann *Announcement) {
	stop := ann.AfterFunc(func() {
		aw.mu.Lock()
		defer aw.mu.Unlock()
		current, exists := aw.actives[sfx]
		if exists && current.announcement == ann {
			delete(aw.actives, sfx)
			err := message.AnnounceBroadcastMessage{
				AnnounceStatus:      message.ENDED,
				BroadcastPathSuffix: sfx,
			}.Encode(aw.stream)
			if err != nil {
				aw.logError("failed to encode ANNOUNCE end message", err)
			}
		}
	})

	aw.actives[sfx].end = func() {
		if !stop() {
			return
		}
		aw.mu.Lock()
		defer aw.mu.Unlock()
		delete(aw.actives, sfx)
		if err := (message.AnnounceBroadcastMessage{
			AnnounceStatus:      message.ENDED,
			BroadcastPathSuffix: sfx,
		}).Encode(aw.stream); err != nil {
			aw.logError("failed to encode ANNOUNCE end message", err)
		}
	}
}

// SendAnnouncement sends an announcement if it's active and not already sent.
// It replaces any existing announcement for the same suffix.
func (aw *AnnouncementWriter) SendAnnouncement(announcement *Announcement) error {
	// Wait for initialization to complete
	select {
	case <-aw.initDone:
		// Initialization complete
	case <-aw.ctx.Done():
		return Cause(aw.ctx)
	}

	aw.mu.RLock()
	initErr := aw.initErr
	aw.mu.RUnlock()
	if initErr != nil {
		return initErr
	}

	if !announcement.IsActive() {
		return nil // No need to send inactive announcements
	}

	if aw.shouldExclude(announcement) {
		return nil
	}

	if aw.holdBack(announcement) {
		return nil
	}

	// Get suffix for this announcement
	suffix, ok := announcement.BroadcastPath().GetSuffix(aw.prefix)
	if !ok {
		return errors.New("moq: broadcast path with invalid prefix")
	}

	aw.mu.Lock()

	active, exists := aw.actives[suffix]
	if exists && active.announcement == announcement {
		aw.mu.Unlock()
		return nil // Already active, no need to re-announce
	}

	// If there's an existing announcement for this suffix, end it first.
	// Release lock before calling end() since it acquires the lock internally.
	var endFunc func()
	if exists && active.end != nil {
		endFunc = active.end
	}
	aw.mu.Unlock()

	if endFunc != nil {
		endFunc()
	}

	aw.mu.Lock()
	defer aw.mu.Unlock()

	// The writer may have been Closed (or CloseWithError'd) while the lock was
	// released above to call the previous announcement's end handler -- Close
	// nils aw.actives. Re-check before encoding: assigning into a nil map panics,
	// and we must not encode onto a closing stream.
	if aw.actives == nil {
		return nil
	}

	// Encode and send ACTIVE announcement
	err := message.AnnounceBroadcastMessage{
		AnnounceStatus:      message.ACTIVE,
		BroadcastPathSuffix: suffix,
		HopIDs:              aw.buildHopIDs(announcement),
	}.Encode(aw.stream)
	if err != nil {
		if strErr, ok := errors.AsType[*transport.StreamError](err); ok {
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

// Close gracefully closes the AnnouncementWriter and ends all active announcements.
func (aw *AnnouncementWriter) Close() error {
	aw.mu.Lock()
	// Collect end functions to call after releasing the lock
	var endFuncs []func()
	for _, active := range aw.actives {
		if active.end != nil {
			endFuncs = append(endFuncs, active.end)
		}
	}
	aw.actives = nil
	aw.dropHidden()
	aw.mu.Unlock()

	// Call end functions without holding the lock to avoid deadlock
	for _, endFunc := range endFuncs {
		endFunc()
	}

	return aw.stream.Close()
}

// CloseWithError ends all active announcements and signals an error condition via the given code.
func (aw *AnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	aw.mu.Lock()
	// Collect end functions to call after releasing the lock
	var endFuncs []func()
	for _, active := range aw.actives {
		if active.end != nil {
			endFuncs = append(endFuncs, active.end)
		}
	}
	aw.actives = nil
	aw.dropHidden()
	aw.mu.Unlock()

	// Call end functions without holding the lock to avoid deadlock
	for _, endFunc := range endFuncs {
		endFunc()
	}

	cancelStreamWithError(aw.stream, transport.StreamErrorCode(code))

	return nil
}

// Context returns the AnnouncementWriter's context.
func (aw *AnnouncementWriter) Context() context.Context {
	if aw == nil || aw.ctx == nil {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}
	return aw.ctx
}

// activeAnnouncement represents an active announcement being managed by AnnouncementWriter.
type activeAnnouncement struct {
	announcement *Announcement
	end          func() // Function to clean up the activeAnnouncement
}

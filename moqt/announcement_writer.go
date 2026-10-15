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

// newAnnouncementWriter serves the peer's announcement request for prefix.
// scope, when set, is what the peer may list: the writer closes with
// UNAUTHORIZED once prefix lies within no root of it.
func newAnnouncementWriter(stream transport.Stream, prefix prefix, localHopID uint64, excludeHop uint64, logger *slog.Logger, scope *authScope) *AnnouncementWriter {
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
		scope:      scope,
	}
	if scope != nil {
		context.AfterFunc(sas.ctx, scope.watch(sas))
	}

	return sas
}

// rescope closes the writer with UNAUTHORIZED once its prefix lies within no
// root of the scope.
func (aw *AnnouncementWriter) rescope() {
	aw.mu.RLock()
	closed := aw.closed
	aw.mu.RUnlock()
	// A writer already closed ended its stream; a reset would undo a clean
	// end.
	if closed || aw.ctx.Err() != nil {
		return
	}
	if !aw.scope.canList(aw.prefix) {
		_ = aw.CloseWithError(AnnounceErrorCodeUnauthorized) // not actionable: it only resets the stream
	}
}

// errAnnouncementWriterClosed is the init error of a writer closed before its
// init could take the active announcements.
var errAnnouncementWriterClosed = errors.New("moqt: announcement writer closed")

// AnnouncementWriter manages the sending of announcements for a specified prefix.
// It handles initialization, sending active announcements, and cleanup.
type AnnouncementWriter struct {
	prefix     prefix
	stream     transport.Stream
	ctx        context.Context
	logger     *slog.Logger
	localHopID uint64 // this node's Hop ID to append when forwarding
	excludeHop uint64 // skip announcements whose HopIDs contain this value
	scope      *authScope

	mu      sync.RWMutex
	actives map[suffix]*activeAnnouncement
	// closed is set once Close or CloseWithError has run; the writer then
	// takes no more announcements, even from an init still running.
	closed bool

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
			// Always replace with the latest active announcement for the suffix.
			actives[sfx] = &activeAnnouncement{announcement: ann}
		}

		aw.mu.Lock()
		if aw.closed {
			err = errAnnouncementWriterClosed
			aw.initErr = err
			aw.mu.Unlock()
			return
		}
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
			aw.registerEndHandlerLocked(sfx, active.announcement)
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

// registerEndHandlerLocked watches the active announcement for sfx, ending it on
// the peer when it ends. Caller MUST hold aw.mu.
func (aw *AnnouncementWriter) registerEndHandlerLocked(sfx suffix, ann *Announcement) {
	active, ok := aw.actives[sfx]
	if !ok || active.announcement != ann {
		// The writer was closed while the caller had released aw.mu.
		return
	}
	active.stop = ann.AfterFunc(func() { aw.remove(sfx, active) })
}

// end stops watching active and sends ENDED for it, unless its own end
// handler already has or the writer has been closed since. It takes aw.mu.
func (aw *AnnouncementWriter) end(sfx suffix, active *activeAnnouncement) {
	if active.stop == nil || !active.stop() {
		return
	}
	aw.remove(sfx, active)
}

// remove forgets active and sends ENDED for it, unless the writer has been
// closed or sfx announced again since. It takes aw.mu.
func (aw *AnnouncementWriter) remove(sfx suffix, active *activeAnnouncement) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	if aw.actives[sfx] == active {
		delete(aw.actives, sfx)
		aw.sendEndedLocked(sfx)
	}
}

// sendEndedLocked sends ANNOUNCE_BROADCAST ENDED for sfx. A stream that has already
// ended, as it has once the session closes, is not written to: its end ends
// every announcement on the peer. Caller MUST hold aw.mu.
func (aw *AnnouncementWriter) sendEndedLocked(sfx suffix) {
	if aw.ctx.Err() != nil {
		return
	}
	err := message.AnnounceBroadcastMessage{
		AnnounceStatus:      message.ENDED,
		BroadcastPathSuffix: sfx,
	}.Encode(aw.stream)
	if err != nil {
		aw.logError("failed to encode ANNOUNCE end message", err)
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
	aw.mu.Unlock()

	if exists {
		aw.end(suffix, active)
	}

	aw.mu.Lock()
	defer aw.mu.Unlock()

	// The writer may have been closed while the lock was released above to
	// end the previous announcement; it must not encode onto a closing stream.
	if aw.closed {
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
	aw.registerEndHandlerLocked(suffix, announcement)

	return nil
}

// Close finishes the stream. The peer then treats every announcement on it as
// ended (moq-lite-05 §5.1.1), so no ENDED is sent.
func (aw *AnnouncementWriter) Close() error {
	aw.stopAll()
	return aw.stream.Close()
}

// CloseWithError resets the stream with the given code, which ends every
// active announcement on the peer as well.
func (aw *AnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	aw.stopAll()
	cancelStreamWithError(aw.stream, transport.StreamErrorCode(code))
	return nil
}

// stopAll empties the writer and stops watching the announcements that were
// active, sending nothing.
func (aw *AnnouncementWriter) stopAll() {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	for _, active := range aw.actives {
		if active.stop != nil {
			active.stop()
		}
	}
	aw.actives = nil
	aw.closed = true
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
	stop         func() bool // stops the watch registerEndHandlerLocked set, if it has not run
}

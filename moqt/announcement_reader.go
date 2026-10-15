package moqt

import (
	"context"
	"iter"
	"sync"

	"github.com/qumo-dev/gomoqt/moqt/internal/message"
	"github.com/qumo-dev/gomoqt/transport"
)

// newAnnouncementReader reads the peer's announcements under prefix. scope,
// when set, is what the peer may publish: an announcement outside it is
// dropped.
func newAnnouncementReader(stream transport.Stream, prefix prefix, initSuffixes []suffix, scope *authScope) *AnnouncementReader {
	if !isValidPrefix(prefix) {
		panic("invalid prefix for AnnouncementReader")
	}

	readerCtx, cancelCause := context.WithCancelCause(stream.Context())

	ar := &AnnouncementReader{
		ctx:         context.WithValue(readerCtx, biStreamTypeCtxKey, message.StreamTypeAnnounce),
		stream:      stream,
		prefix:      prefix,
		actives:     make(map[suffix]*Announcement),
		pendings:    make([]*Announcement, 0),
		announcedCh: make(chan struct{}, 1),
		cancelCause: cancelCause,
		scope:       scope,
	}

	if scope != nil {
		// Watching from before the first read, no scope change misses it.
		context.AfterFunc(ar.ctx, scope.watch(ar))
	}

	for _, suffix := range initSuffixes {
		if scope != nil && !scope.canPublish(BroadcastPath(prefix+suffix)) {
			continue
		}
		ann, _ := NewAnnouncement(readerCtx, BroadcastPath(prefix+suffix))
		ar.actives[suffix] = ann
		ar.pendings = append(ar.pendings, ann)
	}

	// Receive announcements in a separate goroutine
	go func() {
		// The publisher sends ANNOUNCE_OK exactly once before any
		// ANNOUNCE_BROADCAST. Its Hop ID is the implicit trailing entry of
		// every broadcast's Hop ID list on this stream.
		var okMsg message.AnnounceOkMessage
		if err := okMsg.Decode(ar.stream); err != nil {
			// Cancel the reader rather than returning silently: without this
			// a malformed ANNOUNCE_OK (or an unexpected reset) on an
			// otherwise-open stream leaves ReceiveAnnouncement blocked
			// forever, since ar.ctx would never fire.
			ar.fail(err)
			return
		}
		peerHopID := okMsg.HopID

		var am message.AnnounceBroadcastMessage
		var err error

		for {
			err = am.Decode(ar.stream)
			if err != nil {
				// The peer ended or reset the stream, or sent a malformed
				// message. Cancel the reader, as for ANNOUNCE_OK above:
				// that releases a blocked ReceiveAnnouncement and ends every
				// announcement still active, which are children of ar.ctx.
				ar.fail(err)
				return
			}

			// Use an inner scope to ensure Lock/Unlock are paired even across continue/return
			switch am.AnnounceStatus {
			case message.ACTIVE:
				{
					suffix := am.BroadcastPathSuffix
					// Mutate maps under lock. An active for an already-active
					// path atomically replaces the prior advertisement.
					func() {
						ar.announcementsMu.Lock()
						defer func() {
							ar.announcementsMu.Unlock()
						}()
						old, ok := ar.actives[suffix]
						if ok && old.IsActive() {
							old.end()
						}
						// Reconstruct the full hop path: the broadcast's Hop ID
						// list plus the peer's implicit ANNOUNCE_OK Hop ID.
						hopIDs := appendHopID(am.HopIDs, peerHopID)
						// A broadcast the peer may not publish is dropped, not
						// kept for a later scope that might allow it.
						if ar.scope != nil && !ar.scope.canPublish(BroadcastPath(ar.prefix+suffix)) {
							delete(ar.actives, suffix)
							return
						}
						ar.activateLocked(suffix, hopIDs)
					}()
				}
			case message.ENDED:
				{
					suffix := am.BroadcastPathSuffix
					var handled bool
					func() {
						ar.announcementsMu.Lock()
						defer func() {
							ar.announcementsMu.Unlock()
						}()
						old, ok := ar.actives[suffix]
						if ok && old.IsActive() {
							old.end()
							delete(ar.actives, suffix)
							handled = true
						}
					}()
					// A scoped reader dropped or ended the broadcasts outside
					// its scope, so their ENDED finds nothing.
					if ar.scope != nil {
						handled = true
					}
					if handled {
						continue
					}
					// An ended for a path that is not currently active is a
					// protocol violation; reset the stream.
					ar.CloseWithError(AnnounceErrorCodeDuplicated)
					return
				}
			default:
				ar.CloseWithError(AnnounceErrorCodeInvalidStatus)
				return
			}
		}
	}()

	return ar
}

// appendHopID appends the peer's ANNOUNCE_OK Hop ID to a broadcast's Hop ID
// list, reconstructing the full hop path. A peer Hop ID of 0 means the peer
// does not participate in hop tracking and is not appended.
func appendHopID(hopIDs []uint64, peerHopID uint64) []uint64 {
	if peerHopID == 0 {
		return hopIDs
	}
	result := make([]uint64, len(hopIDs)+1)
	copy(result, hopIDs)
	result[len(hopIDs)] = peerHopID
	return result
}

// AnnouncementReader receives and manages broadcast announcements for a prefix from
// a remote peer. It maintains active announcements and notifies the caller when new
// announcements arrive or existing ones end.
type AnnouncementReader struct {
	stream transport.Stream
	prefix prefix

	ctx context.Context

	// Track Suffix -> Announcement
	announcementsMu sync.Mutex

	actives map[suffix]*Announcement

	pendings    []*Announcement
	announcedCh chan struct{} // notify when new announcement is available

	// scope, when set, is what the peer may publish; a broadcast outside it is
	// dropped.
	scope *authScope

	// cancelCause cancels readerCtx (and thus ar.ctx) when the announce
	// goroutine hits a fatal decode error, so ReceiveAnnouncement unblocks.
	cancelCause context.CancelCauseFunc
}

// activateLocked makes a new active announcement for suffix and queues it for
// ReceiveAnnouncement. Caller MUST hold announcementsMu.
func (ar *AnnouncementReader) activateLocked(sfx suffix, hopIDs []uint64) {
	ann, _ := NewAnnouncement(ar.ctx, BroadcastPath(ar.prefix+sfx))
	ann.hopIDs = hopIDs
	ar.actives[sfx] = ann
	ar.pendings = append(ar.pendings, ann)
	select {
	case ar.announcedCh <- struct{}{}:
	default:
	}
}

// rescope re-applies the scope after it changed: it ends each active announcement
// the peer may no longer publish. One the peer announced while it was refused
// stays dropped until the peer announces it again.
func (ar *AnnouncementReader) rescope() {
	if ar.scope == nil {
		return
	}

	ar.announcementsMu.Lock()
	defer ar.announcementsMu.Unlock()
	if ar.ctx.Err() != nil {
		return
	}

	for sfx, ann := range ar.actives {
		if ar.scope.canPublish(ann.BroadcastPath()) {
			continue
		}
		ann.end()
		delete(ar.actives, sfx)
	}
}

// fail cancels the reader context with cause, unblocking ReceiveAnnouncement.
func (ar *AnnouncementReader) fail(err error) {
	if ar.cancelCause != nil {
		ar.cancelCause(err)
	}
}

// ReceiveAnnouncement blocks until an announcement for the configured prefix is available or until ctx or the reader's context is canceled.
// It returns the next pending Announcement or an error describing the cancellation cause.
// Each broadcast path is returned when it becomes active; its ending is not
// returned again but observed on the Announcement (Done, IsActive, AfterFunc).
func (ras *AnnouncementReader) ReceiveAnnouncement(ctx context.Context) (*Announcement, error) {
	for {
		ras.announcementsMu.Lock()

		if len(ras.pendings) > 0 {
			next := ras.pendings[0]
			ras.pendings = ras.pendings[1:]

			ras.announcementsMu.Unlock()

			return next, nil
		}

		if ras.ctx.Err() != nil {
			ras.announcementsMu.Unlock()
			return nil, Cause(ras.ctx)
		}

		announceCh := ras.announcedCh

		ras.announcementsMu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ras.ctx.Done():
			return nil, Cause(ras.ctx)
		case <-announceCh:
			// New announcement available, loop to check pendings
			continue
		}
	}
}

// Announcements returns an iterator that yields Announcement values for the configured prefix until ctx or the reader's context is canceled.
// It yields each broadcast path when it becomes active, as ReceiveAnnouncement
// returns it; an ending is observed on the Announcement itself.
func (ras *AnnouncementReader) Announcements(ctx context.Context) iter.Seq[*Announcement] {
	return func(yield func(*Announcement) bool) {
		for {
			ann, err := ras.ReceiveAnnouncement(ctx)
			if err != nil {
				return
			}

			if !yield(ann) {
				return
			}
		}
	}

}

// Close closes the AnnouncementReader and releases resources. It is safe to call multiple times.
func (ras *AnnouncementReader) Close() error {
	ras.announcementsMu.Lock()
	defer ras.announcementsMu.Unlock()

	if ras.ctx.Err() != nil {
		return nil
	}

	// Cancel the reader context so any blocked ReceiveAnnouncement unblocks
	// promptly (rather than busy-spinning on a closed announcedCh).
	ras.fail(context.Canceled)

	if ras.announcedCh != nil {
		close(ras.announcedCh)
		ras.announcedCh = nil
	}

	return ras.stream.Close()
}

// CloseWithError closes the AnnouncementReader with an error, ending all
// active announcements and canceling the stream with the specified error
// code.
func (ras *AnnouncementReader) CloseWithError(code AnnounceErrorCode) {
	ras.announcementsMu.Lock()
	defer ras.announcementsMu.Unlock()

	ras.fail(context.Canceled)

	if ras.announcedCh != nil {
		close(ras.announcedCh)
		ras.announcedCh = nil
	}

	cancelStreamWithError(ras.stream, transport.StreamErrorCode(code))
}

// Context returns the AnnouncementReader's context. It is canceled when the reader is closed.
func (ras *AnnouncementReader) Context() context.Context {
	if ras == nil || ras.ctx == nil {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}
	return ras.ctx
}

// Alias types for better readability
type suffix = string
type prefix = string

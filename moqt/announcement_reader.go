package moqt

import (
	"context"
	"iter"
	"sync"

	"github.com/qumo-dev/gomoqt/moqt/internal/message"
	"github.com/qumo-dev/gomoqt/transport"
)

func newAnnouncementReader(stream transport.Stream, prefix prefix, initSuffixes []suffix) *AnnouncementReader {
	return newScopedAnnouncementReader(stream, prefix, initSuffixes, nil)
}

// newScopedAnnouncementReader is newAnnouncementReader with allow, which, when
// set, reports whether the peer may publish a broadcast path. An announcement
// it refuses is held back until rescope allows it.
func newScopedAnnouncementReader(stream transport.Stream, prefix prefix, initSuffixes []suffix, allow func(BroadcastPath) bool) *AnnouncementReader {
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
		allow:       allow,
	}

	for _, suffix := range initSuffixes {
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
						if ar.allow != nil && !ar.allow(BroadcastPath(ar.prefix+suffix)) {
							delete(ar.actives, suffix)
							if ar.hidden == nil {
								ar.hidden = make(map[string][]uint64)
							}
							ar.hidden[suffix] = hopIDs
							return
						}
						delete(ar.hidden, suffix)
						ar.activate(suffix, hopIDs)
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
						if _, held := ar.hidden[suffix]; held {
							delete(ar.hidden, suffix)
							handled = true
						}
					}()
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

	// allow, when set, reports whether the peer may publish a broadcast
	// path; hidden holds the Hop IDs of each active broadcast it refused.
	allow  func(BroadcastPath) bool
	hidden map[suffix][]uint64

	// cancelCause cancels readerCtx (and thus ar.ctx) when the announce
	// goroutine hits a fatal decode error, so ReceiveAnnouncement unblocks.
	cancelCause context.CancelCauseFunc
}

// activate makes a new active announcement for suffix and queues it for
// ReceiveAnnouncement. Caller MUST hold announcementsMu.
func (ar *AnnouncementReader) activate(sfx suffix, hopIDs []uint64) {
	ann, _ := NewAnnouncement(ar.ctx, BroadcastPath(ar.prefix+sfx))
	ann.hopIDs = hopIDs
	ar.actives[sfx] = ann
	ar.pendings = append(ar.pendings, ann)
	select {
	case ar.announcedCh <- struct{}{}:
	default:
	}
}

// rescope re-applies allow after it changed: it ends each active announcement
// the peer may no longer publish, holding it back, and activates each
// held-back one it now may.
func (ar *AnnouncementReader) rescope() {
	if ar.allow == nil {
		return
	}

	ar.announcementsMu.Lock()
	defer ar.announcementsMu.Unlock()
	if ar.ctx.Err() != nil {
		return
	}

	for sfx, ann := range ar.actives {
		if ar.allow(ann.BroadcastPath()) {
			continue
		}
		ann.end()
		delete(ar.actives, sfx)
		if ar.hidden == nil {
			ar.hidden = make(map[suffix][]uint64)
		}
		ar.hidden[sfx] = ann.HopIDs()
	}
	for sfx, hopIDs := range ar.hidden {
		if ar.allow(BroadcastPath(ar.prefix + sfx)) {
			delete(ar.hidden, sfx)
			ar.activate(sfx, hopIDs)
		}
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

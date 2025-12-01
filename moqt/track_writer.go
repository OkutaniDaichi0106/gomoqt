package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newTrackWriter(path BroadcastPath, name TrackName,
	subscribeStream *receiveSubscribeStream,
	openUniStreamFunc func() (quic.SendStream, error),
	onCloseTrackFunc func(),
) *TrackWriter {
	track := &TrackWriter{
		BroadcastPath:          path,
		TrackName:              name,
		receiveSubscribeStream: subscribeStream,
		activeGroups:           make(map[*GroupWriter]struct{}),
		openUniStreamFunc:      openUniStreamFunc,
		onCloseTrackFunc:       onCloseTrackFunc,
	}

	return track
}

// TrackWriter writes groups for a published track.
// It manages the lifecycle of active groups for that track.
// The TrackWriter provides methods to open group writers and to inspect the track configuration.
type TrackWriter struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	*receiveSubscribeStream

	groupMapMu   sync.Mutex
	activeGroups map[*GroupWriter]struct{}

	// closeMu controls exclusivity between Close/CloseWithError and OpenGroup.
	// OpenGroup acquires a read lock so multiple OpenGroup calls may run
	// concurrently, while Close/CloseWithError acquires a write lock to
	// make the close exclusive with ongoing and future OpenGroup operations
	// until close completes.
	closeMu sync.RWMutex

	openUniStreamFunc func() (quic.SendStream, error)

	onCloseTrackFunc func()
}

// Close stops publishing and cancels active groups.
func (s *TrackWriter) Close() error {
	// Take the write lock to ensure Close is exclusive with OpenGroup calls.
	// This prevents OpenGroup from running concurrently with Close and
	// provides a deterministic semantics: either the OpenGroup completes
	// entirely before Close proceeds, or Close waits for OpenGroup to finish.
	s.closeMu.Lock()
	defer s.closeMu.Unlock()
	// Take a snapshot of active groups under lock. activeGroups == nil
	// indicates the track has been closed and prevents new groups.
	s.groupMapMu.Lock()
	if s.activeGroups == nil {
		s.groupMapMu.Unlock()
		return nil
	}
	groups := make([]*GroupWriter, 0, len(s.activeGroups))
	for g := range s.activeGroups {
		groups = append(groups, g)
	}
	// Prevent further additions and drop map
	s.activeGroups = nil
	s.groupMapMu.Unlock()

	for _, g := range groups {
		_ = g.Close()
	}

	// Then close the subscribe stream if present
	var err error
	if s.receiveSubscribeStream != nil {
		err = s.receiveSubscribeStream.close()
		s.receiveSubscribeStream = nil
	}

	if s.onCloseTrackFunc != nil {
		s.onCloseTrackFunc()
		s.onCloseTrackFunc = nil
	}

	return err
}

// CloseWithError stops publishing due to an error and cancels active groups.
func (s *TrackWriter) CloseWithError(code SubscribeErrorCode) {
	// Ensure CloseWithError is exclusive with OpenGroup.
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	s.groupMapMu.Lock()
	if s.activeGroups != nil {

		// Cancel all active groups first
		for group := range s.activeGroups {
			group.CancelWrite(PublishAbortedErrorCode)
		}
		s.activeGroups = nil

	}
	s.groupMapMu.Unlock()
	if s.receiveSubscribeStream != nil {
		if err := s.receiveSubscribeStream.closeWithError(code); err != nil {
			slog.Error("failed to close receive subscribe stream with error", "error", err)
		}
		s.receiveSubscribeStream = nil
	}

	if s.onCloseTrackFunc != nil {
		s.onCloseTrackFunc()
		s.onCloseTrackFunc = nil
	}
}

// OpenGroup opens a new group for the provided group sequence and returns a GroupWriter to write frames into it.
// If seq is zero an error is returned.
func (s *TrackWriter) OpenGroup(seq GroupSequence) (*GroupWriter, error) {
	if seq == 0 {
		return nil, errors.New("group sequence must not be zero")
	}

	// Avoid accessing s.ctx directly; it can be nil if the receiveSubscribeStream
	// has been cleared during Close(). Instead, capture the receiveSubscribeStream
	// under lock and validate its context below.

	// Prevent opening a new group if the track has been closed. Capture
	// receiveSubscribeStream under lock so it cannot be set to nil by Close()
	// Acquire a shared lock so multiple OpenGroup calls can proceed
	// concurrently while ensuring Close waits for them to finish.
	s.closeMu.RLock()
	defer s.closeMu.RUnlock()

	// Check the context on the captured receiveSubscribeStream instead of s.ctx
	// to avoid nil deref if the embedded field has been cleared by Close().
	if s.Context().Err() != nil {
		return nil, Cause(s.Context())
	}

	// Write the INFO message to the receive subscribe stream.
	err := s.WriteInfo(Info{})
	if err != nil {
		return nil, err
	}

	stream, err := s.openUniStreamFunc()
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			sessErr := &SessionError{
				ApplicationError: appErr,
			}
			return nil, sessErr
		}
		return nil, err
	}

	err = message.StreamTypeGroup.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			return nil, &GroupError{StreamError: strErr}
		}

		strErrCode := quic.StreamErrorCode(InternalGroupErrorCode)
		stream.CancelWrite(strErrCode)

		return nil, err
	}

	err = message.GroupMessage{
		SubscribeID:   uint64(s.subscribeID),
		GroupSequence: uint64(seq),
	}.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			return nil, &GroupError{StreamError: strErr}
		}

		strErrCode := quic.StreamErrorCode(InternalGroupErrorCode)
		stream.CancelWrite(strErrCode)

		return nil, err
	}

	var group *GroupWriter
	group = newGroupWriter(stream, seq, func() { s.removeGroup(group) })
	s.addGroup(group)

	return group, nil
}

func (s *TrackWriter) addGroup(group *GroupWriter) {
	s.groupMapMu.Lock()
	defer s.groupMapMu.Unlock()

	s.activeGroups[group] = struct{}{}
}

func (s *TrackWriter) removeGroup(group *GroupWriter) {
	s.groupMapMu.Lock()
	defer s.groupMapMu.Unlock()

	delete(s.activeGroups, group)
}

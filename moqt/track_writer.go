package moqt

import (
	"context"
	"errors"
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

	openUniStreamFunc func() (quic.SendStream, error)

	onCloseTrackFunc func()
}

// Close stops publishing and cancels active groups.
func (s *TrackWriter) Close() error {
	s.groupMapMu.Lock()

	// Cancel all active groups first
	for group := range s.activeGroups {
		group.CancelWrite(PublishAbortedErrorCode)
	}
	s.activeGroups = nil

	s.groupMapMu.Unlock()

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
	s.groupMapMu.Lock()

	// Cancel all active groups first
	for group := range s.activeGroups {
		group.CancelWrite(PublishAbortedErrorCode)
	}
	s.activeGroups = nil

	s.groupMapMu.Unlock()

	if s.receiveSubscribeStream != nil {
		s.receiveSubscribeStream.closeWithError(code)
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

	if s.ctx.Err() != nil {
		return nil, Cause(s.ctx)
	}

	err := s.receiveSubscribeStream.WriteInfo(Info{})
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

		return nil, GroupError{
			StreamError: &quic.StreamError{
				StreamID:  stream.StreamID(),
				ErrorCode: strErrCode,
			},
		}
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

// TrackConfig returns the TrackConfig supplied by the subscriber, if any.
func (s *TrackWriter) TrackConfig() *TrackConfig {
	if s.receiveSubscribeStream == nil {
		return nil
	}
	return s.receiveSubscribeStream.TrackConfig()
}

// Context returns the context associated with the TrackWriter.
func (s *TrackWriter) Context() context.Context {
	return s.ctx
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

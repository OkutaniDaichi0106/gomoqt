package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
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
		activeGroups:           make(map[quic.StreamID]*sendGroupStream),
		openUniStreamFunc:      openUniStreamFunc,
		onCloseTrackFunc:       onCloseTrackFunc,
	}

	return track
}

type TrackWriter struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	*receiveSubscribeStream

	accepted bool

	mu           sync.Mutex
	activeGroups map[quic.StreamID]*sendGroupStream

	openUniStreamFunc func() (quic.SendStream, error)

	onCloseTrackFunc func()
}

func (s *TrackWriter) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, group := range s.activeGroups {
		group.CancelWrite(PublishAbortedErrorCode)
	}
	s.activeGroups = nil

	s.receiveSubscribeStream.close()

	if s.onCloseTrackFunc != nil {
		s.onCloseTrackFunc()
	}
}

func (s *TrackWriter) CloseWithError(code SubscribeErrorCode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, group := range s.activeGroups {
		group.CancelWrite(SubscribeCanceledErrorCode)
	}
	s.activeGroups = nil

	s.receiveSubscribeStream.closeWithError(code)
	if s.onCloseTrackFunc != nil {
		s.onCloseTrackFunc()
	}
}

func (s *TrackWriter) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if seq == 0 {
		return nil, errors.New("group sequence must not be zero")
	}

	if err := s.ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()

	if !s.accepted {
		err := s.receiveSubscribeStream.writeInfo(Info{})
		if err != nil {
			s.mu.Unlock()
			return nil, err
		}

		s.accepted = true
	}

	stream, err := s.openUniStreamFunc()
	if err != nil {
		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			sessErr := &SessionError{
				ApplicationError: appErr,
			}
			s.mu.Unlock()
			return nil, sessErr
		}
		s.mu.Unlock()
		return nil, err
	}

	s.mu.Unlock()

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
		SubscribeID:   s.subscribeID,
		GroupSequence: seq,
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

	group := newSendGroupStream(s.ctx, stream, seq, func() {
		s.removeGroup(stream.StreamID())
	})
	s.addGroup(group)

	slog.Debug("track writer opened group")

	return group, nil
}

func (s *TrackWriter) WriteInfo(info Info) error {
	if !s.accepted {
		err := s.receiveSubscribeStream.writeInfo(info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TrackWriter) TrackConfig() *TrackConfig {
	if s.receiveSubscribeStream == nil {
		return nil
	}
	return s.receiveSubscribeStream.TrackConfig()
}

func (s *TrackWriter) addGroup(group *sendGroupStream) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeGroups[group.stream.StreamID()] = group
}

func (s *TrackWriter) removeGroup(id quic.StreamID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeGroups, id)
}

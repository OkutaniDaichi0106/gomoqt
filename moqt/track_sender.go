package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newTrackSender(ctx context.Context, trackLogger *slog.Logger,
	openUniStreamFunc func() (quic.SendStream, error),
	acceptFunc func(Info) error,
	onCloseTrack func(),
) *trackSender {
	track := &trackSender{
		ctx:               ctx,
		groupsMap:         make(map[GroupSequence]func()),
		openUniStreamFunc: openUniStreamFunc,
		logger:            trackLogger,
		acceptFunc:        acceptFunc,
		onCloseTrack:      onCloseTrack,
	}

	// go func() {
	// 	<-ctx.Done()
	// 	track.mu.Lock()
	// 	defer track.mu.Unlock()

	// 	if ctx.Err() != nil {
	// 		for stream := range track.queue {
	// 			stream.CancelWrite(SubscribeCanceledErrorCode)
	// 		}
	// 	} else {
	// 		for stream := range track.queue {
	// 			stream.Close()
	// 		}
	// 	}

	// 	track.queue = nil
	// }()
	return track
}

var _ TrackWriter = (*trackSender)(nil)

type trackSender struct {
	ctx context.Context

	logger *slog.Logger

	acceptFunc func(Info) error

	mu        sync.Mutex
	groupsMap map[GroupSequence]func()

	openUniStreamFunc func() (quic.SendStream, error)

	onCloseTrack func()
}

func (s *trackSender) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, closeGruopFunc := range s.groupsMap {
		closeGruopFunc()
	}

	if s.onCloseTrack != nil {
		s.onCloseTrack()
	}
	s.groupsMap = nil
}

func (s *trackSender) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if seq == 0 {
		return nil, errors.New("group sequence must not be zero")
	}

	if err := s.ctx.Err(); err != nil {
		return nil, err
	}

	// s.mu.Lock()
	// defer s.mu.Unlock()

	err := s.acceptFunc(Info{})
	if err != nil {
		return nil, err
	}

	stream, err := s.openUniStreamFunc()
	if err != nil {
		s.logger.Error("failed to open an unidirectional stream",
			"error", err,
		)

		var appErr *quic.ApplicationError
		if errors.As(err, &appErr) {
			sessErr := &SessionError{
				ApplicationError: appErr,
			}
			return nil, sessErr
		}

		return nil, err
	}

	err = message.StreamTypeMessage{
		StreamType: stream_type_group,
	}.Encode(stream)
	if err != nil {
		s.logger.Error("failed to send stream type message",
			"error", err,
		)

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
		GroupSequence: seq,
	}.Encode(stream)
	if err != nil {
		s.logger.Error("failed to send group message",
			"error", err,
		)

		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			s.logger.Error("group message encoding failed",
				"error", strErr,
			)
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

	group := newSendGroupStream(s.ctx, stream, seq, func() {
		s.removeCloseGroupFunc(seq)
	})

	s.addCloseGroupFunc(group.sequence, func() {
		group.CancelWrite(PublishAbortedErrorCode)
	})

	// s.queue[group] = struct{}{}
	// go func() {
	// 	<-group.ctx.Done()
	// 	s.mu.Lock()
	// 	delete(s.queue, group)
	// 	s.mu.Unlock()
	// }()

	return group, nil
}

func (s *trackSender) addCloseGroupFunc(seq GroupSequence, closeFunc func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groupsMap[seq] = closeFunc
}

func (s *trackSender) removeCloseGroupFunc(seq GroupSequence) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.groupsMap, seq)
}

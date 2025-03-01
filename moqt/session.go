package moqt

import (
	"context"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type Session interface {
	/*
	 * Methods for the Client
	 */
	// Update the session
	UpdateSession(bitrate uint64) error // TODO:

	// Terminate the session
	Terminate(error)

	/*
	 * Methods for the Subscriber
	 */
	// Open an Announce Stream
	OpenAnnounceStream(AnnounceConfig) (AnnouncementReader, error)

	// Open a Track Stream
	OpenTrackStream(SubscribeConfig) (Info, ReceiveTrackStream, error)

	// Request Track Info
	RequestTrackInfo(InfoRequest) (Info, error)

	/*
	 * Methods for the Publisher
	 */
	// Accept an Announce Stream
	AcceptAnnounceStream(context.Context, func(AnnounceConfig) error) (AnnouncementWriter, error)

	// Accept a Track Stream
	AcceptTrackStream(context.Context, func(SubscribeConfig) (Info, error)) (SendTrackStream, error)

	// Accept an Info Stream
	RespondTrackInfo(context.Context, func(InfoRequest) (Info, error)) error
}

var _ Session = (*session)(nil)

type session struct {
	internalSession    *internal.Session
	subscribeIDCounter uint64

	extensions Parameters
}

func (s *session) UpdateSession(bitrate uint64) error {
	return s.internalSession.UpdateSession(bitrate)
}

func (s *session) Terminate(err error) {
	s.internalSession.Terminate(err)
}

func (s *session) OpenAnnounceStream(config AnnounceConfig) (AnnouncementReader, error) {
	apm := message.AnnouncePleaseMessage{
		TrackPrefix: config.TrackPrefix,
	}

	ras, err := s.internalSession.OpenAnnounceStream(apm)
	if err != nil {
		return nil, err
	}

	return newReceiveAnnounceStream(ras), nil
}

func (s *session) OpenTrackStream(config SubscribeConfig) (Info, ReceiveTrackStream, error) {
	sm := message.SubscribeMessage{
		SubscribeID:      s.nextSubscribeID(),
		TrackPath:        string(config.TrackPath),
		GroupOrder:       message.GroupOrder(config.GroupOrder),
		TrackPriority:    message.TrackPriority(config.TrackPriority),
		MinGroupSequence: message.GroupSequence(config.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(config.MaxGroupSequence),
	}

	im, ss, err := s.internalSession.OpenSubscribeStream(sm)
	if err != nil {
		return Info{}, nil, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}

	return info, newReceiveTrackStream(s.internalSession, info, ss), nil
}

func (s *session) RequestTrackInfo(irm InfoRequest) (Info, error) {
	im, err := s.internalSession.OpenInfoStream(message.InfoRequestMessage{
		TrackPath: string(irm.TrackPath),
	})
	if err != nil {
		return Info{}, err
	}

	return Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}, nil
}

func (s *session) AcceptAnnounceStream(ctx context.Context, handler func(AnnounceConfig) error) (AnnouncementWriter, error) {
	as, err := s.internalSession.AcceptAnnounceStream(ctx)
	if err != nil {
		return nil, err
	}

	sas := &sendAnnounceStream{internalStream: as}

	if handler != nil {
		err = handler(sas.AnnounceConfig())
		if err != nil {
			sas.CloseWithError(err)
			return nil, err
		}
	}

	return sas, nil
}

func (s *session) AcceptTrackStream(ctx context.Context, handler func(SubscribeConfig) (Info, error)) (SendTrackStream, error) {
	ss, err := s.internalSession.AcceptSubscribeStream(ctx)
	if err != nil {
		return nil, err

	}

	if ss == nil {
		return nil, ErrInternalError
	}

	sts := &sendTrackStream{subscribeStream: ss}

	info, err := handler(sts.SubscribeConfig())
	if err != nil {
		sts.CloseWithError(err)
		return nil, err
	}

	im := message.InfoMessage{
		TrackPriority:       message.TrackPriority(info.TrackPriority),
		LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(info.GroupOrder),
	}

	_, err = im.Encode(sts.subscribeStream.Stream)
	if err != nil {
		return nil, err
	}

	return sts, nil
}

func (s *session) RespondTrackInfo(ctx context.Context, handler func(InfoRequest) (Info, error)) error {
	irs, err := s.internalSession.AcceptInfoStream(ctx)
	if err != nil {
		return err

	}

	if irs == nil {
		return ErrInternalError
	}

	irm := InfoRequest{
		TrackPath: TrackPath(irs.InfoRequestMessage.TrackPath),
	}

	info, err := handler(irm)
	if err != nil {
		return err
	}

	im := message.InfoMessage{
		TrackPriority:       message.TrackPriority(info.TrackPriority),
		LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(info.GroupOrder),
	}

	_, err = im.Encode(irs.Stream)
	if err != nil {
		return err
	}

	return nil
}

func (s *session) nextSubscribeID() message.SubscribeID {
	// Increment and return the previous value atomically
	id := atomic.AddUint64(&s.subscribeIDCounter, 1) - 1
	return message.SubscribeID(id)
}

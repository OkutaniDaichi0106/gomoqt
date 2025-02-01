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
	UpdateSession(bitrate uint64) error

	// Terminate the session
	Terminate(error)

	/*
	 * Methods for the Subscriber
	 */
	// Open an Announce Stream
	OpenAnnounceStream(AnnounceConfig) (ReceiveAnnounceStream, error)

	// Open a Subscribe Stream
	OpenSubscribeStream(SubscribeConfig) (SendSubscribeStream, Info, error)

	// Open an Info Stream
	OpenInfoStream(InfoRequest) (Info, error)

	// Open a Fetch Stream
	OpenFetchStream(FetchRequest) (SendFetchStream, error)

	// Accept a Group Stream
	AcceptGroupStream(context.Context, SendSubscribeStream) (ReceiveGroupStream, error)

	/*
	 * Methods for the Publisher
	 */
	// Accept an Announce Stream
	AcceptAnnounceStream(context.Context, func(AnnounceConfig) error) (SendAnnounceStream, error)

	// Accept a Subscribe Stream
	AcceptSubscribeStream(context.Context, func(SubscribeConfig) (Info, error)) (ReceiveSubscribeStream, error)

	// Accept a Fetch Stream
	AcceptFetchStream(context.Context, func(FetchRequest) error) (ReceiveFetchStream, error)

	// Accept an Info Stream
	AcceptInfoStream(context.Context, func(InfoRequest) (Info, error)) error

	// Open a Group Stream
	OpenGroupStream(ReceiveSubscribeStream, GroupSequence) (SendGroupStream, error)
}

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

func (s *session) OpenAnnounceStream(config AnnounceConfig) (ReceiveAnnounceStream, error) {
	apm := message.AnnouncePleaseMessage{
		AnnounceParameters: config.Parameters.paramMap,
		TrackPrefix:        config.TrackPrefix,
	}

	ras, err := s.internalSession.OpenAnnounceStream(apm)
	if err != nil {
		return nil, err
	}

	return &receiveAnnounceStream{internalStream: ras}, nil
}

func (s *session) OpenSubscribeStream(config SubscribeConfig) (SendSubscribeStream, Info, error) {
	sm := message.SubscribeMessage{
		SubscribeID:         s.nextSubscribeID(),
		SubscribeParameters: config.SubscribeParameters.paramMap,
		TrackPath:           config.TrackPath,
	}

	ss, im, err := s.internalSession.OpenSubscribeStream(sm)
	if err != nil {
		return nil, Info{}, err
	}

	return &sendSubscribeStream{internalStream: ss}, Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}, nil
}

func (s *session) OpenInfoStream(irm InfoRequest) (Info, error) {
	im, err := s.internalSession.OpenInfoStream(message.InfoRequestMessage{
		TrackPath: irm.TrackPath,
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

func (s *session) OpenFetchStream(fm FetchRequest) (SendFetchStream, error) {
	sfs, err := s.internalSession.OpenFetchStream(message.FetchMessage{})
	if err != nil {
		return nil, err
	}

	return &sendFetchStream{internalStream: sfs}, nil
}

func (s *session) AcceptGroupStream(ctx context.Context, substr SendSubscribeStream) (ReceiveGroupStream, error) {
	str, err := s.internalSession.AcceptGroupStream(ctx, message.SubscribeID(substr.SubscribeID()))
	if err != nil {
		return nil, err
	}

	return &receiveGroupStream{internalStream: str}, nil
}

func (s *session) AcceptAnnounceStream(ctx context.Context, handler func(AnnounceConfig) error) (SendAnnounceStream, error) {
	as, err := s.internalSession.AcceptAnnounceStream(ctx)
	if err != nil {
		return nil, err
	}

	sas := &sendAnnounceStream{internalStream: as}

	err = handler(sas.AnnounceConfig())
	if err != nil {
		return nil, err
	}

	return sas, nil
}

func (s *session) AcceptSubscribeStream(ctx context.Context, handler func(SubscribeConfig) (Info, error)) (ReceiveSubscribeStream, error) {
	ss, err := s.internalSession.AcceptSubscribeStream(ctx)
	if err != nil {
		return nil, err
	}

	if ss == nil {
		return nil, ErrInternalError
	}

	sss := &receiveSubscribeStream{internalStream: ss}

	info, err := handler(sss.SubscribeConfig())
	if err != nil {
		return nil, err
	}

	im := message.InfoMessage{
		TrackPriority:       message.TrackPriority(info.TrackPriority),
		LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(info.GroupOrder),
	}

	_, err = im.Encode(sss.internalStream.Stream)
	if err != nil {
		return nil, err
	}

	return sss, nil
}

func (s *session) AcceptFetchStream(ctx context.Context, handler func(FetchRequest) error) (ReceiveFetchStream, error) {
	rf, err := s.internalSession.AcceptFetchStream(ctx)
	if err != nil {
		return nil, err
	}

	rfs := &receiveFetchStream{internalStream: rf}

	err = handler(rfs.FetchRequest())
	if err != nil {
		return nil, err
	}

	return rfs, nil
}

func (s *session) AcceptInfoStream(ctx context.Context, handler func(InfoRequest) (Info, error)) error {
	irs, err := s.internalSession.AcceptInfoStream(ctx)
	if err != nil {
		return err
	}

	if irs == nil {
		return ErrInternalError
	}

	irm := InfoRequest{
		TrackPath: irs.InfoRequestMessage.TrackPath,
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

func (s *session) OpenGroupStream(ssr ReceiveSubscribeStream, gs GroupSequence) (SendGroupStream, error) {
	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(ssr.SubscribeID()),
		GroupSequence: message.GroupSequence(gs),
	}

	sgs, err := s.internalSession.OpenGroupStream(gm)
	if err != nil {
		return nil, err
	}

	return &sendGroupStream{internalStream: sgs}, nil
}

func (s *session) nextSubscribeID() message.SubscribeID {
	new := message.SubscribeID(atomic.LoadUint64(&s.subscribeIDCounter))

	atomic.AddUint64(&s.subscribeIDCounter, 1)

	return new
}

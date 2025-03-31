package moqt

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type Session interface {
	/*
	 * Methods for the Client
	 */
	// Update the session
	// UpdateSession(bitrate uint64) error // TODO:

	// Terminate the session
	Terminate(error)

	/*
	 * Methods for the Subscriber
	 */
	// Open an Announce Stream
	OpenAnnounceStream(*AnnounceConfig) (AnnouncementReader, error)

	// Open a Track Stream
	OpenTrackStream(TrackPath, *SubscribeConfig) (Info, ReceiveTrackStream, error)

	// Request Track Info
	RequestInfo(TrackPath) (Info, error)

	/*
	 * Methods for the Publisher
	 */
	// Accept an Announce Stream
	AcceptAnnounceStream(context.Context) (AnnouncementWriter, error)

	// Accept a Track Stream
	AcceptTrackStream(context.Context) (SendTrackStream, error)

	// Accept an Info Stream
	// RespondInfo(context.Context) error
}

var _ Session = (*session)(nil)

func newSession(internalSession *internal.Session, handler TrackHandler) Session {
	sess := &session{
		internalSession: internalSession,
		handler:         handler,
	}

	go sess.resolveInfoRequest(context.TODO()) // TODO: context.TODO()?

	return sess
}

type session struct {
	internalSession    *internal.Session
	subscribeIDCounter uint64

	handler TrackHandler
}

func (s *session) Terminate(err error) {
	s.internalSession.Terminate(err)
	slog.Debug("session terminated", "error", err)
}

func (s *session) OpenAnnounceStream(config *AnnounceConfig) (AnnouncementReader, error) {
	if config == nil {
		config = &AnnounceConfig{TrackPattern: "/**"}
	}

	apm := message.AnnouncePleaseMessage{
		TrackPattern: config.TrackPattern,
	}

	ras, err := s.internalSession.OpenAnnounceStream(apm)
	if err != nil {
		return nil, err
	}

	slog.Debug("opened an announce stream", "announce_config", config.String())

	return newReceiveAnnounceStream(ras), nil
}

func (s *session) OpenTrackStream(path TrackPath, config *SubscribeConfig) (Info, ReceiveTrackStream, error) {
	if config == nil {
		config = &SubscribeConfig{}
	}

	id := s.nextSubscribeID()

	slog.Debug("opening track stream", "subscribe_config", config.String(), "subscribe_id", id)

	sm := message.SubscribeMessage{
		SubscribeID:      id,
		TrackPath:        string(path),
		GroupOrder:       message.GroupOrder(config.GroupOrder),
		TrackPriority:    message.TrackPriority(config.TrackPriority),
		MinGroupSequence: message.GroupSequence(config.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(config.MaxGroupSequence),
	}

	im, ss, err := s.internalSession.OpenSubscribeStream(sm)
	if err != nil {
		return NotFoundInfo, nil, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}

	return info, newReceiveTrackStream(s.internalSession, info, ss), nil
}

func (s *session) RequestInfo(path TrackPath) (Info, error) {
	slog.Debug("requesting track info", "track_path", path)

	im, err := s.internalSession.OpenInfoStream(message.InfoRequestMessage{
		TrackPath: string(path),
	})
	if err != nil {
		slog.Error("failed to request track info",
			"track_path", path,
			"error", err,
		)
		return NotFoundInfo, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.TrackPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
	}

	slog.Debug("received track info", "info", info.String())

	return info, nil
}

func (s *session) AcceptAnnounceStream(ctx context.Context) (AnnouncementWriter, error) {
	slog.Debug("accepting announce stream")

	as, err := s.internalSession.AcceptAnnounceStream(ctx)
	if err != nil {
		return nil, err
	}

	sas := newSendAnnounceStream(as)

	return sas, nil
}

func (s *session) AcceptTrackStream(ctx context.Context) (SendTrackStream, error) {
	slog.Debug("accepting track stream")

	ss, err := s.internalSession.AcceptSubscribeStream(ctx)
	if err != nil {
		return nil, err
	}

	if ss == nil {
		return nil, ErrInternalError
	}

	sts := newSendTrackStream(s.internalSession, ss)

	var info Info

	path := sts.TrackPath()
	info, err = s.handler.GetInfo(path)
	if err != nil {
		slog.Error("failed to get track info",
			"track_path", path,
			"error", err,
		)
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

func (s *session) resolveInfoRequest(ctx context.Context) {
	for {
		irs, err := s.internalSession.AcceptInfoStream(ctx)
		if err != nil {
			slog.Error("failed to accept info stream",
				"error", err,
			)

			// return err
		}

		var info Info

		path := TrackPath(irs.InfoRequestMessage.TrackPath)
		info, err = s.handler.GetInfo(path)
		if err != nil {
			slog.Error("failed to get track info",
				"track_path", path,
				"error", err,
			)
			irs.CloseWithError(err)
			// return err
		}

		im := message.InfoMessage{
			TrackPriority:       message.TrackPriority(info.TrackPriority),
			LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
			GroupOrder:          message.GroupOrder(info.GroupOrder),
		}

		_, err = im.Encode(irs.Stream)
		if err != nil {
			slog.Error("failed to send track info",
				"info", info,
				"error", err,
			)
			irs.CloseWithError(err)
			// return err
		}

		// return nil
	}

}

func (s *session) nextSubscribeID() message.SubscribeID {
	// Increment and return the previous value atomically
	id := atomic.AddUint64(&s.subscribeIDCounter, 1) - 1
	return message.SubscribeID(id)
}

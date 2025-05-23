package moqt

// var _ SendTrackStream = (*sendTrackStream)(nil)

// type SendTrackStream interface {
// 	TrackWriter
// 	TrackPath() TrackPath
// 	SubscribeID() SubscribeID
// 	SubuscribeConfig() *SubscribeConfig
// 	Updated() <-chan struct{}
// }

// func newSendTrackStream(session *Session, receiveSubscribeStream *receiveSubscribeStream) *sendTrackStream {
// 	return &sendTrackStream{
// 		session:         session,
// 		subscribeStream: receiveSubscribeStream,
// 	}
// }

// type sendTrackStream struct {
// 	session         *Session
// 	subscribeStream *receiveSubscribeStream
// 	// latestGroupSequence GroupSequence
// 	mu sync.RWMutex
// }

// func (s *sendTrackStream) SubscribeID() SubscribeID {
// 	return s.subscribeStream.SubscribeID()
// }

// func (s *sendTrackStream) SubuscribeConfig() *SubscribeConfig {
// 	return s.subscribeStream.SubuscribeConfig()
// }

// func (s *sendTrackStream) Updated() <-chan struct{} {
// 	return s.subscribeStream.Updated()
// }

// func (s *sendTrackStream) TrackPath() TrackPath {
// 	return s.subscribeStream.TrackPath()
// }

// func (s *sendTrackStream) OpenGroup(sequence GroupSequence) (GroupWriter, error) {
// 	stream, err := s.session.openGroupStream(s.subscribeStream.id, sequence)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// // Update latest group sequence
// 	// if sequence > s.latestGroupSequence {
// 	// 	s.latestGroupSequence = sequence
// 	// }

// 	return stream, nil
// }

// func (s *sendTrackStream) Close() error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	return s.subscribeStream.Close()
// }

// func (s *sendTrackStream) CloseWithError(err error) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	if err == nil {
// 		err = ErrInternalError
// 	}

// 	return s.subscribeStream.CloseWithError(err)
// }

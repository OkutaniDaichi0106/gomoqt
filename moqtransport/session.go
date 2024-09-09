package moqtransport

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type Session struct {
	wtSession      *webtransport.Session
	controlStream  webtransport.Stream
	controlReader  quicvarint.Reader
	controlChannel chan []byte

	selectedVersion Version
	role            Role
}

func (s *Session) ReceiveClientSetup(supportedVersions []Version) (Parameters, error) {
	// Create bidirectional stream to send control messages
	stream, err := s.wtSession.AcceptStream(context.TODO())
	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(stream)

	// Receive SETUP_CLIENT message
	id, err := deserializeHeader(reader)
	if err != nil {
		return nil, err
	}
	if id != CLIENT_SETUP {
		return nil, ErrProtocolViolation
	}
	var cs ClientSetupMessage
	err = cs.deserializeBody(reader)
	if err != nil {
		return nil, err
	}

	// Check if the ROLE parameter is valid
	s.role, err = cs.Parameters.Role()
	if err != nil {
		return nil, err
	}
	// Delete the Role Parameter after getting it
	delete(cs.Parameters, ROLE)

	// Select a version
	s.selectedVersion, err = selectVersion(cs.Versions, supportedVersions)
	if err != nil {
		return nil, err
	}

	s.controlReader = reader

	return cs.Parameters, nil
}

func (s *Session) SendServerSetup(params Parameters) error {

	// Initialise SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: s.selectedVersion,
		Parameters:      params,
	}

	// Send SETUP_SERVER message
	_, err := s.controlStream.Write(ssm.serialize())
	if err != nil {
		return err
	}

	switch s.role {
	case PUB:

	case SUB:
	}

	return nil
}

func (s *Session) OnPublisher(op func(*SessionWithPublisher)) {

}
func (s *Session) OnSubscriber(op func(*SessionWithSubscriber)) {

}

type SessionWithPublisher struct {
	wtSession      *webtransport.Session
	controlStream  webtransport.Stream
	controlReader  quicvarint.Reader
	controlChannel chan []byte

	latestAnnounceMessage AnnounceMessage

	trackAlias TrackAlias

	destinations
}

func newSessionWithPublisher(sess Session) SessionWithPublisher {
	return SessionWithPublisher{
		wtSession:     sess.wtSession,
		controlStream: sess.controlStream,
		controlReader: sess.controlReader,
	}
}

func (s *SessionWithPublisher) ReceiveAnnounce() (Parameters, error) {
	var err error
	// Receive an ANNOUNCE message
	id, err := deserializeHeader(s.controlReader)
	if err != nil {
		return nil, err
	}
	if id != ANNOUNCE {
		return nil, ErrUnexpectedMessage
	}

	//TODO: handle error
	am := AnnounceMessage{}
	err = am.deserializeBody(s.controlReader)
	if err != nil {
		return nil, err
	} // TODO: handle the parameter

	// Register the Track Namespace
	s.latestAnnounceMessage.TrackNamespace = am.TrackNamespace

	// Return Optional Parameter
	return am.Parameters, nil
}

func (s *SessionWithPublisher) SendAnnounceOk() error {
	// Check if receiving announcement is succeeded and the Track Namespace is exist
	if s.latestAnnounceMessage.TrackNamespace == "" {
		return errors.New("no track namespace")
	}

	// Send ANNOUNCE_OK message
	ao := AnnounceOkMessage{
		TrackNamespace: s.latestAnnounceMessage.TrackNamespace,
	}
	_, err := s.controlStream.Write(ao.serialize()) // Handle the error when wrinting message

	publishers.add(s)

	return err
}

func (s *SessionWithPublisher) SendAnnounceError() error {
	// Check if receiving announcement is succeeded and the Track Namespace is exist
	if s.latestAnnounceMessage.TrackNamespace == "" {
		return errors.New("no track namespace")
	}

	// Send ANNOUNCE_ERROR message
	ae := AnnounceError{
		TrackNamespace: s.latestAnnounceMessage.TrackNamespace,
		Code:           ANNOUNCE_INTERNAL_ERROR,
		Reason:         ANNOUNCE_ERROR_REASON[ANNOUNCE_INTERNAL_ERROR],
	}
	_, err := s.controlStream.Write(ae.serialize()) // Handle the error when wrinting message

	return err
}

func (s *SessionWithPublisher) ReceiveObjects(ctx context.Context) error {
	errCh := make(chan error, 1)

	for {
		// Accept a new unidirectional stream
		stream, err := s.wtSession.AcceptUniStream(ctx)
		if err != nil {
			errCh <- err
		}
		go func(stream webtransport.ReceiveStream) {
			buf := make([]byte, 1<<8)
			data := make([]byte, 0, 1<<8)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					if err == io.EOF {
						// Append final data
						data = append(data, buf[:n]...)
						break
					}
					errCh <- err
					return
				}
				// Append read data
				data = append(data, buf[:n]...)
			}

			// Distribute the data to all Subscribers
			go s.distribute(data, ctx)

		}(stream)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err = <-errCh:
			return err
		default:
			continue
		}
	}
}

func (s *SessionWithPublisher) distribute(b []byte, ctx context.Context) {
	var wg sync.WaitGroup
	for _, sess := range s.destinations.sessions {

		// Increment the wait group by 1 and count the number of current processes
		wg.Add(1)
		go func(sess *webtransport.Session) {
			defer wg.Done()
			// Set context to terminate the operation upon timeout
			ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second) // TODO: Left the duration to the user's implementation
			defer cancel()

			// Open a unidirectional stream
			stream, err := sess.OpenStreamSync(ctx)
			if err != nil {
				log.Println(err)
				return //TODO: handle the error
			}

			// Close the stream after whole data was sent
			defer stream.Close()

			// Send data on the stream
			_, err = stream.Write(b)
			if err != nil {
				log.Println(err)
				return //TODO: handle the error
			}
		}(sess)
	}

	// Wait untill the data has been sent to all sessions
	wg.Wait()
}

type destinations struct {
	sessions []*webtransport.Session
	mu       sync.Mutex
}

func (dest *destinations) add(sess *webtransport.Session) {
	dest.mu.Lock()
	defer dest.mu.Unlock()
	dest.sessions = append(dest.sessions, sess)
}

type SessionWithSubscriber struct {
	wtSession      *webtransport.Session
	controlStream  webtransport.Stream
	controlReader  quicvarint.Reader
	controlChannel chan []byte

	//subscribedPublisher SessionWithPublisher

	latestSubscribeMessage SubscribeMessage

	origin *SessionWithPublisher
}

func newSessionWithSubscriber(sess Session) SessionWithSubscriber {
	return SessionWithSubscriber{
		wtSession:      sess.wtSession,
		controlStream:  sess.controlStream,
		controlReader:  sess.controlReader,
		controlChannel: sess.controlChannel,
	}
}

func (s *SessionWithSubscriber) Advertise(announcements []AnnounceMessage) error {
	var err error
	// Send all ANNOUNCE messages
	for _, am := range announcements {
		_, err = s.controlStream.Write(am.serialize())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SessionWithSubscriber) ReceiveSubscribe() (Parameters, error) {
	// Receive a SUBSCRIBE message
	id, err := deserializeHeader(s.controlReader)
	if err != nil {
		return nil, err
	}
	if id != SUBSCRIBE {
		return nil, ErrUnexpectedMessage //TODO: handle as protocol violation
	}
	sm := SubscribeMessage{}
	err = sm.deserializeBody(s.controlReader)
	if err != nil {
		return nil, err
	}

	originSession, ok := publishers.index[sm.TrackNamespace]
	if !ok {
		return nil, errors.New("publisher not found")
	}

	s.origin = originSession

	s.latestSubscribeMessage = sm

	return sm.Parameters, nil
}

func (s *SessionWithSubscriber) SendSubscribeResponce() error {
	var err error
	data := <-s.controlChannel
	switch MessageID(data[0]) {
	case SUBSCRIBE_OK:
		_, err = s.controlStream.Write(data)
		if err != nil {
			return err
		}

		//TODO: handle
	case SUBSCRIBE_ERROR:
		_, err = s.controlStream.Write(data)
		if err != nil {
			return err
		}
		// TODO: Handle the error
	default:
		return ErrUnexpectedMessage
	}

	return nil
}

func (s *SessionWithSubscriber) SendSubscribeInternalError() error {
	sm := s.latestSubscribeMessage
	se := SubscribeError{
		subscribeID: sm.subscribeID,
		Code:        SUBSCRIBE_INTERNAL_ERROR,
		Reason:      SUBSCRIBE_ERROR_REASON[SUBSCRIBE_INTERNAL_ERROR],
		TrackAlias:  sm.TrackAlias,
	}
	_, err := s.controlStream.Write(se.serialize()) // TODO: handle the error

	return err
}

func (s *SessionWithSubscriber) DeliverObjects() {
	s.origin.destinations.add(s.wtSession)
	return
}

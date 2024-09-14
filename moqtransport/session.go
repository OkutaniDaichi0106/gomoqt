package moqtransport

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type ClientSession struct {
	wtSession     *webtransport.Session
	controlStream webtransport.Stream
	controlReader quicvarint.Reader

	supportedVersions []Version
	selectedVersion   Version
	role              Role
}

func (s *ClientSession) ReceiveClientSetup() (Parameters, error) {
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
	s.selectedVersion, err = selectVersion(cs.Versions, s.supportedVersions)
	if err != nil {
		return nil, err
	}

	s.controlStream = stream

	s.controlReader = reader

	return cs.Parameters, nil
}

func (s *ClientSession) SendServerSetup(params Parameters) error {
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

	// switch s.role {//TODO
	// case PUB:

	// case SUB:
	// }

	return nil
}

func (cs ClientSession) OnPublisher(op func(sess *PublisherSession)) {
	if cs.role != PUB {
		return
	}
	sess := &PublisherSession{
		wtSession:      cs.wtSession,
		controlStream:  cs.controlStream,
		controlReader:  cs.controlReader,
		controlChannel: make(chan []byte, 1<<3), // TOOD: tune the size

	}

	op(sess)
}

func (cs ClientSession) OnSubscriber(op func(sess *SubscriberSession)) {
	if cs.role != SUB {
		return
	}
	sess := &SubscriberSession{
		wtSession:     cs.wtSession,
		controlStream: cs.controlStream,
		controlReader: cs.controlReader,
	}

	op(sess)

}

type PublisherSession struct {
	wtSession      *webtransport.Session
	controlStream  webtransport.Stream
	controlReader  quicvarint.Reader
	controlChannel chan []byte

	latestAnnounceMessage AnnounceMessage

	maxSubscriberID subscribeID

	trackAlias TrackAlias

	contentExists   bool
	largestGroupID  groupID
	largestObjectID objectID

	destinations
}

func (s *PublisherSession) ReceiveAnnounce() (Parameters, error) {
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
	}

	// Register the ANNOUNCE message
	announcements.add(am)

	// Register the Track Namespace
	s.latestAnnounceMessage.TrackNamespace = am.TrackNamespace

	// Return Optional Parameter
	return am.Parameters, nil
}

func (s *PublisherSession) SendAnnounceOk() error {
	// Check if receiving announcement is succeeded and the Track Namespace is exist
	fullTrackNamespace := strings.Join(s.latestAnnounceMessage.TrackNamespace, "")
	if fullTrackNamespace == "" {
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

func (s *PublisherSession) SendAnnounceError(code uint, reason string) error {
	// Send ANNOUNCE_ERROR message
	ae := AnnounceError{
		TrackNamespace: s.latestAnnounceMessage.TrackNamespace,
		Code:           AnnounceErrorCode(code),
		Reason:         reason,
	}

	_, err := s.controlStream.Write(ae.serialize())

	return err
}

func (s *PublisherSession) ReceiveObjects(ctx context.Context) error {
	errCh := make(chan error, 1)

	for {
		// Accept a new unidirectional stream
		stream, err := s.wtSession.AcceptUniStream(ctx)
		if err != nil {
			errCh <- err
		}

		// Distribute the data to all Subscribers
		go s.distribute(stream, errCh)

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

func (s *PublisherSession) distribute(src webtransport.ReceiveStream, errCh chan<- error) {
	var wg sync.WaitGroup
	dataCh := make(chan []byte, 1<<4)
	buf := make([]byte, 1<<8)

	go func(src webtransport.ReceiveStream) {
		for {
			n, err := src.Read(buf)
			if err != nil {
				if err == io.EOF {
					// Send final data chunk
					dataCh <- buf[:n]
					return
				}
				log.Println(err)
				errCh <- err
				return
			}
			// Send data chunk
			dataCh <- buf[:n]
		}
	}(src)

	for _, sess := range s.destinations.sessions {
		// Increment the wait group by 1 and count the number of current processes
		wg.Add(1)

		go func(sess *SubscriberSession) {
			defer wg.Done()

			// Open a unidirectional stream
			dst, err := sess.wtSession.OpenUniStream()
			if err != nil {
				log.Println(err)
				errCh <- err
				return //TODO: handle the error
			}

			// Close the stream after whole data was sent
			defer dst.Close()

			for data := range dataCh {
				if len(data) == 0 {
					continue
				}
				_, err := dst.Write(data)
				if err != nil {
					log.Println(err)
					errCh <- err
					return
				}
			}
		}(sess)
	}

	// Wait untill the data has been sent to all sessions
	wg.Wait()
}

type destinations struct {
	sessions []*SubscriberSession
	mu       sync.Mutex
}

func (dest *destinations) add(sess *SubscriberSession) {
	dest.mu.Lock()
	defer dest.mu.Unlock()
	dest.sessions = append(dest.sessions, sess)
}

func (dest *destinations) delete(sess *SubscriberSession) {
	dest.mu.Lock()
	defer dest.mu.Unlock()
	dest.sessions = append(dest.sessions, sess)
}

type SubscriberSession struct {
	wtSession     *webtransport.Session
	controlStream webtransport.Stream
	controlReader quicvarint.Reader
	//controlChannel chan []byte

	//subscribedPublisher SessionWithPublisher

	latestSubscribeMessage SubscribeMessage

	maxSubscriberID subscribeID

	origin *PublisherSession
}

func (s *SubscriberSession) ReceiveSubscribe() (Parameters, error) {
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

func (s *SubscriberSession) SendSubscribeOk(expires time.Duration) error {
	so := SubscribeOkMessage{
		subscribeID:     s.latestSubscribeMessage.subscribeID,
		Expires:         expires,
		GroupOrder:      s.latestSubscribeMessage.GroupOrder,
		ContentExists:   s.origin.contentExists,
		LargestGroupID:  s.origin.largestGroupID,
		LargestObjectID: s.origin.largestObjectID,
		Parameters:      s.latestSubscribeMessage.Parameters,
	}

	_, err := s.controlStream.Write(so.serialize())
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriberSession) SendSubscribeError(code uint, reason string) error {
	se := SubscribeError{
		subscribeID: s.latestSubscribeMessage.subscribeID,
		Code:        SubscribeErrorCode(code),
		Reason:      reason,
		TrackAlias:  s.origin.trackAlias,
	}

	_, err := s.controlStream.Write(se.serialize())
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriberSession) DeliverObjects() {
	// Add the subscriber session to the publisher's destinations
	s.origin.destinations.add(s)
}

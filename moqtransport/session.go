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

type Session interface {
	getWebtransportSession() *webtransport.Session
	getControlStream() webtransport.Stream
	getControlReader() quicvarint.Reader
	HandleRole()
}

type clientSession struct {
	wtSession     *webtransport.Session
	controlStream webtransport.Stream
	controlReader quicvarint.Reader

	roleHandler func()

	setupParameters Parameters

	selectedVersion Version

	// Parameters
	role           Role
	maxSubscribeID subscribeID
}

func (s clientSession) getWebtransportSession() *webtransport.Session {
	return s.wtSession
}

func (s clientSession) getControlStream() webtransport.Stream {
	return s.controlStream
}

func (s clientSession) getControlReader() quicvarint.Reader {
	return s.controlReader
}

func (s *clientSession) receiveClientSetup() ([]Version, error) {
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

	if s.role == PUB {
		// Check if the MAX_SUBSCRIBE_ID parameter is valid
		s.maxSubscribeID, err = cs.Parameters.MaxSubscribeID()
		if err != nil {
			return nil, err
		}
		// Delete the Parameter after getting it
		delete(cs.Parameters, MAX_SUBSCRIBE_ID)
	}

	s.controlStream = stream

	s.controlReader = reader

	return cs.Versions, nil
}

func (s *clientSession) sendServerSetup() error {
	// Initialise SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: s.selectedVersion,
		Parameters:      s.setupParameters,
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

func (s clientSession) HandleRole() {
	s.roleHandler()
}

type PublisherSession struct {
	Session
	controlChannel chan []byte

	latestAnnounceMessage AnnounceMessage

	maxSubscriberID subscribeID

	trackNamespace TrackNamespace

	trackAlias TrackAlias

	contentExists   bool
	largestGroupID  groupID
	largestObjectID objectID
}

func (s *PublisherSession) ReceiveAnnounce() (Parameters, error) {
	var err error
	// Receive an ANNOUNCE message
	id, err := deserializeHeader(s.getControlReader())
	if err != nil {
		return nil, err
	}
	if id != ANNOUNCE {
		return nil, ErrUnexpectedMessage
	}

	//TODO: handle error
	am := AnnounceMessage{}
	err = am.deserializeBody(s.getControlReader())
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
	_, err := s.getControlStream().Write(ao.serialize()) // Handle the error when wrinting message

	publishers.add(s)

	return err
}

func (s *PublisherSession) SendAnnounceError(code uint, reason string) error {
	// Send ANNOUNCE_ERROR message
	ae := AnnounceErrorMessage{
		TrackNamespace: s.latestAnnounceMessage.TrackNamespace,
		Code:           AnnounceErrorCode(code),
		Reason:         reason,
	}

	_, err := s.getControlStream().Write(ae.serialize())

	return err
}

func (s *PublisherSession) ReceiveObjects(ctx context.Context) error {
	errCh := make(chan error, 1)

	for {
		// Accept a new unidirectional stream
		stream, err := s.getWebtransportSession().AcceptUniStream(ctx)
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

	fullTrackNamespace := s.trackNamespace.getFullName()

	dests, ok := subscribers[fullTrackNamespace]
	if !ok {
		//TODO: handle this as internal error
		panic("destinations not found")
	}

	for _, sess := range dests.sessions {
		// Increment the wait group by 1 and count the number of current processes
		wg.Add(1)

		go func(sess *SubscriberSession) {
			defer wg.Done()

			// Open a unidirectional stream
			dst, err := sess.getWebtransportSession().OpenUniStream()
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

type SubscriberSession struct {
	Session
	controlChannel chan []byte

	latestSubscribeMessage SubscribeMessage

	subscriptions map[TrackAlias]struct {
		maxSubscribeID   subscribeID
		trackSubscribeID subscribeID
	}

	origin *PublisherSession
}

func (s *SubscriberSession) ReceiveSubscribe() (Parameters, error) {
	// Receive a SUBSCRIBE message
	id, err := deserializeHeader(s.getControlReader())
	if err != nil {
		return nil, err
	}
	if id != SUBSCRIBE {
		return nil, ErrUnexpectedMessage //TODO: handle as protocol violation
	}

	sm := SubscribeMessage{}
	err = sm.deserializeBody(s.getControlReader())
	if err != nil {
		return nil, err
	}

	pubSess, ok := publishers.index[sm.TrackNamespace]
	if !ok {
		return nil, errors.New("publisher not found")
	}

	// Initialize the map of subscriptions if not exists
	if s.subscriptions == nil {
		s.subscriptions = make(map[TrackAlias]struct {
			maxSubscribeID   subscribeID
			trackSubscribeID subscribeID
		})
	}

	// Check if the Track is already subscirbed
	trackSubscription, ok := s.subscriptions[sm.TrackAlias]

	if !ok {
		// Initialize the structure if the Track is not already subscirbed
		s.subscriptions[sm.TrackAlias] = struct {
			maxSubscribeID   subscribeID
			trackSubscribeID subscribeID
		}{
			maxSubscribeID:   pubSess.maxSubscriberID,
			trackSubscribeID: 0,
		}
	} else if ok {
		// Increment the Subscribe ID in a Track by 1 if the Track is already subscribed
		trackSubscription.trackSubscribeID++

		// Check if the Subscribe ID is not over the Max Subscribe ID
		if trackSubscription.maxSubscribeID == trackSubscription.trackSubscribeID {
			return nil, ErrTooManySubscribes
		}
	}

	// Register the Publisher Session
	s.origin = pubSess

	// Register the SUBSCRIBE message
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

	_, err := s.getControlStream().Write(so.serialize())
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

	_, err := s.getControlStream().Write(se.serialize())
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriberSession) DeliverObjects() {
	// Add the subscriber session to the publisher's destinations
	fullTrackNamespace := s.origin.trackNamespace.getFullName()

	subscribers[fullTrackNamespace].add(s)
}

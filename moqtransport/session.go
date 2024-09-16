package moqtransport

import (
	"context"
	"errors"
	"go-moq/moqtransport/moqtmessage"
	"go-moq/moqtransport/moqtversion"
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
	getMaxSubscribeID() moqtmessage.SubscribeID
	HandleRole()
}

type clientSession struct {
	wtSession     *webtransport.Session
	controlStream webtransport.Stream
	controlReader quicvarint.Reader

	roleHandler func()

	setupParameters moqtmessage.Parameters

	selectedVersion moqtversion.Version

	// Parameters
	role           moqtmessage.Role
	maxSubscribeID moqtmessage.SubscribeID
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

func (s clientSession) getMaxSubscribeID() moqtmessage.SubscribeID {
	return s.maxSubscribeID
}

func (s *clientSession) receiveClientSetup() ([]moqtversion.Version, error) {
	// Create bidirectional stream to send control messages
	stream, err := s.wtSession.AcceptStream(context.TODO())
	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(stream)

	// Receive SETUP_CLIENT message
	id, err := moqtmessage.DeserializeMessageID(reader)
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.CLIENT_SETUP {
		return nil, ErrProtocolViolation
	}
	var cs moqtmessage.ClientSetupMessage
	err = cs.DeserializeBody(reader)
	if err != nil {
		return nil, err
	}

	// Check if the ROLE parameter is valid
	s.role, err = cs.Parameters.Role()
	if err != nil {
		return nil, err
	}
	// Delete the Role Parameter after getting it
	delete(cs.Parameters, moqtmessage.ROLE)

	if s.role == moqtmessage.PUB {
		// Check if the MAX_SUBSCRIBE_ID parameter is valid
		s.maxSubscribeID, err = cs.Parameters.MaxSubscribeID()
		if err != nil {
			return nil, err
		}
		// Delete the Parameter after getting it
		delete(cs.Parameters, moqtmessage.MAX_SUBSCRIBE_ID)
	}

	s.controlStream = stream

	s.controlReader = reader

	return cs.Versions, nil
}

func (s *clientSession) sendServerSetup() error {
	// Initialise SETUP_SERVER message
	ssm := moqtmessage.ServerSetupMessage{
		SelectedVersion: s.selectedVersion,
		Parameters:      s.setupParameters,
	}

	// Send SETUP_SERVER message
	_, err := s.controlStream.Write(ssm.Serialize())
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

	latestAnnounceMessage moqtmessage.AnnounceMessage

	maxSubscriberID moqtmessage.SubscribeID

	trackNamespace moqtmessage.TrackNamespace

	trackAlias moqtmessage.TrackAlias

	contentExists   bool
	largestGroupID  moqtmessage.GroupID
	largestObjectID moqtmessage.ObjectID
}

func (s *PublisherSession) ReceiveAnnounce() (moqtmessage.Parameters, error) {
	var err error
	// Receive an ANNOUNCE message
	id, err := moqtmessage.DeserializeMessageID(s.getControlReader())
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.ANNOUNCE {
		return nil, ErrUnexpectedMessage
	}

	//TODO: handle error
	am := moqtmessage.AnnounceMessage{}
	err = am.DeserializeBody(s.getControlReader())
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
	ao := moqtmessage.AnnounceOkMessage{
		TrackNamespace: s.latestAnnounceMessage.TrackNamespace,
	}
	_, err := s.getControlStream().Write(ao.Serialize()) // Handle the error when wrinting message

	publishers.add(s)

	return err
}

func (s *PublisherSession) SendAnnounceError(aerr AnnounceError) error {
	// Send ANNOUNCE_ERROR message
	ae := moqtmessage.AnnounceErrorMessage{
		TrackNamespace: s.latestAnnounceMessage.TrackNamespace,
		Code:           aerr.Code(),
		Reason:         aerr.Error(),
	}

	_, err := s.getControlStream().Write(ae.Serialize())

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

	fullTrackNamespace := s.trackNamespace.GetFullName()

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

	latestSubscribeMessage moqtmessage.SubscribeMessage

	origin *PublisherSession
}

func (s *SubscriberSession) ReceiveSubscribe() (moqtmessage.Parameters, error) {
	// Receive a SUBSCRIBE message
	id, err := moqtmessage.DeserializeMessageID(s.getControlReader())
	if err != nil {
		return nil, err
	}
	if id != moqtmessage.SUBSCRIBE {
		return nil, ErrUnexpectedMessage //TODO: handle as protocol violation
	}

	sm := moqtmessage.SubscribeMessage{}
	err = sm.DeserializeBody(s.getControlReader())
	if err != nil {
		return nil, err
	}

	pubSess, ok := publishers.index[sm.TrackNamespace]
	if !ok {
		return nil, errors.New("publisher not found")
	}

	// Register the Publisher Session
	s.origin = pubSess

	// Register the SUBSCRIBE message
	s.latestSubscribeMessage = sm

	return sm.Parameters, nil
}

func (s *SubscriberSession) SendSubscribeOk(expires time.Duration) error {
	so := moqtmessage.SubscribeOkMessage{
		SubscribeID:     s.latestSubscribeMessage.SubscribeID,
		Expires:         expires,
		GroupOrder:      s.latestSubscribeMessage.GroupOrder,
		ContentExists:   s.origin.contentExists,
		LargestGroupID:  s.origin.largestGroupID,
		LargestObjectID: s.origin.largestObjectID,
		Parameters:      s.latestSubscribeMessage.Parameters,
	}

	_, err := s.getControlStream().Write(so.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriberSession) SendSubscribeError(serr SubscribeError) error {
	se := moqtmessage.SubscribeError{
		SubscribeID: s.latestSubscribeMessage.SubscribeID,
		Code:        serr.Code(),
		Reason:      serr.Error(),
		TrackAlias:  s.origin.trackAlias,
	}

	_, err := s.getControlStream().Write(se.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriberSession) DeliverObjects() {
	// Add the subscriber session to the publisher's destinations
	fullTrackNamespace := s.origin.trackNamespace.GetFullName()

	subscribers[fullTrackNamespace].add(s)
}

package gomoq

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

type Agent struct {
	//Server *Server
	/*
	 * WebTransport Session
	 */
	session *webtransport.Session

	/*
	 * Bidirectional stream to send control message on
	 */
	controlStream webtransport.Stream

	/*
	 *
	 */
	controlReader quicvarint.Reader

	/*
	 * Role
	 */
	role Role

	/*
	 * MOQT version
	 */
	version Version

	origin *Agent

	/*
	 * the key of the map is a Track Name
	 */
	destinations chan *webtransport.Session

	/*
	 *
	 */
	contentExist bool //TODO: use?

	/*
	 *
	 */
	controlCh chan Messager
}

/*
 * Initialize the Agent
 */
func (a *Agent) init() error {
	// Initialize the channel to send and receive control messages
	a.controlCh = make(chan Messager, 1<<5) //TODO: Tune the size
	go a.listenControlChannel()

	return nil
}

func (a *Agent) listenControlChannel() {
	for cmsg := range a.controlCh {
		switch cmsg.(type) {
		case *SubscribeMessage:
		}
	}
}

func Activate(a *Agent) error {
	err := a.setup()
	if err != nil {
		return err
	}

	switch a.role {
	case PUB:
		server.onPublisher(a) // TODO: goroutine?
	case SUB:
		server.onSubscriber(a)
	case PUB_SUB:
		server.onPubSub(a)
	default:
		return ErrInvalidRole
	}
	return nil
}

/*
 * Exchange SETUP messages
 */
func (a *Agent) setup() error {
	//TODO: Check if the role and the versions is setted
	var err error

	// Init
	err = a.init()
	if err != nil {
		return err
	}

	// Create bidirectional stream to send control messages
	stream, err := a.session.AcceptStream(context.Background())
	if err != nil {
		return err
	}

	// Register the stream as control stream
	a.controlStream = stream

	// Create quic varint reader
	a.controlReader = quicvarint.NewReader(stream)

	// Receive SETUP_CLIENT message
	id, err := deserializeHeader(a.controlReader)
	if err != nil {
		return err
	}
	if id != CLIENT_SETUP {
		return ErrProtocolViolation
	}
	var cs ClientSetupMessage
	err = cs.deserializeBody(a.controlReader)
	if err != nil {
		return err
	}

	// Check if the ROLE parameter is valid
	// Register the Client's role to the Agent
	ok, v := cs.Parameters.Contain(ROLE)
	if !ok {
		return errors.New("no role is specified")
	}
	switch v.(Role) {
	case PUB, SUB, PUB_SUB:
		a.role = v.(Role) // TODO: need?
	default:
		return ErrInvalidRole
	}

	// Select a version
	a.version, err = selectVersion(cs.Versions, server.SupportedVersions)
	if err != nil {
		return err
	}

	// Initialise SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: a.version,
		Parameters:      server.setupParameters,
	}

	// Send SETUP_SERVER message
	_, err = a.controlStream.Write(ssm.serialize())
	if err != nil {
		return err
	}

	return nil
}

/*
 * Select a newest moqt version from a pair of version sets
 */
func selectVersion(vs1, vs2 []Version) (Version, error) {
	versions := []Version{}
	for _, v1 := range vs1 {
		for _, v2 := range vs2 {
			if v1 == v2 {
				versions = append(versions, v1)
			}
		}
	}
	if len(versions) < 1 {
		// Throw an error if there are no common versions between the sets
		return INVALID_VERSION, errors.New("no valid versions")
	}
	var latestVersion Version
	// Select
	for _, version := range versions {
		if latestVersion < version {
			latestVersion = version
		}
	}
	return latestVersion, nil
}

// Only Subscriber's Agents
/*
 * Advertise announcements
 */
func Advertise(agent *Agent, announcements []AnnounceMessage) error {
	if agent.role != SUB && agent.role != PUB_SUB {
		return ErrInvalidRole
	}
	return agent.advertise(announcements)
}

func (a *Agent) advertise(announcements []AnnounceMessage) error {
	var err error
	// Send all ANNOUNCE messages
	for _, am := range announcements {
		_, err = a.controlStream.Write(am.serialize())
		if err != nil {
			return err
		}
	}

	return nil
}

func DeliverObjects(agent *Agent) {
	// Register the session with the Subscriber to the Publisher's Agent
	agent.origin.destinations <- agent.session
	log.Println(agent.session, " is in ", agent.origin.destinations)
}

/*
 * Exchange SUBSCRIBE messages
 */
func AcceptSubscription(agent *Agent) error {
	if agent.role != SUB && agent.role != PUB_SUB {
		return ErrInvalidRole
	}
	return agent.acceptSubscription()
}

func (a *Agent) acceptSubscription() error {
	// Receive a SUBSCRIBE message
	id, err := deserializeHeader(a.controlReader)
	if err != nil {
		return err
	}
	if id != SUBSCRIBE {
		return ErrUnexpectedMessage //TODO: handle as protocol violation
	}
	s := SubscribeMessage{}
	err = s.deserializeBody(a.controlReader)
	if err != nil {
		return err
	}

	// Find the Publisher's Agent from the Track Namespace
	pAgent, err := server.getPublisherAgent(s.TrackNamespace)
	if err != nil {
		se := SubscribeError{
			SubscribeID: s.SubscribeID,
			Code:        SUBSCRIBE_INTERNAL_ERROR,
			Reason:      SUBSCRIBE_ERROR_REASON[SUBSCRIBE_INTERNAL_ERROR],
			TrackAlias:  s.TrackAlias,
		}
		_, err2 := a.controlStream.Write(se.serialize()) // TODO: handle the error
		log.Println(err2)

		return err
	}

	pAgent.controlCh <- &s

	// Receive SUBSCRIBE_OK message or SUBSCRIBE_ERROR message from Publisher's Agent
	// and send it to Subscriber
	cmsg := <-a.controlCh
	switch cmsg.(type) {
	case *SubscribeOkMessage:
		_, err = a.controlStream.Write(cmsg.serialize())
		if err != nil {
			return err
		}
		// Add the agent to the Index
		server.subscribers.add(s.TrackNamespace, a)
		// TODO: when delete agents from the index
	case *SubscribeError:
		_, err = a.controlStream.Write(cmsg.serialize())
		if err != nil {
			return err
		}
	default:
		return ErrUnexpectedMessage // TODO: protocol violation
	}

	a.origin = pAgent

	return nil
}

// Only Publisher's Agents
/*
 * Handle announcement exchange
 * - Receive an ANNOUNCE message from the publisher
 * - Send ANNOUNCE_OK or ANNOUNCE_ERROR message to the publisher
 */
func AcceptAnnounce(agent *Agent) error {
	if agent.role != PUB && agent.role != PUB_SUB {
		return ErrInvalidRole
	}
	return agent.acceptAnnounce()
}

func (a *Agent) acceptAnnounce() error {
	var err error
	var ae AnnounceError
	// Receive an ANNOUNCE message
	id, err := deserializeHeader(a.controlReader)
	if err != nil {
		return err
	}
	if id != ANNOUNCE {
		return ErrUnexpectedMessage
	}

	//TODO: handle error
	am := AnnounceMessage{}
	err = am.deserializeBody(a.controlReader)
	if err != nil {
		ae = AnnounceError{
			TrackNamespace: am.TrackNamespace,
			Code:           AnnounceErrorCode(ANNOUNCE_INTERNAL_ERROR),
			Reason:         ANNOUNCE_ERROR_REASON[ANNOUNCE_INTERNAL_ERROR],
		}
		_, err2 := a.controlStream.Write(ae.serialize()) // Handle the error when wrinting message
		log.Println(err2)

		return err
	} // TODO: handle the parameter

	// Register the ANNOUNCE message
	server.announcements.add(am)

	//TODO
	_, ok := server.publishers.index[am.TrackNamespace]
	if ok {
		ae = AnnounceError{
			TrackNamespace: am.TrackNamespace,
			Code:           DUPLICATE_TRACK_NAMESPACE,
			Reason:         ANNOUNCE_ERROR_REASON[DUPLICATE_TRACK_NAMESPACE],
		}
		_, err2 := a.controlStream.Write(ae.serialize()) // Handle the error when wrinting message
		log.Println(err2)

		return ErrDuplicatedNamespace
	}
	// Register the Publishers' Agent
	server.publishers.add(am.TrackNamespace, a)

	// Send ANNOUNCE_OK message or ANNOUNCE_ERROR message as responce
	aom := AnnounceOkMessage{
		TrackNamespace: am.TrackNamespace,
	}
	_, err = a.controlStream.Write(aom.serialize())
	if err != nil {
		ae = AnnounceError{
			TrackNamespace: am.TrackNamespace,
			Code:           AnnounceErrorCode(ANNOUNCE_INTERNAL_ERROR),
			Reason:         ANNOUNCE_ERROR_REASON[ANNOUNCE_INTERNAL_ERROR],
		}
		_, err2 := a.controlStream.Write(ae.serialize()) // Handle the error when wrinting message
		log.Println(err2)

		return err
	}

	//

	return nil
}

func AcceptObjects(agent *Agent, ctx context.Context) <-chan error {
	errCh := make(chan error, 1) // TODO: Consider the buffer size

	if agent.role != PUB && agent.role != PUB_SUB {
		errCh <- ErrInvalidRole
		return errCh
	}
	return agent.acceptObjects(ctx, errCh)
}

func (a *Agent) acceptObjects(ctx context.Context, errCh chan error) <-chan error {
	buf := make([]byte, 1<<8)
	// Receive data on a stream in a goroutine
	go func() {
		// Close the error channel after finishing accepting objects
		defer close(errCh)

		for {
			// Catch the cancel call
			select {
			case <-ctx.Done():
				// Cancel the current process
				errCh <- ctx.Err()
				return
			default:
				// Create a unidirectional stream
				stream, err := a.session.AcceptUniStream(ctx)
				if err != nil {
					errCh <- err
				}

				// Read whole data on the stream
				data := make([]byte, 0, 1<<8)
				for {
					n, err := stream.Read(buf)
					if err != nil {
						if err == io.EOF {
							break
						}
						errCh <- err
						// Stop to read chunk when some error is detected
						// but continue to receive data if any error was detected
						break
					}
					data = append(data, buf[:n]...)
				}
				// Distribute the data to Subscribers or Relay Servers
				if len(data) > 0 {
					distribute(data, a.destinations)
				}
			}
		}
	}()

	// Return the channels as read only channel
	return errCh
}

func distribute(b []byte, sessions chan *webtransport.Session) {
	var wg sync.WaitGroup

	for sess := range sessions {
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
	log.Println("Finish Distributing!")
}

var ErrProtocolViolation = errors.New("protocol violation")

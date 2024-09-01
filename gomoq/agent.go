package gomoq

import (
	"context"
	"errors"

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
	 * Role
	 */
	role Role

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

	return nil
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

	// Receive SETUP_CLIENT message
	qvReader := quicvarint.NewReader(stream)
	id, err := deserializeHeader(qvReader)
	if id != CLIENT_SETUP {
		return ErrProtocolViolation
	}
	var cs ClientSetupMessage
	err = cs.deserializeBody(qvReader)
	if err != nil {
		return err
	}

	// Check if the ROLE parameter is valid
	// Register the Client's role
	ok, v := cs.Parameters.Contain(ROLE)
	if !ok {
		return errors.New("no role is specified")
	}
	switch Role(v.(uint64)) {
	case PUB:
		a.role = PUB
	case SUB:
		a.role = SUB
	case PUB_SUB:
		a.role = PUB_SUB
	default:
		return ErrInvalidRole
	}

	// Select a version
	version, err := selectVersion(cs.Versions, server.SupportedVersions)
	if err != nil {
		return err
	}

	// Initialise SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: version,
		Parameters:      server.setupParameters,
	}

	// Send SETUP_SERVER message
	_, err = stream.Write(ssm.serialize())
	if err != nil {
		return err
	}

	// If exchange of SETUP messages is complete, set the stream as control stream
	a.controlStream = stream

	return nil
}

func (a *Agent) Channel(to Agent, ctx context.Context) error {
	errCh := make(chan error, 1<<1)

	// Create a channle for sending data
	ch := make(chan []byte)
	// Create a unidirectional stream for receiving data
	rcvStream, err := a.session.AcceptUniStream(context.Background())
	if err != nil {
		return err
	}
	// Create a unidirectional stream for sending data
	sndStreaem, err := to.session.OpenUniStreamSync(context.Background())
	if err != nil {
		return err
	}

	// Goroutine to send data from the stream to the channel
	go func() {
		buf := make([]byte, 1<<10) // TODO: Tune the size
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				n, err := rcvStream.Read(buf)
				if err != nil { // TODO: Handle io.EOF error
					errCh <- err
					return
				}
				ch <- buf[:n]
			}
		}
	}()

	// Goroutine to send data from the channel to the stream
	go func() {
		// Exit the loop when the channel is closed and has no more data
		for data := range ch {
			_, err := sndStreaem.Write(data)
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	select {
	case <-errCh:
		return <-errCh
	default:
		return nil
	}
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
func Advertise(agent *Agent, announcements []*AnnounceMessage) error {
	return agent.advertise(announcements)
}

func (a *Agent) advertise(announcements []*AnnounceMessage) error {
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

/*
 * Exchange SUBSCRIBE messages
 */
func AcceptSubscription(a *Agent) error {
	return a.acceptSubscription()
}

func (a *Agent) acceptSubscription() error {
	// Receive a SUBSCRIBE message
	sReader := quicvarint.NewReader(a.controlStream)
	id, err := deserializeHeader(sReader)
	if err != nil {
		return err
	}
	if id != SUBSCRIBE {
		return ErrUnexpectedMessage //TODO: handle as protocol violation
	}
	s := SubscribeMessage{}
	err = s.deserializeBody(sReader)
	if err != nil {
		return err
	}

	// Find the Publisher's Agent from the Track Namespace
	pAgent, err := server.getPublisherAgent(s.TrackNamespace)
	if err != nil {
		//TODO: configure SUBSCRIBE_ERROR message
		var se SubscribeError
		a.controlCh <- &se
		return err
	}

	// Send the SUBSCRIBE message to the publisher's Agent
	pAgent.controlCh <- &s

	// Receive SUBSCRIBE_OK message or SUBSCRIBE_ERROR message from Publisher
	// and send it to Subscriber
	pReader := quicvarint.NewReader(pAgent.controlStream)
	id, err = deserializeHeader(pReader)
	if err != nil {
		return err
	}
	switch id {
	case SUBSCRIBE_OK:
		so := SubscribeOkMessage{}
		err = so.deserializeBody(pReader)
		if err != nil {
			return err
		}
		_, err = a.controlStream.Write(so.serialize())
		if err != nil {
			return err
		}
		// Add the agent to the Index
		server.subscribersIndex[s.TrackNamespace] = append(a.Server.subscribersIndex[s.TrackNamespace], a)
		// TODO: when delete agents from the index
	case SUBSCRIBE_ERROR:
		se := SubscribeError{}
		err = se.deserializeBody(pReader)
		if err != nil {
			return err
		}
		_, err = a.controlStream.Write(se.serialize())
		if err != nil {
			return err
		}
	default:
		return ErrUnexpectedMessage // TODO: protocol violation
	}

	return nil
}

// Only Publisher's Agents
/*
 * Handle announcement exchange
 * - Receive an ANNOUNCE message from the publisher
 * - Send ANNOUNCE_OK or ANNOUNCE_ERROR message to the publisher
 */
func AcceptAnnounce(a *Agent) error {
	return a.acceptAnnounce()
}

func (a *Agent) acceptAnnounce() error {
	var err error

	// Receive an ANNOUNCE message
	qvReader := quicvarint.NewReader(a.controlStream)
	id, err := deserializeHeader(qvReader)
	if err != nil {
		return err
	}
	if id != ANNOUNCE {
		return ErrUnexpectedMessage
	}

	//TODO: handle error
	am := AnnounceMessage{}
	err = am.deserializeBody(qvReader)
	if err != nil {
		return err
	} // TODO: handle the parameter

	// Register the ANNOUNCE message
	server.announcements.add(am)

	//TODO
	_, ok := server.publisherStorage[am.TrackNamespace]
	if ok {
		return ErrDuplicatedNamespace
	}
	server.publisherStorage[am.TrackNamespace] = a

	// Send ANNOUNCE_OK message or ANNOUNCE_ERROR message as responce
	aom := AnnounceOkMessage{
		TrackNamespace: am.TrackNamespace,
	}
	_, err = a.controlStream.Write(aom.serialize())
	if err != nil {
		return err
	}

	// If failed to authorization ANNOUNCE_ERROR message
	// ae := AnnounceError{}
	// _, err = a.controlStream.Write(ae.serialize())
	// if err != nil {
	// 	return err
	// }

	return nil
}

func AcceptObjects(a *Agent, ctx context.Context) error {
	return a.acceptObjects(ctx)
}

func (a *Agent) acceptObjects(ctx context.Context) error {
	//
	a.Channel()

	return nil
}

var ErrProtocolViolation = errors.New("protocol violation")

package gomoq

import (
	"context"
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

/*
 * Server Agent
 * You should use this in a goroutine such as http.HandlerFunc
 *
 * Server will perform the following operation
 * - Waiting connections by Client
 * - Accepting bidirectional stream to send control messages
 * - Receiving SETUP_CLIENT message from the client
 * - Sending SETUP_SERVER message to the client
 * - Terminating sessions
 */

type Server struct {
	/*
	 *
	 */
	WebtransportServer *webtransport.Server

	/*
	 * Supported versions by the server
	 */
	SupportedVersions []Version

	/*
	 * Supported Track Name by the server
	 */
	TrackNames []string // Original

	/*
	 * Agents runs on the server
	 */
	agents []*Agent

	/*
	 * Index for searching Agent by Namespace
	 */
	namespaceIndex map[string]*Agent

	/*
	 * Index for searching Agent by ID
	 */
	idIndex map[AgentID]*Agent

	/*
	 * Announcements received from publishers
	 */
	Announcements map[AgentID]*AnnounceMessage
}

type AgentID uint64

type Agent struct {
	Server
	/*
	 * Bidirectional stream to send control stream
	 * Set this after connection to the server is completed
	 */
	Session *webtransport.Session

	/*
	 * Agent ID
	 */
	agentID AgentID

	/*
	 * Role of the Client's Agent
	 */
	agentOf Role

	/*
	 * Bidirectional stream to send control stream
	 * Set the first bidirectional stream
	 */
	controlStream webtransport.Stream

	/*
	 * Using selectedVersion which is specifyed by the client and is selected by the server
	 */
	selectedVersion Version

	/*
	 * The operation performed on publishers
	 */
	onPublisher *func() error

	/*
	 * The operation performed on subscribers
	 */
	onSubscriber *func() error

	/*
	 * The operation performed on subscribers
	 */
	onPubSub *func() error

	//acceptedTrackNamespace []string
}

/*
 * Client connect to the server
 * Dial to the server and establish a session
 * Accept bidirectional stream to send control message
 *
 */
func (a *Agent) Setup() error {
	//TODO: Check if the role and the versions is setted
	var err error

	stream, err := a.Session.AcceptStream(context.Background())
	if err != nil {
		return err
	}

	// Receive SETUP_CLIENT message
	qvReader := quicvarint.NewReader(stream)
	var cs ClientSetupMessage
	err = cs.deserialize(qvReader)
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
	case PUB, SUB, PUB_SUB:
		a.agentOf = Role(v.(uint64))
	default:
		return errors.New("invalid role is specified")
	}

	// Select version
	versionIsOK := false
	for _, cv := range cs.Versions {
		for _, sv := range a.SupportedVersions {
			if cv == sv {
				a.selectedVersion = sv
				versionIsOK = true
				break
			}
		}
	}
	if !versionIsOK {
		return errors.New("no version is selected")
	}

	// Initialise SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: a.selectedVersion,
	}

	// Add Track Name defined in the Server to the SETUP_SERVER parameters
	for _, v := range a.TrackNames { // Original
		ssm.AddStringParameter(TRACK_NAME, v)
	}

	// Send SETUP_SERVER message
	_, err = stream.Write(ssm.serialize())
	if err != nil {
		return err
	}

	// If exchang of SETUP messages is complete, set the stream as control stream
	a.controlStream = stream

	// Assign an Agent ID
	a.agentID = AgentID(len(a.agents))
	// Register the Agent to the server
	a.agents = append(a.agents, a)
	// Register the Agent to an index
	a.idIndex[a.agentID] = a

	return nil
}

/*
 * Advertise announcements
 */
func (a *Agent) Advertise() error {
	// Only Publiser's Agent perform this operation
	if a.agentOf == SUB { //TODO: handle this as protocol violation
		return ErrUnsuitableRole
	}
	var err error
	// Send all ANNOUNCE messages
	for _, am := range a.Announcements {
		_, err = a.controlStream.Write(am.serialize())
		if err != nil {
			return err
		}
	}

	return nil
}

func (a Agent) AcceptSubscribe() error {
	// Only Publiser's Agent perform this operation
	if a.agentOf == SUB { //TODO: handle this as protocol violation
		return ErrUnsuitableRole
	}

	// Receive an ANNOUNCE message
	sReader := quicvarint.NewReader(a.controlStream)
	id, err := deserializeHeader(sReader)
	if err != nil {
		return err
	}
	if id != SUBSCRIBE {
		return ErrUnexpectedMessage
	}

	// Command the Agent of the subscribed publisher to send the SUBSCRIBE message
	s := SubscribeMessage{}
	err = s.deserializeBody(sReader) //TODO: Transfer data as a raw byte stream?
	if err != nil {
		return err
	}

	// Find the Publisher's Agent which send the SUBSCRIBE message from specified Track Namespace
	pAgent, ok := a.namespaceIndex[s.TrackNamespace]
	if !ok {
		//TODO: send SUBSCRIBE_ERROR message
		return ErrNoAgent
	}

	// Send the SUBSCRIBE message to the publisher
	_, err = pAgent.controlStream.Write(s.serialize())
	if err != nil {
		return err
	}

	// Receive SUBSCRIBE_OK message or SUBSCRIBE_ERROR message from Publisher
	// and send it to Subscriber
	pReader := quicvarint.NewReader(pAgent.controlStream)
	id, err = deserializeHeader(pReader)
	if err != nil {
		return err
	}
	if id == SUBSCRIBE_OK {
		so := SubscribeOkMessage{}
		err = so.deserializeBody(pReader)
		if err != nil {
			return err
		}
		_, err = a.controlStream.Write(so.serialize())
		if err != nil {
			return err
		}
	} else if id == SUBSCRIBE_ERROR {
		se := SubscribeError{}
		err = se.deserializeBody(pReader)
		if err != nil {
			return err
		}
		_, err = a.controlStream.Write(se.serialize())
		if err != nil {
			return err
		}
	} else {
		return ErrUnexpectedMessage // TODO: protocol violation
	}

	return nil
}

func (a *Agent) AcceptAnnounce() error {
	// Only Publiser's Agent perform this operation
	if a.agentOf == SUB { //TODO: handle this as protocol violation
		return ErrUnsuitableRole
	}

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
	an := AnnounceMessage{}
	err = an.deserializeBody(qvReader)
	if err != nil {
		return err
	} // TODO: handle the parameter

	// Register the ANNOUNCE message
	_, ok := a.namespaceIndex[an.TrackNamespace]
	if ok {
		return ErrDuplicatedNamespace
	}
	a.namespaceIndex[an.TrackNamespace] = a

	// Send ANNOUNCE_OK message or ANNOUNCE_ERROR message as responce
	aom := AnnounceOkMessage{
		TrackNamespace: an.TrackNamespace,
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

var ErrUnsuitableRole = errors.New("the role cannot perform the operation ")
var ErrUnexpectedMessage = errors.New("received message is not a expected message")
var ErrInvalidRole = errors.New("given role is invalid")
var ErrDuplicatedNamespace = errors.New("given namespace is already registered")
var ErrNoAgent = errors.New("no agent")

//TODO: can reduce the number of quicvarint.Reader?

func (a *Agent) PublisherHandle(op func() error) error {
	a.onPublisher = &op
	return nil
}

func (a *Agent) SubscriberHandle(op func() error) error {
	a.onSubscriber = &op
	return nil
}

func (a *Agent) PubSubHandle(op func() error) error {
	a.onPubSub = &op
	return nil
}

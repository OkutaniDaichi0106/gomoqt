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
	 * Announcements received from publishers
	 */
	Announcements map[AgentID]AnnounceMessage
}

type AgentID uint64

type Agent struct {
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
	 * Role of the Client connceted to the Agent
	 */
	connectTo Role

	/*
	 * Bidirectional stream to send control stream
	 * Set the first bidirectional stream
	 */
	controlStream webtransport.Stream

	/*
	 * Using selectedVersion which is specifyed by the client and is selected by the server
	 */
	selectedVersion Version

	//
	acceptedTrackNamespace []string
}

/*
 * Client connect to the server
 * Dial to the server and establish a session
 * Accept bidirectional stream to send control message
 *
 */
func (a *Agent) Setup(server *Server) error {
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
	case pub, sub, pub_sub:
		a.connectTo = Role(v.(uint64))
	default:
		return errors.New("invalid role is specified")
	}

	// Select version
	versionIsOK := false
	for _, cv := range cs.Versions {
		for _, sv := range server.SupportedVersions {
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
	for _, v := range server.TrackNames {
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
	a.agentID = AgentID(len(server.agents))
	// Register the agent to the server
	server.agents = append(server.agents, a)

	return nil
}

/*
 * Advertise announcements
 */
func (a *Agent) Advertise(server Server) error {
	var err error
	// Send all ANNOUNCE messages
	for _, am := range server.Announcements {
		_, err = a.controlStream.Write(am.serialize())
		if err != nil {
			return err
		}
	}

	return nil
}

func (a Agent) AcceptSubscribe() error {
	//
	return nil
}

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
	 * Available versions for the server
	 */
	SupportedVersions []Version
}

type Agent struct {
	Server
	/*
	 * Bidirectional stream to send control stream
	 * Set this after connection to the server is completed
	 */
	Session *webtransport.Session

	/*
	 * Bidirectional stream to send control stream
	 * Set the first bidirectional stream
	 */
	controlStream *webtransport.Stream

	/*
	 * Using selectedVersion which is specifyed by the client and is selected by the server
	 */
	selectedVersion Version
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

	// Send SETUP_SERVER message
	ssm := ServerSetupMessage{
		SelectedVersion: a.selectedVersion,
	}
	_, err = stream.Write(ssm.serialize())
	if err != nil {
		return err
	}

	// If exchang of SETUP messages is complete, set the stream as control stream
	a.controlStream = &stream

	return nil
}

package gomoq

import (
	"context"
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

/*
 * Server
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
	 * Bidirectional stream to send control stream
	 * Set this after connection to the server
	 */
	session *webtransport.Session

	/*
	 * Bidirectional stream to send control stream
	 * Set the first bidirectional stream
	 */
	controlStream webtransport.Stream

	/*
	 * SERVER_SETUP message
	 */
	ServerSetup ServerSetupMessage

	/*
	 * Available versions for the server
	 */
	AvailableVersions []Version

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
func (s *Server) Setup(sess *webtransport.Session) error {
	//TODO: Check if the role and the versions is setted
	var err error

	s.session = sess

	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		return err
	}

	s.controlStream = stream

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
		for _, sv := range s.AvailableVersions {
			if cv == sv {
				s.selectedVersion = sv
				versionIsOK = true
				break
			}

		}
	}
	if !versionIsOK {
		return errors.New("no version is selected")
	}

	// Send SETUP_SERVER message
	stream.Write(s.ServerSetup.serialize())

	// If exchang of SETUP messages is complete, set the stream as control stream
	s.controlStream = stream

	return nil
}

// TODO: should use sync.Once?
func (s *Server) ListenAndServeTLS(certFile string, keyFile string) error {
	return s.WebtransportServer.ListenAndServeTLS(certFile, keyFile)
}

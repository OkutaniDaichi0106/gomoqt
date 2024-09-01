package gomoq

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

/*
 * Client including Publisher and Subscriber
 *
 * Client will perform the following operation
 * - Connencting to the server
 * - Opening bidirectional stream to send control messages
 * - Sending SETUP_CLIENT message to the server
 * - Receiving SETUP_SERVER message from the server
 * - Terminating sessions
 */

type Client struct {
	/*
	 * TLS configuration
	 */
	TLSConfig *tls.Config

	/*
	 * Versions supported by the client
	 */
	Versions []Version

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
	 * Using selectedVersion which is specifyed by the client and is selected by the server
	 */
	selectedVersion Version

	ClientParameterHandler

	/*
	 * CLIENT_SETUP message
	 */
	//clientSetupMessage ClientSetupMessage
}

type ClientParameterHandler interface {
	ClientSetupParameters() Parameters
	AnnounceParameters() Parameters
}

/*
 * Client connect to the server
 * Dial to the server and establish a session
 * Open bidirectional stream to send control message
 *
 */
func (c *Client) connect(url string, role Role) error {
	//TODO: Check if the role and the versions is setted
	var err error
	// Define new Dialer
	var d webtransport.Dialer
	// Set tls configuration
	if c.TLSConfig == nil {
		panic("no TLS configuration")
	}
	d.TLSClientConfig = c.TLSConfig

	// Set header //TODO: How to handle header
	var headers http.Header

	// Dial to the server with Extended CONNECT request
	_, sess, err := d.Dial(context.Background(), url, headers)
	if err != nil {
		return err
	}

	// Register the session to the client
	c.session = sess

	// Open first stream to send control messages
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	// Send SETUP_CLIENT message
	csm := ClientSetupMessage{
		Versions:   c.Versions,
		Parameters: c.ClientSetupParameters(),
	}
	csm.addIntParameter(ROLE, uint64(role))

	_, err = stream.Write(csm.serialize())
	if err != nil {
		return err
	}

	// Receive SETUP_SERVER message
	qvReader := quicvarint.NewReader(stream)
	id, err := deserializeHeader(qvReader)
	if err != nil {
		return err
	}
	if id != SERVER_SETUP {
		return ErrProtocolViolation
	}
	var ss ServerSetupMessage
	err = ss.deserializeBody(qvReader)
	if err != nil {
		return err
	}

	// Check specified version is selected
	versionIsOK := false
	for _, v := range c.Versions {
		if v == ss.SelectedVersion {
			versionIsOK = true
			break
		}
	}
	if !versionIsOK {
		return errors.New("unexcepted version is selected")
	}
	c.selectedVersion = ss.SelectedVersion
	// TODO: Handle ServerSetup Parameters

	// If exchang of SETUP messages is complete, set the stream as control stream
	c.controlStream = stream

	return nil
}

func (c *Client) terminate() error {
	// Send Error message to the server before close the stream

	// Close the controll stream
	c.controlStream.Close() //TODO: must it be closed?
	//c.session.CloseWithError() //TODO:
	return nil
}

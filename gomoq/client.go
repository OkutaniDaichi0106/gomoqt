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

	/*
	 * CLIENT_SETUP message
	 */
	clientSetupMessage ClientSetupMessage
}

/*
 * Client connect to the server
 * Dial to the server and establish a session
 * Open bidirectional stream to send control message
 *
 */
func (c *Client) connect(url string) error {
	//TODO: Check if the role and the versions is setted
	var err error
	// Define new Dialer
	var d webtransport.Dialer
	// Set tls configuration
	d.TLSClientConfig = &tls.Config{}

	// Set header
	var headers http.Header

	// Dial to the server with Extended CONNECT request
	_, sess, err := d.Dial(context.Background(), url, headers)
	if err != nil {
		return err
	}
	c.session = sess

	// Open first stream to send control messages
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	// Send SETUP_CLIENT message
	stream.Write(c.clientSetupMessage.serialize())

	// Receive SETUP_SERVER message
	qvReader := quicvarint.NewReader(stream)
	var ss ServerSetupMessage
	err = ss.deserialize(qvReader)
	if err != nil {
		return err
	}

	// Check specifyed version is selected
	versionIsOK := false
	for _, v := range c.clientSetupMessage.Versions {
		if v == ss.SelectedVersion {
			versionIsOK = true
			break
		}
	}
	if !versionIsOK {
		return errors.New("unexcepted version is selected")
	}
	c.selectedVersion = ss.SelectedVersion

	// If exchang of SETUP messages is complete, set the stream as control stream
	c.controlStream = stream

	return nil
}

/*
func (c *Client) Reconnect() error {
	// Client connect to the server with the same settings as previouse one

	// Set header
	var headers http.Header

	// Dial to the server with Extended CONNECT request
	_, sess, err := c.dialer.Dial(context.Background(), c.url, headers)
	if err != nil {
		return err
	}
	c.session = sess

	// Open first stream to send control messages
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}
	c.controlStream = stream

	return nil
}
*/

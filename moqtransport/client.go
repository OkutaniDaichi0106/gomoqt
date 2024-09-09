package moqtransport

import (
	"context"
	"crypto/tls"
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

	/***/
	controlReader quicvarint.Reader

	/*
	 * Using selectedVersion which is specifyed by the client and is selected by the server
	 */
	selectedVersion Version

	ClientHandler

	/*
	 * CLIENT_SETUP message
	 */
	//clientSetupMessage ClientSetupMessage
}

type ClientHandler interface {
	ClientSetupParameters() Parameters
}

// Check the Publisher inplement Publisher Handler
var _ ClientHandler = Client{}

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

	return nil
}

func (c *Client) setup(role Role) (Parameters, error) {
	var err error

	// Open first stream to send control messages
	c.controlStream, err = c.session.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	// Send SETUP_CLIENT message
	err = c.sendSetupMessage(role)
	if err != nil {
		return nil, err
	}

	// Initialize control reader
	c.controlReader = quicvarint.NewReader(c.controlStream)

	// Receive SETUP_SERVER message
	return c.receiveServerSetup()
}

func (c Client) sendSetupMessage(role Role) error {
	// Initialize SETUP_CLIENT message
	csm := ClientSetupMessage{
		Versions:   c.Versions,
		Parameters: c.ClientSetupParameters(),
	}
	csm.AddParameter(ROLE, uint64(role))

	_, err := c.controlStream.Write(csm.serialize())
	if err != nil {
		return err
	}

	return nil
}
func (c *Client) receiveServerSetup() (Parameters, error) {
	// Receive SETUP_SERVER message

	id, err := deserializeHeader(c.controlReader)
	if err != nil {
		return nil, err
	}
	if id != SERVER_SETUP {
		return nil, ErrUnexpectedMessage
	}
	var ss ServerSetupMessage
	err = ss.deserializeBody(c.controlReader)
	if err != nil {
		return nil, err
	}

	// Check specified version is selected
	err = contain(ss.SelectedVersion, c.Versions)
	if err != nil {
		return nil, err
	}

	// Register the selected version
	c.selectedVersion = ss.SelectedVersion

	return ss.Parameters, nil
}

func (c *Client) terminate() error {
	// Send Error message to the server before close the stream

	// Close the controll stream
	c.controlStream.Close() //TODO: must it be closed?
	//c.session.CloseWithError() //TODO:
	return nil
}
